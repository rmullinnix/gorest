package authorizers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/rmullinnix/gorest"
	"github.com/rmullinnix/logger"
	"time"
)

var keystore		map[string]interface{}
var curScheme		string

func Oauth2Jwt(token string, scheme string, scopes []string, method string, rb *gorest.ResponseBuilder) bool {

	logger.Info.Println("oauth2jwt", token, scheme, scopes)
	curScheme = scheme

	jwtToken, err := jwt.Parse(token, jwtKey)

	curScheme = ""

	if err != nil {
		logger.Error.Println("jwt error", err)
		return false
	}


	claim, found := jwtToken.Claims["scope"]
	if !found {
		logger.Error.Println("No scope claims in the token" )
		return false
	}
	
	logger.Error.Println("claim", claim)
	arrClaim := claim.([]interface{})

	authorized := false
	for i := range scopes {
		for j := range arrClaim {
			if scopes[i] == arrClaim[j].(string) {
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
		logger.Error.Println("Key not found")
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
	keyIndex := "default"

	if item, found := token.Header["kid"]; found {
		keyIndex = item.(string)
	}

	if key := GetKey(curScheme, keyIndex); key == nil {
		return nil, errors.New("key for " + token.Header["kid"].(string) + " does not exist")
	} else {
		return key, nil
	}
}
