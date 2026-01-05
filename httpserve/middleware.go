package httpserve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/doptime/redisdb"
	"github.com/google/uuid"
)

func (svc *DoptimeReqCtx) ReplaceKeyFieldTagWithJwtClaims() (err error) {
	if svc.Key, err = svc.replaceTags(svc.Key); err != nil {
		return err
	}
	for i, field := range svc.Fields {
		if svc.Fields[i], err = svc.replaceTags(field); err != nil {
			return err
		}
	}
	return nil
}
func (svc *DoptimeReqCtx) replaceTags(input string) (string, error) {
	var (
		val interface{}
		ok  bool
	)
	parts := strings.Split(input, "@")
	if len(parts) <= 1 {
		return input, nil
	}

	var sb strings.Builder
	// 写入第一个 @ 之前的原始文本
	sb.WriteString(parts[0])
	// 遍历 @ 之后的每一个标签名
	for _, tag := range parts[1:] {
		switch tag {
		case "uuid":
			val = uuid.New().String()
			continue
		case "nanoid":
			n := 21
			for i := 6; i < len(tag); i++ {
				if tag[i] < '0' || tag[i] > '9' {
					continue
				}
				if m, err := strconv.Atoi(tag[i:]); err == nil {
					n = max(m, 1, min(m, 21))
					break
				}
			}
			val = redisdb.NanoId(n)
		default:
			if svc.JwtClaims == nil {
				return "", fmt.Errorf("JWT token is nil")
			}
			val, ok = svc.JwtClaims[tag]
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

		sb.WriteString(fmt.Sprint(val))
	}

	return sb.String(), nil
}
