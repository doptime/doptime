package httpserve

import (
	"fmt"
	"strings"
)

func (svc *DoptimeReqCtx) ReplaceKeyFieldTagWithJwtClaims() (err error) {
	//return if no @ in key or field to be replaced
	if !strings.Contains(svc.Key, "@") && !strings.Contains(svc.Field, "@") {
		return nil
	}

	if svc.Claims == nil {
		return fmt.Errorf("JWT token is nil")
	}
	if svc.Key, err = svc.replaceTags(svc.Key); err != nil {
		return err
	}
	if svc.Field, err = svc.replaceTags(svc.Field); err != nil {
		return err
	}
	return nil
}
func (svc *DoptimeReqCtx) replaceTags(input string) (string, error) {
	parts := strings.Split(input, "@")
	if len(parts) <= 1 {
		return input, nil
	}

	var sb strings.Builder
	// 写入第一个 @ 之前的原始文本
	sb.WriteString(parts[0])

	// 遍历 @ 之后的每一个标签名
	for _, tag := range parts[1:] {
		val, ok := svc.Claims[tag]
		if !ok {
			return "", fmt.Errorf("jwt missing key: %s", tag)
		}

		// 处理数字精度问题（JSON 默认将数字解析为 float64）
		if f64, isFloat := val.(float64); isFloat && f64 == float64(int64(f64)) {
			sb.WriteString(fmt.Sprintf("%d", int64(f64)))
		} else {
			sb.WriteString(fmt.Sprint(val))
		}
	}

	return sb.String(), nil
}
