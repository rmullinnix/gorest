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
		logger.Error.Println("[sec] oauth2-jwt userid: unknown useruuid: unknown active: true locked: false auth: false failcnt: 0 response: 401 reason: jwt parse error", err)
		return false
	}

	uid := "unknown"
	uuid := "unknown"
	if user, ufnd := jwtToken.Claims["user"]; ufnd {
		uid = user.(string)
		rb.Session().Set("UserId", uid)
	}

	if userUUID, uifnd := jwtToken.Claims["useruuid"]; uifnd {
		uuid = userUUID.(string)
		rb.Session().Set("UserUUID", uuid)
	}

	claim, found := jwtToken.Claims["scope"]
	if !found {
		logger.Error.Println("[sec] oauth2-jwt userid: " + uid + " useruuid: " + uuid + " active: true locked: false auth: false failcnt: 0 response: 401 reason: No scope claims in the token")
		return false
	}
	
	arrClaim := claim.([]interface{})
	rb.Session().Set("Scope", arrClaim)

	authorized := false
	for i := range scopes {
		// just interested in a valid jwt with no specific privileges
		if scopes[i] == "<valid>" {
			authorized = true
			break
		}

		contextAuth := -1
		contextKey := ""
		scopeName := scopes[i]
		if contextAuth = strings.Index(scopes[i], "["); contextAuth > -1 {
			contextKey = scopes[i][contextAuth + 1 : strings.Index(scopes[i], "]")]
			scopeName = scopes[i][:contextAuth]
		}

		for j := range arrClaim {
			arrStr := arrClaim[j].(string)

			if strings.HasPrefix(arrStr, scopeName) {
				if contextAuth = strings.Index(arrStr, "["); contextAuth > -1 {
					rb.Session().Set("ScopeContext", arrStr[contextAuth + 1 : len(arrStr) - 1])
					keys := strings.Split(arrStr[contextAuth + 1 : len(arrStr) - 1], ",")

					// restricted list, filtered in application code
					if contextKey == "" {
						authorized = true
						break
					}

					for k := range keys {
						if keys[k] == contextKey {
							authorized = true
							break
						}
					}
					if authorized {
						break
					}
				} else {
					authorized = true
					break
				}

			}
			
			if len(contextKey) > 0 {
				if scopes[i] == arrClaim[j].(string) {
					authorized = true
					break
				}
			}
		}
	}

	if !authorized {
		arr := make([]string, len(arrClaim))
		for j := range arrClaim {
			arr[j] = arrClaim[j].(string)
		}
		logger.Error.Println("[sec] oauth2-jwt userid: " + uid + " useruuid: " + uuid + " active: true locked: false auth: false failcnt: 0 response: 401 reason: user not authorized for scope " + strings.Join(arr, ", "))
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

	return token.SignedString(signingKey.key)
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
