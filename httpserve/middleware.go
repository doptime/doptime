package httpserve

import (
	"fmt"
	"strings"
)

func (svc *DoptimeReqCtx) UpdateKeyFieldWithJwtClaims() (operation string, err error) {
	if svc.Claims == nil {
		return operation, fmt.Errorf("JWT token is nil")
	}
	skey := strings.Split(svc.Key, ":")[0]
	skey = strings.Split(skey, "@")[0]
	operation = strings.ToLower(svc.Cmd) + "::" + skey

	// Field contains @*, replace @* with jwt value
	// 只要设置的时候，有@id,@pub，可以确保写不越权，因为 是"@" + operation
	keyParts, fieldPars := strings.Split(svc.Key, "@"), strings.Split(svc.Field, "@")
	if len(keyParts) > 1 {
		operation += "@" + strings.Join(keyParts[1:], "@")
		svc.Key, err = replaceAtWithJwtClaims(svc.Claims, keyParts)
		if err != nil {
			return operation, err
		}
	}
	if len(fieldPars) > 1 {
		operation += "::@" + strings.Join(fieldPars[1:], "@")
		svc.Field, err = replaceAtWithJwtClaims(svc.Claims, fieldPars)
		if err != nil {
			return operation, err
		}
	}
	return operation, nil
}

func replaceAtWithJwtClaims(claims map[string]interface{}, KeyParts []string) (newKey string, err error) {
	var (
		ok         bool
		obj        interface{}
		strBuilder strings.Builder
		f64        float64
	)
	strBuilder.WriteString(KeyParts[0])
	for i, l := 1, len(KeyParts); i < l; i++ {
		keyPart := KeyParts[i]
		if obj, ok = claims[keyPart]; !ok {
			return "", fmt.Errorf("jwt missing key " + keyPart[1:])
		} else {
			// if 64 is int, convert to int
			if f64, ok = obj.(float64); ok && f64 == float64(int64(f64)) {
				obj = int64(f64)
			}
			strBuilder.WriteString(fmt.Sprintf("%v", obj))
		}
	}
	return strBuilder.String(), nil
}
