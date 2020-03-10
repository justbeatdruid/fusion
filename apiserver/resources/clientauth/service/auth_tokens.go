package service

import (
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"os"
	"time"
)

const (
	D_EXPTIME_IN_DAYS = 30
)

type Token struct {
	Sub string `json:"sub"`
	Alg string `json:"alg"`
	Iat string `json:"iat"`
	Exp string `json:"exp"`
}

//创建token
func (t *Token) Create() (string, error) {
	//获取当前路径
	path, _ := os.Getwd()

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
