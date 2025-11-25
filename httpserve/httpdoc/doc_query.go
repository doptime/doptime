package httpdoc

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/doptime/redisdb"
)

// GetApiDocs 生成基于 createApi<TIn, TOut> 的 TypeScript 代码
func GetApiDocs() (string, error) {
	result, err := KeyApiDataDocs.HGetAll()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	// 1. 头部引入 createApi
	sb.WriteString("import createApi from \"./api\";\n\n")

	var now = time.Now().Unix()

	// 为了保证生成文件内容的稳定性，建议对 map 的 key 进行排序遍历
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := result[k]

		// 过滤规则：20分钟未更新则忽略
		if v.UpdateAt < now-20*60 {
			continue
		}

		// 处理 API 名称
		apiName := strings.ReplaceAll(v.KeyName, "api:", "")
		if len(apiName) < 1 {
			continue
		}
		// 首字母大写，用于命名 Interface
		apiNamePascal := strings.ToUpper(apiName[0:1]) + apiName[1:]

		// 定义接口名称
		interfaceInName := apiNamePascal + "In"
		interfaceOutName := apiNamePascal + "Out"

		// 2. 生成 Input Interface 定义
		// 假设 v.ParamIn 存储的是具体的数据结构或 map[string]interface{}
		tsInterfaceIn := generateTSInterface(interfaceInName, v.ParamIn)
		sb.WriteString(tsInterfaceIn + "\n")

		// 3. 生成 Output Interface 定义
		tsInterfaceOut := generateTSInterface(interfaceOutName, v.ParamOut)
		sb.WriteString(tsInterfaceOut + "\n")

		// 4. 生成 createApi 调用代码
		// export const apiGetInfo = createApi<GetInfoIn, GetInfoOut>("getInfo");
		sb.WriteString(fmt.Sprintf("export const api%s = createApi<%s, %s>(\"%s\");\n",
			apiNamePascal,
			interfaceInName,
			interfaceOutName,
			apiName,
		))

		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// GetDataDocs 保持原有逻辑，略微调整以匹配整体风格
func GetDataDocs() (string, error) {
	result, err := redisdb.KeyWebDataSchema.HGetAll()
	if err != nil {
		return "", err
	}
	var ret strings.Builder
	ret.WriteString("import { hashKey, stringKey, listKey, setKey, zsetKey, streamKey } from \"doptime-client\"\n\n")
	var now = time.Now().Unix()

	for k, v := range result {
		if v.UpdateAt < now-20*60 {
			continue
		}
		if len(v.KeyName) < 1 {
			continue
		}

		keyWithFirstCharUpper := strings.ToUpper(v.KeyName[0:1]) + v.KeyName[1:]
		keyWithFirstCharUpper = strings.Split(keyWithFirstCharUpper, ":")[0]

		ret.WriteString(v.TSInterface + "\n")

		// 使用 switch case 稍微整洁一点
		var classType string
		switch v.KeyType {
		case "hash":
			classType = "hashKey"
		case "string":
			classType = "stringKey"
		case "list":
			classType = "listKey"
		case "set":
			classType = "setKey"
		case "zset":
			classType = "zsetKey"
		case "stream":
			classType = "streamKey"
		default:
			continue
		}

		ret.WriteString(fmt.Sprintf("export const key%s = new %s<%s>(\"%s\");\n\n",
			keyWithFirstCharUpper, classType, v.ValueTypeName, k))
	}
	return ret.String(), nil
}

// --- 辅助工具函数 ---

// generateTSInterface 通过反射将 Go 的结构/Map 值转换为 TypeScript interface 字符串
func generateTSInterface(interfaceName string, val interface{}) string {
	if val == nil {
		return fmt.Sprintf("export type %s = any;", interfaceName)
	}

	t := reflect.TypeOf(val)
	v := reflect.ValueOf(val)

	// 如果是指针，获取其指向的元素
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("export interface %s {\n", interfaceName))

	// 处理 Map (通常 JSON 反序列化后是 map[string]interface{})
	if t.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			val := iter.Value().Interface()
			tsType := getTSType(val)
			sb.WriteString(fmt.Sprintf("    %s: %s;\n", key, tsType))
		}
	} else if t.Kind() == reflect.Struct {
		// 处理 Struct
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// 忽略未导出的字段
			if field.PkgPath != "" {
				continue
			}

			jsonTag := field.Tag.Get("json")
			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] == "-" {
					continue
				}
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			// 获取字段值的类型
			fieldVal := v.Field(i).Interface()
			tsType := getTSType(fieldVal)
			sb.WriteString(fmt.Sprintf("    %s: %s;\n", fieldName, tsType))
		}
	} else {
		// 如果既不是 Map 也不是 Struct，可能是基本类型，直接用 type 别名
		return fmt.Sprintf("export type %s = %s;", interfaceName, getTSType(val))
	}

	sb.WriteString("}")
	return sb.String()
}

// getTSType 简单的类型映射辅助函数
func getTSType(val interface{}) string {
	if val == nil {
		return "any"
	}
	t := reflect.TypeOf(val)
	k := t.Kind()

	switch k {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Slice, reflect.Array:
		// 处理数组/切片
		v := reflect.ValueOf(val)
		if v.Len() > 0 {
			// 取第一个元素推断类型 (简单处理)
			elemType := getTSType(v.Index(0).Interface())
			return elemType + "[]"
		}
		return "any[]"
	case reflect.Map:
		// 嵌套对象简单处理为 any，如果需要深度递归，可以在这里递归调用 generateTSInterface 的逻辑
		// 但为了生成的简洁性，如果不生成嵌套 interface，用 Record 或 object 替代
		return "Record<string, any>"
	case reflect.Struct:
		return "Record<string, any>"
	default:
		return "any"
	}
}
