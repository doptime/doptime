package httpserve

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/doptime/doptime/config"

	"github.com/golang-jwt/jwt/v5"
)

func Jwt2ClaimMap(jwtStr string, secret string) (mpclaims jwt.MapClaims, err error) {
	//decode jwt string to map[string] interface{} with jwtSrcrets as jwt secret
	keyFunction := func(token *jwt.Token) (value interface{}, err error) {
		return []byte(secret), nil
	}
	jwtToken, err := jwt.ParseWithClaims(jwtStr, jwt.MapClaims{}, keyFunction)
	if err != nil {
		return nil, err
	}
	var ok bool
	if mpclaims, ok = jwtToken.Claims.(jwt.MapClaims); !ok {
		return nil, errors.New("invalid JWT token")
	}
	for k, v := range mpclaims {
		if f64, ok := v.(float64); ok && f64 == float64(int64(f64)) {
			mpclaims[k] = int64(f64)
		}
	}
	return mpclaims, nil
}

func (svc *DoptimeReqCtx) ParseJwtToken(r *http.Request) (err error) {
	jwtStr := r.Header.Get("Authorization")
	jwtStr = strings.TrimPrefix(jwtStr, "Bearer ")
	if len(jwtStr) == 0 {
		return errors.New("no JWT token")
	}
	if svc.Claims, err = Jwt2ClaimMap(jwtStr, config.Cfg.Http.JwtSecret); err != nil {
		return fmt.Errorf("invalid JWT token: %v", err)
	}
	return nil
}
func (svc *DoptimeReqCtx) MergeJwtField(r *http.Request, paramIn map[string]interface{}) {
	//remove nay field that starts with "Jwt" in paramIn
	//prevent forged jwt field
	for k := range paramIn {
		if strings.HasPrefix(k, "Jwt") {
			delete(paramIn, k)
		}
	}

	if err := svc.ParseJwtToken(r); err != nil {
		return
	}
	//save every field in svc.Jwt.Claims to in
	if svc.Claims == nil {
		return
	}
	for k, v := range svc.Claims {
		//convert first letter of k to upper case
		k = strings.ToUpper(k[:1]) + k[1:]
		paramIn["Jwt"+k] = v
	}
}

func ConvertMapToJwtString(param map[string]interface{}) (jwtString string, err error) {
	//convert map to jwt.claims
	claims := jwt.MapClaims{}
	for k, v := range param {
		claims[k] = v
	}
	//create jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//sign jwt token
	jwtString, err = token.SignedString([]byte(config.Cfg.Http.JwtSecret))
	return jwtString, err
}
