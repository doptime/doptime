package mapper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Marshal 是对标准库 json.Marshal 的替代，专门用于支持 doptime 的特殊 Tag 格式。
// 特性：
// 1. Tag 使用空格分隔 (Space Separator)。
// 2. 忽略以 '@' 开头的指令 (如 @100, @@field)。
// 3. 默认将字段名转换为小写 (如果 Tag 未指定名称)。
func Marshal(v interface{}) ([]byte, error) {
	// 使用 bytes.Buffer 高效拼接
	buf := &bytes.Buffer{}
	if err := encodeValue(buf, reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeValue 递归处理各种类型
func encodeValue(buf *bytes.Buffer, v reflect.Value) error {
	if !v.IsValid() {
		buf.WriteString("null")
		return nil
	}

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			buf.WriteString("null")
			return nil
		}
		return encodeValue(buf, v.Elem())

	case reflect.Struct:
		return encodeStruct(buf, v)

	case reflect.Map:
		return encodeMap(buf, v)

	case reflect.Slice, reflect.Array:
		return encodeSlice(buf, v)

	case reflect.String:
		// 借用标准库来处理字符串的转义（引号、换行符等），保证安全性
		safeStr, _ := json.Marshal(v.String())
		buf.Write(safeStr)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprintf(buf, "%d", v.Int())
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprintf(buf, "%d", v.Uint())
		return nil

	case reflect.Float32, reflect.Float64:
		// 保持与 JSON 数字格式一致，避免科学计数法造成的困扰
		val := v.Float()
		// 如果是整数，去掉小数点
		if val == float64(int64(val)) {
			fmt.Fprintf(buf, "%.0f", val)
		} else {
			// 使用 %g 去掉末尾多余的0
			fmt.Fprintf(buf, "%g", val)
		}
		return nil

	case reflect.Bool:
		if v.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil

	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

// encodeStruct 核心逻辑：解析你的特殊 Tag 格式
func encodeStruct(buf *bytes.Buffer, v reflect.Value) error {
	t := v.Type()
	buf.WriteByte('{')

	first := true
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 忽略未导出字段
		if field.PkgPath != "" {
			continue
		}

		// 获取 Tag 并按照 mapper.go 的逻辑解析
		tagVal := field.Tag.Get("json")
		if tagVal == "-" {
			continue
		}

		jsonKey := parseEncoderTag(field.Name, tagVal)

		if !first {
			buf.WriteByte(',')
		}
		first = false

		// 写入 Key
		buf.WriteByte('"')
		buf.WriteString(jsonKey)
		buf.WriteString("\":")

		// 递归写入 Value
		if err := encodeValue(buf, v.Field(i)); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

// parseEncoderTag 复刻 mapper.go 的解析逻辑，但用于编码
func parseEncoderTag(fieldName string, tag string) string {
	// 1. 如果没有 Tag，直接返回小写字段名
	if tag == "" {
		return strings.ToLower(fieldName)
	}

	// 2. 按空格分割 (Space Separator)
	parts := strings.Fields(tag)

	// 3. 提取名称部分
	if len(parts) > 0 {
		candidate := parts[0]
		// 如果第一部分不是指令（不以 @ 开头），则它就是名字
		if !strings.HasPrefix(candidate, "@") {
			return candidate
		}
	}

	// 4. 如果 Tag 全是指令（如 "@100"）或为空，回退到字段名小写
	return strings.ToLower(fieldName)
}

func encodeMap(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('{')

	// 为了输出确定性（和 json.Marshal 行为一致），我们需要对 Key 进行排序
	keys := v.MapKeys()

	// 简单的字符串 Key 排序
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	first := true
	for _, key := range keys {
		if !first {
			buf.WriteByte(',')
		}
		first = false

		// Map Key 必须是字符串 (JSON 标准)
		// 这里简单处理，如果 Key 不是 string，用 fmt.Sprint 转
		if key.Kind() == reflect.String {
			safeKey, _ := json.Marshal(key.String())
			buf.Write(safeKey)
		} else {
			buf.WriteByte('"')
			buf.WriteString(fmt.Sprint(key.Interface()))
			buf.WriteByte('"')
		}

		buf.WriteByte(':')

		if err := encodeValue(buf, v.MapIndex(key)); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

func encodeSlice(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := encodeValue(buf, v.Index(i)); err != nil {
			return err
		}
	}
	buf.WriteByte(']')
	return nil
}
