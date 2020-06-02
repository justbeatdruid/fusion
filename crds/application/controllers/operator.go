package controllers

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"

	nlptv1 "github.com/chinamobile/nlpt/crds/application/api/v1"

	"k8s.io/klog"
)

const path string = "/consumers"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

type Operator struct {
	Host       string
	Port       int
	CAFile     string
	TopicToken string
}

type ConsumerRequestBody struct {
	ConsumerName string   `json:"username"`
	ConsumerID   string   `json:"custom_id"`
	Tags         []string `json:"tags"`
}

type ConsumerCreBody struct {
	Algorithm string `json:"algorithm"`
}

type ConsumerResponseBody struct {
	CustomID  string      `json:"custom_id"`
	CreatedAt int         `json:"created_at"`
	ID        string      `json:"id"`
	Tags      interface{} `json:"tags"`
	Username  string      `json:"username"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields"`
	Code      int         `json:"code"`
}

type ConsumerCreRspBody struct {
	RsaPublicKey interface{} `json:"rsa_public_key"`
	CreatedAt    int         `json:"created_at"`
	Consumer     struct {
		ID string `json:"id"`
	} `json:"consumer"`
	ID        string      `json:"id"`
	Tags      []string    `json:"tags"`
	Key       string      `json:"key"`
	Secret    string      `json:"secret"`
	Algorithm string      `json:"algorithm"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields"`
	Code      int         `json:"code"`
}

type jwtCustomClaims struct {
	jwt.StandardClaims
	// 追加自己需要的信息
	//Uid uuid.UUID     `json:"uid,omitempty"`
}

/*
{"message":"UNIQUE violation detected on '{name=\"app-manager\"}'","name":"unique constraint violation","fields":{"name":"app-manager"},"code":5}
*/
type FailMsg struct {
	Message string      `json:"message"`
	Name    string      `json:"name"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
}

type requestLogger struct {
	prefix string
}

var logger = &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

func NewOperator(host string, port int, cafile string, token string) (*Operator, error) {
	klog.Infof("NewOperator  event:%s %d %s", host, port, cafile)
	return &Operator{
		Host:       host,
		Port:       port,
		CAFile:     cafile,
		TopicToken: token,
	}, nil
}

func (r *Operator) CreateConsumerByKong(db *nlptv1.Application) (err error) {
	klog.Infof("Enter CreateConsumerByKong name:%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.Host, r.Port, path))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)

	requestBody := &ConsumerRequestBody{
		ConsumerName: db.ObjectMeta.Name, //test app的id
		ConsumerID:   db.ObjectMeta.Name,
	}
	responseBody := &ConsumerResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create consumer error: %+v", errs)
	}
	klog.V(5).Infof("creation response body: %s", string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, responseBody.Message)
		return fmt.Errorf("request for create consumer error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("create consumer id : %s", responseBody.ID)
	//update consumer id
	(*db).Spec.ConsumerInfo.ConsumerID = responseBody.ID
	if err != nil {
		return fmt.Errorf("create consumer error %s", responseBody.Message)
	}
	return nil
}

func (r *Operator) CreateConsumerCredentials(db *nlptv1.Application) (err error) {
	id := db.Spec.ConsumerInfo.ConsumerID
	klog.Infof("begin create credentials id %s", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s/%s/jwt", schema, r.Host, r.Port, path, id))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)

	requestBody := &ConsumerCreBody{
		Algorithm: "HS256", //加密算法
	}
	responseBody := &ConsumerCreRspBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create consumer error: %+v", errs)
	}
	klog.V(5).Infof("create consumer credentials rsp: code %d body %s", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		return fmt.Errorf("request for create consumer credentials error: receive wrong status code: %s", string(body))
	}

	(*db).Spec.ConsumerInfo.Key = responseBody.Key
	(*db).Spec.ConsumerInfo.Secret = responseBody.Secret
	if err != nil {
		return fmt.Errorf("create consumer error %s", responseBody.Message)
	}
	return nil
}

func CreateToken(db *nlptv1.Application) (err error) {
	ser := db.Spec.ConsumerInfo.Secret
	key := db.Spec.ConsumerInfo.Key
	//uid := uuid.NewV4()
	claims := &jwtCustomClaims{
		jwt.StandardClaims{
			//1小时超时
			ExpiresAt: int64(time.Now().Add(time.Hour * 1).Unix()),
			Issuer:    key,
			IssuedAt:  1516239022,
		},
		//uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(ser))
	(*db).Spec.ConsumerInfo.Token = tokenString
	return err
}

func (r *Operator) DeleteConsumerByKong(db *nlptv1.Application) (err error) {
	id := db.ObjectMeta.Name
	klog.Infof("delete consumer %s.", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	klog.Infof("delete consumer is %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id)).End()
	request = request.Retry(3, 5*time.Second, retryStatus...)

	if len(errs) > 0 {
		return fmt.Errorf("request for delete consumer error: %+v", errs)
	}

	klog.V(5).Infof("delete consumer response code: %d%s", response.StatusCode, string(body))
	if response.StatusCode != 204 {
		return fmt.Errorf("request for delete consumer error: receive wrong status code: %d", response.StatusCode)
	}

	return nil
}

func (r *Operator) CreateTopicToken(username string) (string, error) {
	var content = []byte(r.TopicToken)
	jwtToken := jwt.New(jwt.SigningMethodHS256)
	header := make(map[string]interface{})

	//默认用HS256算法
	header["alg"] = jwt.SigningMethodHS256.Name
	claims := make(jwt.MapClaims)
	claims["sub"] = username

	jwtToken.Claims = claims
	jwtToken.Header = header
	ts, err := jwtToken.SignedString(content)
	if err != nil {
		return "", err
	}
	return ts, nil
}
