package httpserve

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/doptime/config/cfghttp"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/golang-jwt/jwt/v5"
)

var tokenKeyRSA interface{}

func ParseAndValidateToken(jwtToken string, secret string) (jwt.MapClaims, error) {
	// 1. Configure parser: UseJSONNumber prevents float64 precision issues
	parser := jwt.NewParser(jwt.WithJSONNumber())

	token, err := parser.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		// Security check: restrict alg types (no "none" alg allowed)
		switch token.Method.(type) {
		case *jwt.SigningMethodHMAC:
			// HS256: secret is the raw key bytes
			return []byte(secret), nil

		case *jwt.SigningMethodRSA:
			// RS256: secret is the PEM Public Key
			if tokenKeyRSA != nil {
				return tokenKeyRSA, nil
			}
			block, _ := pem.Decode([]byte(secret))
			if block == nil {
				return nil, errors.New("failed to parse PEM block")
			}
			pub, err := x509.ParsePKIXPublicKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse DER key: %v", err)
			}
			tokenKeyRSA = pub
			return pub, nil

		default:
			return nil, fmt.Errorf("unexpected alg: %v", token.Header["alg"])
		}
	})

	// 2. Check errors (Parse automatically handles exp, nbf, and signature verification)
	if err != nil {
		return nil, err
	}

	// 3. Extract and return claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

var tokenCache, _ = lru.New[string, jwt.MapClaims](10000)

func (svc *DoptimeReqCtx) ParseJwtClaim(r *http.Request) (err error) {
	var ok bool
	jwtToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(jwtToken) == 0 {
		return nil
	}
	//fast return from cache
	if svc.Claims, ok = tokenCache.Get(jwtToken); ok {
		var exp int64
		switch v := svc.Claims["exp"].(type) {
		case float64:
			exp = int64(v)
		case int64:
			exp = v
		case string:
			exp, err = strconv.ParseInt(v, 10, 64)
		case json.Number:
			exp, err = strconv.ParseInt(v.String(), 10, 64)
		default:
			goto parsecontinue
		}
		if exp < time.Now().Unix() {
			return errors.New("JWT token is expired")
		}
		return nil
	}
parsecontinue:
	//parse jwt token
	if svc.Claims, err = ParseAndValidateToken(jwtToken, cfghttp.JWTSecret); err != nil {
		return fmt.Errorf("invalid JWT token: %v", err)
	}
	//save jwt token to cache
	tokenCache.Add(jwtToken, svc.Claims)
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
	jwtString, err = token.SignedString([]byte(cfghttp.JWTSecret))
	return jwtString, err
}
