package service

import (
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"net/http"
	"strings"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/auth/cas"
)

type ApiPlugin struct {
	// basic information
	Id          string              `json:"id"`
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Namespace   string              `json:"namespace"`
	User        string              `json:"user"`
	UserName    string              `json:"userName"`
	Description string              `json:"description"`
	ConsumerId  string              `json:"consumerId"`
	Config      interface{}         `json:"config"`
	ReplaceUri  string              `json:"replaceUri"`
	ApiRelation []ApiPluginRelation `json:"apirelation"`

	CreatedAt           time.Time `json:"createdAt"`
	CreatedAtTimestamp  int64     `json:"createdAtTimestamp"`
	ReleasedAt          time.Time `json:"releasedAt"`
	ReleasedAtTimestamp int64     `json:"releasedAtTimestamp"`

	Status string `json:"status"`
}

type ApiPluginRelation struct {
	Id           int    `json:"id"`
	ApiPluginId  string `json:"apiGroupId"`
	TargetId     string `json:"targetId"`
	TargetType   string `json:"targetType"`
	KongPluginId string `json:"kongPluginId"`
	Status       string `json:"status"`
	Detail       string `json:"detail"`
	Enable       bool   `json:"enable"`
}

type ApiBind struct {
	ID string `json:"id"`
}

type BindReq struct {
	Operation string    `json:"operation"`
	Type      string    `json:"type"`
	Apis      []ApiBind `json:"apis"`
}

type FieldValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FieldDetailInfo struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
type InputResTransformerConfig struct {
	Remove struct {
		Json    []string `json:"json,omitempty"`
		Headers []string `json:"headers,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Headers []FieldValue `json:"headers,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Json    []FieldDetailInfo `json:"json,omitempty"`
		Headers []FieldValue      `json:"headers,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Json    []FieldDetailInfo `json:"json,omitempty"`
		Headers []FieldValue      `json:"headers,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
	} `json:"append,omitempty"`
}

type OutResponseTransformer struct {
	Name   string                    `json:"name"`
	Config InputResTransformerConfig `json:"config,omitempty"`
}

type ResponseTransformer struct {
	Name   string               `json:"name"`
	Config ResTransformerConfig `json:"config,omitempty"`
}

type ResTransformerConfig struct {
	Remove struct {
		Json    []string `json:"json,omitempty"`
		Headers []string `json:"headers,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Headers []string `json:"headers,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
		Json       []string `json:"json,omitempty"`
		Json_types []string `json:"json_types,omitempty"`
		Headers    []string `json:"headers,omitempty"`
	} `json:"append,omitempty"`
}
type Consumer struct {
	Id string `json:"id"`
}
type ResTransformerRequestBody struct {
	Consumer *Consumer `json:"consumer,omitempty"`
	Name   string               `json:"name"`
	Config ResTransformerConfig `json:"config"`
}
type ResTransformerResponseBody struct {
	CreatedAt int                  `json:"created_at"`
	Config    ResTransformerConfig `json:"config"`
	ID        string               `json:"id"`
	Service   interface{}          `json:"service"`
	Name      string               `json:"name"`
	Protocols []string             `json:"protocols"`
	Enabled   bool                 `json:"enabled"`
	RunOn     string               `json:"run_on"`
	Consumer  interface{}          `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Fields  interface{} `json:"fields"`
}

type InputReqTransformerConfig struct {
	HttpMethod string `json:"http_method"`
	Remove     struct {
		Body        []string     `json:"json,omitempty"`
		Headers     []string     `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Body        []FieldValue `json:"json,omitempty"`
		Headers     []FieldValue `json:"headers,omitempty"`
		Querystring []FieldValue `json:"querystring,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Body        []FieldValue `json:"json,omitempty"`
		Headers     []FieldValue `json:"headers,omitempty"`
		Querystring []FieldValue `json:"querystring,omitempty"`
		//Urls        string       `json:"urls,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Body        []FieldValue `json:"json,omitempty"`
		Headers     []FieldValue `json:"headers,omitempty"`
		Querystring []FieldValue `json:"querystring,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
		Body        []FieldValue `json:"json,omitempty"`
		Headers     []FieldValue `json:"headers,omitempty"`
		Querystring []FieldValue `json:"querystring,omitempty"`
	} `json:"append,omitempty"`
}
type OutRequestTransformer struct {
	Name   string                    `json:"name"`
	Config InputReqTransformerConfig `json:"config,omitempty"`
}

type RequestTransformer struct {
	Name   string               `json:"name"`
	Config ReqTransformerConfig `json:"config,omitempty"`
}

type ReqTransformerConfig struct {
	HttpMethod string `json:"http_method,omitempty"`
	Remove     struct {
		Body        []string `json:"json,omitempty"`
		Headers     []string `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
	} `json:"remove,omitempty"`
	Rename struct {
		Body        []string `json:"json,omitempty"`
		Headers     []string `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
	} `json:"rename,omitempty"`
	Replace struct {
		Body        []string `json:"json,omitempty"`
		Headers     []string `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
		Uri         string   `json:"uri,omitempty"`
	} `json:"replace,omitempty"`
	Add struct {
		Body        []string `json:"json,omitempty"`
		Headers     []string `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
	} `json:"add,omitempty"`
	Append struct {
		Body        []string `json:"json,omitempty"`
		Headers     []string `json:"headers,omitempty"`
		Querystring []string `json:"querystring,omitempty"`
	} `json:"append,omitempty"`
}

// kong
type ReqTransformerRequestBody struct {
	Consumer *Consumer `json:"consumer,omitempty"`
	Name   string               `json:"name"`
	Config ReqTransformerConfig `json:"config"`
}
type ReqTransformerResponseBody struct {
	CreatedAt int                  `json:"created_at"`
	Config    ResTransformerConfig `json:"config"`
	ID        string               `json:"id"`
	Service   interface{}          `json:"service"`
	Name      string               `json:"name"`
	Protocols []string             `json:"protocols"`
	Enabled   bool                 `json:"enabled"`
	RunOn     string               `json:"run_on"`
	Consumer  interface{}          `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Fields  interface{} `json:"fields"`
}

