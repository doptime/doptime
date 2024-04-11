package lib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SigningMethod string

const (
	SigningMethodNone SigningMethod = "none"

	SigningMethodHS256 SigningMethod = "HS256"
	SigningMethodHS384 SigningMethod = "HS384"
	SigningMethodHS512 SigningMethod = "HS512"

	SigningMethodRS256 SigningMethod = "RS256"
	SigningMethodRS384 SigningMethod = "RS384"
	SigningMethodRS512 SigningMethod = "RS512"

	SigningMethodES256 SigningMethod = "ES256"
	SigningMethodES384 SigningMethod = "ES384"
	SigningMethodES512 SigningMethod = "ES512"

	SigningMethodPS256 SigningMethod = "PS256"
	SigningMethodPS384 SigningMethod = "PS384"
	SigningMethodPS512 SigningMethod = "PS512"
)

type JwtEncodingIn struct {
	Params     map[string]interface{}
	JwtSecret  string
	SignMethod SigningMethod
	Duration   int64                  // seconds of expire time, format is unix time
	Other      map[string]interface{} `mapstructure:",remain" msgpack:"-" `
}

type JwtEncodingOut struct {
	Token string `json:"token"`
}

func ApiJwtSign(in *JwtEncodingIn) (out *JwtEncodingOut, err error) {
	if in == nil {
		return nil, errors.New("input parameter is nil")
	}
	if in.Params == nil {
		return nil, errors.New("params is nil")
	}
	if in.JwtSecret == "" {
		return nil, errors.New("jwt secret is empty")
	}
	if in.SignMethod == "" {
		return nil, errors.New("sign method is empty")
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
		return nil, err
	}

	return &JwtEncodingOut{Token: tokenString}, nil
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
