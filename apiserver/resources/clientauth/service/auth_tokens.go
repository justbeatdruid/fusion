package service

import (
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"time"
)

const (
	D_EXPTIME_IN_DAYS = 30 //token默认30天过期
)
//Token的结构体
type Token struct {
	Sub string `json:"sub"` //token的主题
	Alg string `json:"alg"` //token签名算法
	Iat string `json:"iat"` //token签发时间
	Exp string `json:"exp"` //token过期时间
}

//创建token
func (t *Token) Create() (string, error) {
	//获取当前路径
	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(path + "/key/my-secret.key")
	if err != nil {
		return "", err
	}

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	header := make(map[string]interface{})

	if len(t.Alg) == 0 {
		//默认用HS256算法
		header["alg"] = jwt.SigningMethodHS256.Name
	} else {
		header["alg"] = t.Alg
	}

	claims := make(jwt.MapClaims)
	claims["sub"] = t.Sub

	now := time.Now()
	//token签发时间
	claims["iat"] = now.Unix()
	if len(t.Exp) == 0 {
		//未设置过期时间，默认过期30天
		claims["exp"] = now.AddDate(0, 0, D_EXPTIME_IN_DAYS).Unix()

	} else {
		claims["exp"] = t.Exp
	}

	jwtToken.Claims = claims
	jwtToken.Header = header
	ts, err := jwtToken.SignedString(content)
	if err != nil {
		return "", err
	}
	return ts, nil
}
