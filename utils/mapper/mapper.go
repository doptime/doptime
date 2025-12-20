package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Decode 对外入口
func Decode(input interface{}, output interface{}) error {
	config := &DecoderConfig{
		Result:           output,
		TagName:          "json",
		WeaklyTypedInput: true,
	}
	decoder, err := NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

type DecoderConfig struct {
	Result           interface{}
	TagName          string
	WeaklyTypedInput bool
}

func NewDecoder(config *DecoderConfig) (*Decoder, error) {
	val := reflect.ValueOf(config.Result)
	if val.Kind() != reflect.Ptr {
		return nil, errors.New("result must be a pointer")
	}
	val = val.Elem()
	if !val.CanAddr() {
		return nil, errors.New("result must be addressable")
	}
	if config.TagName == "" {
		config.TagName = "json"
	}
	return &Decoder{config: config}, nil
}

type Decoder struct {
	config *DecoderConfig
}

func (d *Decoder) Decode(input interface{}) error {
	return d.decode("root", input, reflect.ValueOf(d.config.Result).Elem())
}

func (d *Decoder) decode(name string, input interface{}, outVal reflect.Value) error {
	var inputVal reflect.Value
	if input != nil {
		inputVal = reflect.ValueOf(input)
		if inputVal.Kind() == reflect.Ptr && inputVal.IsNil() {
			input = nil
		}
	}

	if input == nil {
		if outVal.Kind() == reflect.Ptr {
			if !outVal.IsNil() && outVal.CanSet() {
				outVal.Set(reflect.Zero(outVal.Type()))
			}
		}
		return nil
	}

	if !inputVal.IsValid() {
		return nil
	}

	// 弱类型 Hook (增强：支持 json.Number)
	if d.config.WeaklyTypedInput {
		input = d.weaklyTypedHook(input, outVal.Type())
	}

	switch outVal.Kind() {
	case reflect.Bool:
		return d.decodeBool(input, outVal)
	case reflect.Interface:
		return d.decodeBasic(input, outVal)
	case reflect.String:
		return d.decodeString(input, outVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.decodeInt(input, outVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return d.decodeUint(input, outVal)
	case reflect.Float32, reflect.Float64:
		return d.decodeFloat(input, outVal)
	case reflect.Struct:
		return d.decodeStruct(name, input, outVal)
	case reflect.Map:
		return d.decodeMap(name, input, outVal)
	case reflect.Slice, reflect.Array:
		return d.decodeSlice(name, input, outVal)
	case reflect.Ptr:
		return d.decodePtr(name, input, outVal)
	default:
		return nil
	}
}

// FIX #3: 增加 AssignableTo 检查，防止 interface{} 赋值 Panic
func (d *Decoder) decodeBasic(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	if !dataVal.Type().AssignableTo(val.Type()) {
		// 如果不能直接赋值，尝试做一次暴力转换（针对自定义类型）
		if dataVal.Type().ConvertibleTo(val.Type()) {
			val.Set(dataVal.Convert(val.Type()))
			return nil
		}
		// 如果完全无法赋值，忽略而不是 Panic
		return nil
	}
	val.Set(dataVal)
	return nil
}

func (d *Decoder) decodeString(data interface{}, val reflect.Value) error {
	val.SetString(fmt.Sprintf("%v", data))
	return nil
}

func (d *Decoder) decodeInt(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	switch dataVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetInt(dataVal.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val.SetInt(int64(dataVal.Uint()))
	case reflect.Float32, reflect.Float64:
		val.SetInt(int64(dataVal.Float()))
	case reflect.String:
		s := dataVal.String()
		if idx := strings.Index(s, "."); idx != -1 {
			s = s[:idx]
		}
		i, err := strconv.ParseInt(s, 0, val.Type().Bits())
		if err == nil {
			val.SetInt(i)
		}
	}
	return nil
}

func (d *Decoder) decodeUint(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	switch dataVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetUint(uint64(dataVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val.SetUint(dataVal.Uint())
	case reflect.Float32, reflect.Float64:
		val.SetUint(uint64(dataVal.Float()))
	case reflect.String:
		s := dataVal.String()
		if idx := strings.Index(s, "."); idx != -1 {
			s = s[:idx]
		}
		i, err := strconv.ParseUint(s, 0, val.Type().Bits())
		if err == nil {
			val.SetUint(i)
		}
	}
	return nil
}

func (d *Decoder) decodeFloat(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	switch dataVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetFloat(float64(dataVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val.SetFloat(float64(dataVal.Uint()))
	case reflect.Float32, reflect.Float64:
		val.SetFloat(dataVal.Float())
	case reflect.String:
		f, err := strconv.ParseFloat(dataVal.String(), val.Type().Bits())
		if err == nil {
			val.SetFloat(f)
		}
	}
	return nil
}

func (d *Decoder) decodeBool(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	switch dataVal.Kind() {
	case reflect.Bool:
		val.SetBool(dataVal.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetBool(dataVal.Int() != 0)
	case reflect.String:
		b, err := strconv.ParseBool(dataVal.String())
		if err == nil {
			val.SetBool(b)
		}
	}
	return nil
}

// decodeStruct V4: 支持 remain 功能
// decodeStruct V5: 工业级健壮版 (修复 Panic 隐患)
func (d *Decoder) decodeStruct(name string, input interface{}, outVal reflect.Value) error {
	// 防御性处理 Input，确保拿到实体
	var dataVal reflect.Value
	if input != nil {
		dataVal = reflect.ValueOf(input)
		if dataVal.Kind() == reflect.Ptr {
			if dataVal.IsNil() {
				return nil
			}
			dataVal = dataVal.Elem()
		}
	} else {
		return nil
	}

	// 1. 如果输入是 Struct，尝试直接赋值
	if dataVal.Kind() == reflect.Struct {
		if dataVal.Type().AssignableTo(outVal.Type()) {
			outVal.Set(dataVal)
			return nil
		}
		return nil
	}
	// 2. 必须是 Map
	if dataVal.Kind() != reflect.Map {
		return nil
	}

	// 准备 Map Keys (ToLower) 用于匹配
	dataMapLower := make(map[string]reflect.Value, dataVal.Len())
	// 追踪哪些 Key 被用掉了
	usedKeys := make(map[string]bool)

	iter := dataVal.MapRange()
	for iter.Next() {
		k := iter.Key()
		// 只处理 string 类型的 key，防止奇怪的 map[int]any 导致 panic
		if k.Kind() == reflect.String {
			dataMapLower[strings.ToLower(k.String())] = iter.Value()
		}
	}

	outType := outVal.Type()
	var remainFieldVal reflect.Value

	for i := 0; i < outType.NumField(); i++ {
		field := outType.Field(i)
		fieldVal := outVal.Field(i)

		if field.PkgPath != "" { // 忽略未导出字段
			continue
		}

		// --- Tag 解析 ---
		tagParts := strings.Split(field.Tag.Get(d.config.TagName), ",")
		tagName := tagParts[0]

		isRemain := false
		for _, opt := range tagParts[1:] {
			if opt == "remain" {
				isRemain = true
				break
			}
		}

		if isRemain {
			remainFieldVal = fieldVal
			continue
		}

		if tagName == "-" {
			continue
		}
		if tagName == "" {
			tagName = field.Name
		}

		// --- 匹配逻辑 ---
		lowerTagName := strings.ToLower(tagName)
		val, ok := dataMapLower[lowerTagName]

		// 嵌入结构体 (Anonymous) 处理
		if !ok && field.Anonymous && fieldVal.Kind() == reflect.Struct {
			if err := d.decode(name, input, fieldVal); err != nil {
				return err
			}
			continue
		}

		if ok {
			usedKeys[lowerTagName] = true
			if err := d.decode(name+"."+field.Name, val.Interface(), fieldVal); err != nil {
				return err
			}
		}
	}

	// --- Remain 处理逻辑 (V5 修复核心) ---
	if remainFieldVal.IsValid() && remainFieldVal.Kind() == reflect.Map {
		// FIX Bug 1: 如果 map 是 nil 才 Make，否则复用，防止覆盖用户初始化的数据
		if remainFieldVal.IsNil() {
			remainFieldVal.Set(reflect.MakeMap(remainFieldVal.Type()))
		}

		remainKeyType := remainFieldVal.Type().Key()
		remainElemType := remainFieldVal.Type().Elem()

		iter := dataVal.MapRange()
		for iter.Next() {
			k := iter.Key()
			// 必须是 string key 才能判断是否 used
			if k.Kind() == reflect.String {
				if !usedKeys[strings.ToLower(k.String())] {
					// FIX Bug 2: 类型安全检查
					// 确保 Key 可以赋值给 remain map 的 key (通常都是 string，但以防万一)
					if !k.Type().AssignableTo(remainKeyType) {
						continue
					}

					// 尝试解码 Value
					newVal := reflect.New(remainElemType).Elem()
					if err := d.decode("remainVal", iter.Value().Interface(), newVal); err == nil {
						// 再次检查 Value 是否可赋值 (decode 内部虽然处理了，但多一层保险)
						remainFieldVal.SetMapIndex(k, newVal)
					}
				}
			}
		}
	}

	return nil
}

func (d *Decoder) decodeMap(name string, input interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(input))
	if dataVal.Kind() != reflect.Map {
		return nil
	}

	valType := val.Type()
	keyType := valType.Key()
	elemType := valType.Elem()

	if val.IsNil() {
		val.Set(reflect.MakeMap(valType))
	}

	iter := dataVal.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		currentKey := reflect.New(keyType).Elem()
		if err := d.decode(name+"[key]", k.Interface(), currentKey); err != nil {
			continue
		}

		currentVal := reflect.New(elemType).Elem()
		if err := d.decode(name+"["+k.String()+"]", v.Interface(), currentVal); err != nil {
			return err
		}

		val.SetMapIndex(currentKey, currentVal)
	}
	return nil
}

func (d *Decoder) decodeSlice(name string, input interface{}, val reflect.Value) error {
	dataVal := reflect.Indirect(reflect.ValueOf(input))
	if dataVal.Kind() != reflect.Slice && dataVal.Kind() != reflect.Array {
		return nil
	}

	valType := val.Type()

	slice := reflect.MakeSlice(valType, dataVal.Len(), dataVal.Len())

	for i := 0; i < dataVal.Len(); i++ {
		currentData := dataVal.Index(i).Interface()
		currentField := slice.Index(i)
		if err := d.decode(fmt.Sprintf("%s[%d]", name, i), currentData, currentField); err != nil {
			return err
		}
	}
	val.Set(slice)
	return nil
}

func (d *Decoder) decodePtr(name string, input interface{}, val reflect.Value) error {
	if input == nil {
		val.Set(reflect.Zero(val.Type()))
		return nil
	}
	if val.IsNil() {
		val.Set(reflect.New(val.Type().Elem()))
	}
	return d.decode(name, input, val.Elem())
}

func (d *Decoder) weaklyTypedHook(data interface{}, targetType reflect.Type) interface{} {
	dataVal := reflect.ValueOf(data)

	// FIX #6: 优先处理 json.Number (必须在其他 convert 之前)
	if _, ok := data.(json.Number); ok {
		return d.decodeJsonNumber(data.(json.Number), targetType)
	}

	// String -> Time
	if targetType == reflect.TypeOf(time.Time{}) && dataVal.Kind() == reflect.String {
		str := dataVal.String()
		formats := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02", "2006/01/02"}
		for _, f := range formats {
			if t, err := time.ParseInLocation(f, str, time.Local); err == nil {
				return t
			}
		}
	}

	// Map Key 兼容
	if targetType.Kind() == reflect.Map && targetType.Key().Kind() == reflect.String {
		if dataVal.Kind() == reflect.Map && dataVal.Type().Key().Kind() == reflect.Interface {
			m := make(map[string]interface{})
			iter := dataVal.MapRange()
			for iter.Next() {
				m[fmt.Sprintf("%v", iter.Key())] = iter.Value().Interface()
			}
			return m
		}
	}

	return data
}

// FIX #6: 专门处理 json.Number
func (d *Decoder) decodeJsonNumber(jn json.Number, targetType reflect.Type) interface{} {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := jn.Int64(); err == nil {
			return i
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, err := strconv.ParseUint(string(jn), 10, 64); err == nil {
			return i
		}
	case reflect.Float32, reflect.Float64:
		if f, err := jn.Float64(); err == nil {
			return f
		}
	case reflect.String:
		return string(jn)
	}
	// 默认返回 string 以防万一
	return string(jn)
}
