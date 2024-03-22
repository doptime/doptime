package httpserve

import (
	"errors"
	"fmt"
	"strings"

	"github.com/doptime/doptime/config"

	"github.com/golang-jwt/jwt/v5"
)

func (svc *HttpContext) ParseJwtToken() (err error) {
	var (
		jwtStr string
	)
	if svc.jwtToken != nil {
		return nil
	}
	if jwtStr = svc.Req.Header.Get("Authorization"); len(jwtStr) == 0 {
		return errors.New("no JWT token")
	}
	//decode jwt string to map[string] interface{} with jwtSrcrets as jwt secret
	keyFunction := func(token *jwt.Token) (value interface{}, err error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(config.Cfg.Http.JwtSecret), nil
	}
	if svc.jwtToken, err = jwt.ParseWithClaims(jwtStr, jwt.MapClaims{}, keyFunction); err != nil {
		return fmt.Errorf("invalid JWT token: %v", err)
	}
	//for each element of svc.jwtToken ,if it's type if flaot64, convert it to int64, it it's
	if mpclaims, ok := svc.jwtToken.Claims.(jwt.MapClaims); ok {
		for k, v := range mpclaims {
			if f64, ok := v.(float64); ok && f64 == float64(int64(f64)) {
				mpclaims[k] = int64(f64)
			}
		}
	}
	return nil
}
func (svc *HttpContext) MergeJwtField(paramIn map[string]interface{}) {
	//remove nay field that starts with "Jwt" in paramIn
	//prevent forged jwt field
	for k := range paramIn {
		if strings.HasPrefix(k, "Jwt") {
			delete(paramIn, k)
		}
	}

	if err := svc.ParseJwtToken(); err != nil {
		return
	}
	//save every field in svc.Jwt.Claims to in
	mpclaims, ok := svc.jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return
	}
	for k, v := range mpclaims {
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
