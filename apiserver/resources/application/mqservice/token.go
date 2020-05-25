package mqservice

import "github.com/dgrijalva/jwt-go"

type Token struct {
	Secret string `json:"secret"`
}

func NewToken(secret string) *Token{
	return &Token{secret}
}
func (t *Token) Create(username string) (*string, error) {
	var content = []byte(t.Secret)
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
		return nil, err
	}
	return &ts, nil
}