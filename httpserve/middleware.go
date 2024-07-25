package httpserve

import (
	"fmt"
	"strings"
)

func (svc *DoptimeReqCtx) UpdateKeyFieldWithJwtClaims() (operation string, err error) {
	var (
		ok     bool
		obj    interface{}
		subTag string
		f64    float64
	)
	operation = strings.ToLower(svc.Cmd)
	KeyContainsAt := strings.Contains(svc.Key, "@")
	FieldContainsAt := strings.Contains(svc.Field, "@")
	if !KeyContainsAt && !FieldContainsAt {
		return operation, nil
	}
	operation = "@" + operation

	if svc.Claims == nil {
		return operation, fmt.Errorf("JWT token is nil")
	}
	// Field contains @*, replace @* with jwt value
	// 只要设置的时候，有@id,@pub，可以确保写不越权，因为 是"@" + operation
	if FieldContainsAt {
		FieldParts := strings.Split(svc.Field, "@")
		if subTag = FieldParts[len(FieldParts)-1]; len(subTag) == 0 {
			return operation, fmt.Errorf("jwt missing subTag " + subTag)
		}
		if obj, ok = svc.Claims[subTag]; !ok {
			return operation, fmt.Errorf("jwt missing subTag " + subTag)
		}
		// if 64 is int, convert to int
		if f64, ok = obj.(float64); ok && f64 == float64(int64(f64)) {
			obj = int64(f64)
		}
		FieldParts[len(FieldParts)-1] = fmt.Sprintf("%v", obj)
		svc.Field = strings.Join(FieldParts, "")
	}
	if KeyContainsAt {
		KeyParts := strings.Split(svc.Key, "@")
		if subTag = KeyParts[len(KeyParts)-1]; len(subTag) == 0 {
			return operation, fmt.Errorf("jwt missing subTag " + subTag)
		}
		if obj, ok = svc.Claims[subTag]; !ok {
			return operation, fmt.Errorf("jwt missing subTag " + subTag)
		}
		// if 64 is int, convert to int
		if f64, ok = obj.(float64); ok && f64 == float64(int64(f64)) {
			obj = int64(f64)
		}
		KeyParts[len(KeyParts)-1] = fmt.Sprintf("%v", obj)
		svc.Key = strings.Join(KeyParts, "")
	}
	return operation, nil
}
