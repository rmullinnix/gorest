package authorizers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/rmullinnix/gorest"
	"time"
)

var keystore		map[string]interface{}
var curScheme		string

func Oauth2Jwt(token string, scheme string, scopes []string, method string, rb *gorest.ResponseBuilder) bool {

	curScheme = scheme

	jwtToken, err := jwt.Parse(token, jwtKey)

	curScheme = ""

	if err != nil {
		return false
	}


	claim, found := jwtToken.Claims["scope"]
	if ! found {
		return false
	}
	
	arrClaim := claim.([]string)

	authorized := false
	for i := range scopes {
		for j := range arrClaim {
			if scopes[i] == arrClaim[j] {
				authorized = true
				break
			}
		}
	}

	return authorized
}

func AddKey(scheme string, keyid string, key interface{}) {
	if keystore == nil {
		keystore = make(map[string]interface{})
	}
	mapKey := scheme + ":" + keyid
	keystore[mapKey] = key
}

func GetKey(scheme string, keyid string) interface{} {
	mapKey := scheme + ":" + keyid
	key, found := keystore[mapKey]
	if found {
		return key
	} else {
		return nil
	}
}

func SetSigningKey(key interface{}) {
	AddKey("", "sign", key)
}

func NewToken(method jwt.SigningMethod, userId string, scopes []string, expireMins int) (string, error) {
	token := jwt.New(method)

	token.Claims["scope"] = scopes
	token.Claims["exp"] = time.Now().Add(time.Minute * time.Duration(expireMins)).Unix()
	token.Claims["user"] = userId

	signingKey := GetKey("", "sign")

	return token.SignedString(signingKey)
}

func jwtKey(token *jwt.Token) (interface{}, error) {
	if key := GetKey(curScheme, token.Header["kid"].(string)); key == nil {
		return nil, errors.New("key for " + token.Header["kid"].(string) + " does not exist")
	} else {
		return key, nil
	}
}
