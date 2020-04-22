package service

import (
	"github.com/dgrijalva/jwt-go"
)

//Token的结构体
type Token struct {
	Sub    string `json:"sub"`    //token的主题
	Alg    string `json:"alg"`    //token签名算法
	Iat    int64  `json:"iat"`    //token签发时间
	Exp    int64  `json:"exp"`    //token过期时间
	Secret string `json:"secret"` //token secret key
}

//创建token
func (t *Token) Create() (string, error) {
	var content []byte = []byte(t.Secret)
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

	if t.Exp != 0 {
		claims["iat"] = t.Iat
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
