package apis

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtEncodingIn struct {
	Params    map[string]interface{}
	JwtSecret string
	// SigningMethod is the signing method for the JWT token
	// Possible values are HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512, PS256, PS384, PS512
	SignMethod string
	Duration   int64                  // seconds of expire time, format is unix time
	Other      map[string]interface{} `json:"remain" msgpack:"-" `
}

func ApiJwtSign(in *JwtEncodingIn) (AccessToken string, err error) {
	if in == nil {
		return "", errors.New("input parameter is nil")
	}
	if in.Params == nil {
		return "", errors.New("params is nil")
	}
	if in.JwtSecret == "" {
		return "", errors.New("jwt secret is empty")
	}
	if in.SignMethod == "" {
		return "", errors.New("sign method is empty")
	}

	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Second * time.Duration(in.Duration)).Unix(),
	}
	//merge in.Params to claims
	for k, v := range in.Params {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(string(in.SignMethod)), claims)
	tokenString, err := token.SignedString([]byte(in.JwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type JwtDecodingIn struct {
	Token     string
	JwtSecret string
}

type JwtDecodingOut struct {
	Params  map[string]interface{} `json:"-"`
	Expired bool                   `json:"-"`
}

func ApiJwtParse(in *JwtDecodingIn) (out *jwt.Token, err error) {
	if in == nil {
		return nil, errors.New("input parameter is nil")
	}
	if in.Token == "" {
		return nil, errors.New("token is empty")
	}
	if in.JwtSecret == "" {
		return nil, errors.New("jwt secret is empty")
	}
	//parse with JwtSecret
	return jwt.ParseWithClaims(in.Token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(in.JwtSecret), nil
	})
}
