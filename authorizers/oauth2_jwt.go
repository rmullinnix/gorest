package authorizers

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/rmullinnix/gorest"
	"github.com/rmullinnix/logger"
	"strings"
	"time"
)

var keystore		map[string]signingKey

type signingKey struct {
	key		interface{}
	signType	string
}

var curScheme		string

func Oauth2Jwt(token string, scheme string, scopes []string, method string, rb *gorest.ResponseBuilder) bool {

	curScheme = scheme

	jwtToken, err := jwt.Parse(token, jwtKey)

	curScheme = ""

	if err != nil {
		logger.Error.Println("jwt error", err)
		return false
	}

	if user, ufnd := jwtToken.Claims["user"]; ufnd {
		rb.Session().Set("UserId", user.(string))
	}

	if userUUID, uifnd := jwtToken.Claims["useruuid"]; uifnd {
		rb.Session().Set("UserUUID", userUUID.(string))
	}

	claim, found := jwtToken.Claims["scope"]
	if !found {
		logger.Error.Println("No scope claims in the token" )
		return false
	}
	
	arrClaim := claim.([]interface{})

	authorized := false
	for i := range scopes {
		// just interested in a valid jwt with no specific privileges
		if scopes[i] == "<valid>" {
			authorized = true
			break
		}

		contextAuth := -1
		contextKey := ""
		if contextAuth = strings.Index(scopes[i], "["); contextAuth > -1 {
			contextKey = scopes[i][contextAuth + 1 : strings.Index(scopes[i], "]")]
		}

		for j := range arrClaim {
			if len(contextKey) > 0 {
				arrStr := arrClaim[j].(string)
				if strings.HasPrefix(arrStr, scopes[i][:contextAuth]) {
					keys := strings.Split(arrStr[contextAuth + 1 : len(arrStr) - 1], ",")
					for k := range keys {
						if keys[k] == contextKey {
							authorized = true
							break
						}
					}
					if authorized {
						break
					}
				}
			} else {
				if scopes[i] == arrClaim[j].(string) {
					authorized = true
					break
				}
			}
		}
	}

	return authorized
}

func AddKey(scheme string, keyid string, key interface{}, signType string) {
	if keystore == nil {
		keystore = make(map[string]signingKey)
	}
	var sKey	signingKey

	sKey.key = key
	sKey.signType = signType
	mapKey := scheme + ":" + keyid
	keystore[mapKey] = sKey
}

func getKey(scheme string, keyid string) *signingKey {
	mapKey := scheme + ":" + keyid
	key, found := keystore[mapKey]
	if found {
		return &key
	} else {
		logger.Error.Println("Key not found")
		return nil
	}
}

func SetSigningKey(key interface{}) {
	AddKey("", "sign", key, "RSA")
}

func NewToken(method jwt.SigningMethod, userId string, userUUID string, scopes []string, expireMins int) (string, error) {
	token := jwt.New(method)

	token.Claims["scope"] = scopes
	token.Claims["exp"] = time.Now().Add(time.Minute * time.Duration(expireMins)).Unix()
	token.Claims["user"] = userId
	token.Claims["useruuid"] = userUUID

	signingKey := getKey("", "sign")

	return token.SignedString(signingKey)
}

func jwtKey(token *jwt.Token) (interface{}, error) {
	keyIndex := "default"

	if item, found := token.Header["kid"]; found {
		keyIndex = item.(string)
	}

	if sKey := getKey(curScheme, keyIndex); sKey != nil {
		if sKey.signType == "RSA"  {
			if token.Method == jwt.SigningMethodRS256 || token.Method == jwt.SigningMethodRS384 || token.Method == jwt.SigningMethodRS512 {
				return sKey.key, nil
			} else {
				return nil, errors.New("invalid signing method for key")
			}
		} else if sKey.signType == "HMAC" {
			if token.Method == jwt.SigningMethodHS256 || token.Method == jwt.SigningMethodHS384 || token.Method == jwt.SigningMethodHS512 {
				return sKey.key, nil
			} else {
				return nil, errors.New("invalid signing method for key")
			}
		} else {
			return nil, errors.New("invalid signing algorithm on key in keystore")
		}
	}
	return nil, errors.New("key for " + token.Header["kid"].(string) + " does not exist")
}
