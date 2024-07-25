package httpserve

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/doptime/doptime/config"
	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/golang-jwt/jwt/v5"
)

func Jwt2Claim(jwtStr string, secret string) (mpclaims jwt.MapClaims, err error) {
	//decode jwt string to map[string] interface{} with jwtSrcrets as jwt secret
	keyFunction := func(token *jwt.Token) (value interface{}, err error) {
		return []byte(secret), nil
	}
	var jwtToken *jwt.Token
	jwtToken, err = jwt.ParseWithClaims(jwtStr, jwt.MapClaims{}, keyFunction)
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
	//ensure there's exp field in jwt token
	if _, ok := mpclaims["exp"].(int64); !ok {
		//set expiration time to maxInt64 if not found
		mpclaims["exp"] = math.MaxInt64
		if exp, ok := jwtToken.Header["exp"].(int64); ok {
			mpclaims["exp"] = exp
		}
	}
	return mpclaims, nil
}

var mapClaims cmap.ConcurrentMap[string, jwt.MapClaims] = cmap.New[jwt.MapClaims]()

func (svc *DoptimeReqCtx) ParseJwtClaim(r *http.Request) (err error) {
	var ok bool
	jwtStr := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(jwtStr) == 0 {
		return errors.New("no JWT token")
	}
	if svc.Claims, ok = mapClaims.Get(jwtStr); ok {
		exp := svc.Claims["exp"].(int64)
		if exp < time.Now().Unix() {
			return errors.New("JWT token is expired")
		}
	} else if svc.Claims, err = Jwt2Claim(jwtStr, config.Cfg.Http.JwtSecret); err != nil {
		return fmt.Errorf("invalid JWT token: %v", err)
	} else {
		mapClaims.Set(jwtStr, svc.Claims)
	}
	return nil
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