type Transformer struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config,omitempty"`
}

const ApiType = "api"
const ServiceunitType = "serviceunit"
const RequestTransType string = "request-transformer"
const ResponseTransType string = "response-transformer"
const BindSuccess string = "bindSuccess"
const BindFailed string = "bindFailed"
const BindInit string = "bindInit"
const UnbindSuccess string = "unbindSuccess"
const UnbindFailed string = "unbindFailed"
const UnbindInit string = "unbindInit"

type ApiRes struct {
	Id         string
	Name       string
	BindStatus string
	Enable     bool
	Detail     string
}

func ToInputTransformerInfo(info ResTransformerConfig) InputResTransformerConfig {
	var output InputResTransformerConfig
	output.Remove = info.Remove

	for i := 0; i < len(info.Rename.Headers); i++ {
		arr := strings.Split(info.Rename.Headers[i], ":")
		output.Rename.Headers = append(output.Rename.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Headers); i++ {
		arr := strings.Split(info.Replace.Headers[i], ":")
		output.Replace.Headers = append(output.Replace.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Json); i++ {
		arr := strings.Split(info.Replace.Json[i], ":")
		output.Replace.Json = append(output.Replace.Json,
			FieldDetailInfo{Key: arr[0], Value: arr[1], Type: info.Replace.Json_types[i]})
	}
	for i := 0; i < len(info.Add.Headers); i++ {
		arr := strings.Split(info.Add.Headers[i], ":")
		output.Add.Headers = append(output.Add.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Add.Json); i++ {
		arr := strings.Split(info.Add.Json[i], ":")
		output.Add.Json = append(output.Add.Json,
			FieldDetailInfo{Key: arr[0], Value: arr[1], Type: info.Add.Json_types[i]})
	}
	return output
}

func ToInputReqTransformerInfo(info ReqTransformerConfig) InputReqTransformerConfig {
	var output InputReqTransformerConfig
	if len(info.HttpMethod) != 0 {
		output.HttpMethod = info.HttpMethod
	}
	output.Remove.Headers = info.Remove.Headers
	output.Remove.Body = info.Remove.Body
	output.Remove.Querystring = info.Remove.Querystring

	for i := 0; i < len(info.Rename.Headers); i++ {
		arr := strings.Split(info.Rename.Headers[i], ":")
		output.Rename.Headers = append(output.Rename.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Rename.Body); i++ {
		arr := strings.Split(info.Rename.Body[i], ":")
		output.Rename.Body = append(output.Rename.Body, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Rename.Querystring); i++ {
		arr := strings.Split(info.Rename.Querystring[i], ":")
		output.Rename.Querystring = append(output.Rename.Querystring, FieldValue{Key: arr[0], Value: arr[1]})
	}

	for i := 0; i < len(info.Replace.Headers); i++ {
		arr := strings.Split(info.Replace.Headers[i], ":")
		output.Replace.Headers = append(output.Replace.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Body); i++ {
		arr := strings.Split(info.Replace.Body[i], ":")
		output.Replace.Body = append(output.Replace.Body, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Replace.Querystring); i++ {
		arr := strings.Split(info.Replace.Querystring[i], ":")
		output.Replace.Querystring = append(output.Replace.Querystring, FieldValue{Key: arr[0], Value: arr[1]})
	}
	//output.Replace.Urls = info.Replace.Uri
	for i := 0; i < len(info.Add.Headers); i++ {
		arr := strings.Split(info.Add.Headers[i], ":")
		output.Add.Headers = append(output.Add.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Add.Body); i++ {
		arr := strings.Split(info.Add.Body[i], ":")
		output.Add.Body = append(output.Add.Body, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Add.Querystring); i++ {
		arr := strings.Split(info.Add.Querystring[i], ":")
		output.Add.Querystring = append(output.Add.Querystring, FieldValue{Key: arr[0], Value: arr[1]})
	}

	for i := 0; i < len(info.Append.Headers); i++ {
		arr := strings.Split(info.Append.Headers[i], ":")
		output.Append.Headers = append(output.Append.Headers, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Append.Body); i++ {
		arr := strings.Split(info.Append.Body[i], ":")
		output.Append.Body = append(output.Append.Body, FieldValue{Key: arr[0], Value: arr[1]})
	}
	for i := 0; i < len(info.Append.Querystring); i++ {
		arr := strings.Split(info.Append.Querystring[i], ":")
		output.Append.Querystring = append(output.Append.Querystring, FieldValue{Key: arr[0], Value: arr[1]})
	}

	return output
}

func FromModel(m model.ApiPlugin, ss []model.ApiPluginRelation) (ApiPlugin, error) {
	result := ApiPlugin{
		Id:                  m.Id,
		Name:                m.Name,
		Namespace:           m.Namespace,
		Type:                m.Type,
		ConsumerId:          m.ConsumerId,
		User:                m.User,
		CreatedAt:           m.CreatedAt,
		CreatedAtTimestamp:  m.CreatedAt.Unix(),
		ReleasedAt:          m.ReleasedAt,
		ReleasedAtTimestamp: m.ReleasedAt.Unix(),
		ReplaceUri:          m.ReplaceUri,
		Status:              m.Status,
	}

	switch m.Type {
	case ResponseTransType:
		config := &ResponseTransformer{}
		err := json.Unmarshal([]byte(m.Raw), config)
		if err != nil {
			return result, fmt.Errorf("unmarshal crd v1.api error: %+v", err)
		}
		inputConfig := ToInputTransformerInfo(config.Config)
		outConfig := &OutResponseTransformer{}
		outConfig.Config = inputConfig
		result.Config = outConfig
		klog.Infof("get input result config %+v", config)
	case RequestTransType:
		config := &RequestTransformer{}
		err := json.Unmarshal([]byte(m.Raw), config)
		if err != nil {
			return result, fmt.Errorf("unmarshal crd v1.api error: %+v", err)
		}
		inputConfig := ToInputReqTransformerInfo(config.Config)
		outConfig := &OutRequestTransformer{}
		outConfig.Config = inputConfig
		result.Config = outConfig
		klog.Infof("get input result config %+v", config)
	default:
		klog.Infof("plugin type is invaild config %s", m.Type)
	}

	if ss != nil {
		scenarios := make([]ApiPluginRelation, len(ss))
		for i := range ss {
			scenarios[i] = FromModelScenario(ss[i])
		}
		result.ApiRelation = scenarios
	}

	username, err := cas.GetUserNameByID(m.User)
	if err == nil {
		result.UserName = username
	} else {
		result.UserName = "用户数据错误"
	}
	return result, nil
}

func ToModel(a ApiPlugin) (model.ApiPlugin, []model.ApiPluginRelation, error) {
	apis := make([]model.ApiPluginRelation, len(a.ApiRelation))
	for i := range a.ApiRelation {
		apis[i] = ToModelScenario(a.ApiRelation[i], a.Id, "")
	}
	result := model.ApiPlugin{
		Id:          a.Id,
		Name:        a.Name,
		Type:        a.Type,
		Namespace:   a.Namespace,
		User:        a.User,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		ReleasedAt:  a.ReleasedAt,
		ConsumerId:  a.ConsumerId,
		Status:      a.Status,
		ReplaceUri:  a.ReplaceUri,
	}
	if a.Config != nil {
		err, config := assignmentConfig(a.Type, a.Config)
		if err != nil {
			return result, apis, fmt.Errorf("assignnmentConfig is error,req data: %v", err)
		}
		//**
		b, err := json.Marshal(config)
		if err != nil {
			return result, apis, fmt.Errorf("json.Marshal error,: %v", err)
		}
		switch a.Type {
		case ResponseTransType:
			var resConfig ResponseTransformer
			if err = json.Unmarshal(b, &resConfig); err != nil {
				return result, apis, fmt.Errorf("json.Unmarshal error,: %v", err)
			}
			configJson, err := json.Marshal(resConfig)
			if err != nil {
				return result, apis, fmt.Errorf("json.Marshal resConfig error,: %v", err)
			}
			result.Raw = string(configJson)
			klog.Infof("get result raw resConfig %+v, raw %s", resConfig, result.Raw)
		case RequestTransType:
			var reqConfig RequestTransformer
			if err = json.Unmarshal(b, &reqConfig); err != nil {
				return result, apis, fmt.Errorf("json.Unmarshal error,: %v", err)
			}
			if len(a.ReplaceUri) != 0 {
				reqConfig.Config.Replace.Uri = a.ReplaceUri
			}
			configJson, err := json.Marshal(reqConfig)
			if err != nil {
				return result, apis, fmt.Errorf("json.Marshal reqConfig error,: %v", err)
			}
			result.Raw = string(configJson)
			klog.Infof("get result raw reqConfig %+v, raw %s", reqConfig, result.Raw)
		}
	}
	return result, apis, nil
}

func FromModelScenario(m model.ApiPluginRelation) ApiPluginRelation {
	return ApiPluginRelation{
		Id:           m.Id,
		ApiPluginId:  m.ApiPluginId,
		TargetId:     m.TargetId,
		TargetType:   m.TargetType,
		KongPluginId: m.KongPluginId,
		Status:       m.Status,
		Detail:       m.Detail,
		Enable:       m.Enable,
	}
}

func ToModelScenario(a ApiPluginRelation, productId string, apiId string) model.ApiPluginRelation {
	return model.ApiPluginRelation{
		Id:           a.Id,
		ApiPluginId:  productId,
		TargetId:     a.TargetId,
		TargetType:   a.TargetType,
		KongPluginId: a.KongPluginId,
		Status:       a.Status,
		Detail:       a.Detail,
		Enable:       a.Enable,
	}
}
func assignmentConfig(name string, reqData interface{}) (error, interface{}) {
	klog.Infof("enter assignmentConfig name %s, config %+v", name, reqData)
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData), nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err), nil
	}

	switch name {
	case ResponseTransType:
		var input InputResTransformerConfig
		if err = json.Unmarshal(b, &input); err != nil {
			return fmt.Errorf("json.Unmarshal error,: %v", err), nil
		}
		responseTransformer := &ResponseTransformer{}
		responseTransformer.Name = name

		responseTransformer.Config.Remove = input.Remove
		for i := 0; i < len(input.Rename.Headers); i++ {
			responseTransformer.Config.Rename.Headers = append(responseTransformer.Config.Rename.Headers,
				input.Rename.Headers[i].Key+":"+input.Rename.Headers[i].Value)
		}
		for i := 0; i < len(input.Replace.Headers); i++ {
			responseTransformer.Config.Replace.Headers = append(responseTransformer.Config.Replace.Headers,
				input.Replace.Headers[i].Key+":"+input.Replace.Headers[i].Value)
		}
		for i := 0; i < len(input.Replace.Json); i++ {
			responseTransformer.Config.Replace.Json = append(responseTransformer.Config.Replace.Json,
				input.Replace.Json[i].Key+":"+input.Replace.Json[i].Value)
			responseTransformer.Config.Replace.Json_types = append(
				responseTransformer.Config.Replace.Json_types, input.Replace.Json[i].Type)
		}
		for i := 0; i < len(input.Add.Headers); i++ {
			responseTransformer.Config.Add.Headers = append(responseTransformer.Config.Add.Headers,
				input.Add.Headers[i].Key+":"+input.Add.Headers[i].Value)
		}
		for i := 0; i < len(input.Add.Json); i++ {
			responseTransformer.Config.Add.Json = append(responseTransformer.Config.Add.Json,
				input.Add.Json[i].Key+":"+input.Add.Json[i].Value)
			responseTransformer.Config.Add.Json_types = append(responseTransformer.Config.Add.Json_types,
				input.Add.Json[i].Type)
		}
		klog.V(5).Infof("assignmentConfig target responseTransformer %+v", *responseTransformer)
		return nil, responseTransformer
	case RequestTransType:
		var input InputReqTransformerConfig
		if err = json.Unmarshal(b, &input); err != nil {
			return fmt.Errorf("json.Unmarshal error,: %v", err), nil
		}
		requestTransformer := &RequestTransformer{}
		requestTransformer.Name = name

		//requestTransformer.Config.Remove = input.Remove
		if len(input.HttpMethod) != 0 {
			requestTransformer.Config.HttpMethod = input.HttpMethod
		}
		for i, _ := range input.Remove.Body {
			requestTransformer.Config.Remove.Body = append(requestTransformer.Config.Remove.Body, input.Remove.Body[i])
		}
		for i, _ := range input.Remove.Headers {
			requestTransformer.Config.Remove.Headers = append(requestTransformer.Config.Remove.Headers, input.Remove.Headers[i])
		}
		for i, _ := range input.Remove.Querystring {
			requestTransformer.Config.Remove.Querystring = append(requestTransformer.Config.Remove.Querystring, input.Remove.Querystring[i])
		}

		for i, _ := range input.Rename.Body {
			requestTransformer.Config.Rename.Body = append(requestTransformer.Config.Rename.Body,
				input.Rename.Body[i].Key+":"+input.Rename.Body[i].Value)
		}
		for i, _ := range input.Rename.Headers {
			requestTransformer.Config.Rename.Headers = append(requestTransformer.Config.Rename.Headers,
				input.Rename.Headers[i].Key+":"+input.Rename.Headers[i].Value)
		}
		for i, _ := range input.Rename.Querystring {
			requestTransformer.Config.Rename.Querystring = append(requestTransformer.Config.Rename.Querystring,
				input.Rename.Querystring[i].Key+":"+input.Rename.Querystring[i].Value)
		}

		for i, _ := range input.Replace.Body {
			requestTransformer.Config.Replace.Body = append(requestTransformer.Config.Replace.Body,
				input.Replace.Body[i].Key+":"+input.Replace.Body[i].Value)
		}
		for i, _ := range input.Replace.Headers {
			requestTransformer.Config.Replace.Headers = append(requestTransformer.Config.Replace.Headers,
				input.Replace.Headers[i].Key+":"+input.Replace.Headers[i].Value)
		}
		for i, _ := range input.Replace.Querystring {
			requestTransformer.Config.Replace.Querystring = append(requestTransformer.Config.Replace.Querystring,
				input.Replace.Querystring[i].Key+":"+input.Replace.Querystring[i].Value)
		}
		//requestTransformer.Config.Replace.Uri = i

		for i, _ := range input.Add.Body {
			requestTransformer.Config.Add.Body = append(requestTransformer.Config.Add.Body,
				input.Add.Body[i].Key+":"+input.Add.Body[i].Value)
		}
		for i, _ := range input.Add.Headers {
			requestTransformer.Config.Add.Headers = append(requestTransformer.Config.Add.Headers,
				input.Add.Headers[i].Key+":"+input.Add.Headers[i].Value)
		}
		for i, _ := range input.Add.Querystring {
			requestTransformer.Config.Add.Querystring = append(requestTransformer.Config.Add.Querystring,
				input.Add.Querystring[i].Key+":"+input.Add.Querystring[i].Value)
		}

		for i, _ := range input.Append.Body {
			requestTransformer.Config.Append.Body = append(requestTransformer.Config.Append.Body,
				input.Append.Body[i].Key+":"+input.Append.Body[i].Value)
		}
		for i, _ := range input.Append.Headers {
			requestTransformer.Config.Append.Headers = append(requestTransformer.Config.Append.Headers,
				input.Append.Headers[i].Key+":"+input.Append.Headers[i].Value)
		}
		for i, _ := range input.Append.Querystring {
			requestTransformer.Config.Append.Querystring = append(requestTransformer.Config.Append.Querystring,
				input.Append.Querystring[i].Key+":"+input.Append.Querystring[i].Value)
		}
		klog.V(5).Infof("assignmentConfig target responseTransformer %+v", *requestTransformer)
		return nil, requestTransformer
	}
	klog.V(5).Infof("assignmentConfig target error")
	return nil, nil
}

/////////////////////
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

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}
