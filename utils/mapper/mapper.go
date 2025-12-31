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

// Decode is the external entry point
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
	return &Decoder{
		config: config,
	}, nil
}

type Decoder struct {
	config *DecoderConfig
	// refTracker is used to prevent potential circular dependencies if default values
	// ever support chaining or complex computed logic.
	// Currently used as a defensive mechanism.
	refTracker map[string]bool
}

func (d *Decoder) Decode(input interface{}) error {
	// Reset tracker for each fresh decode call
	d.refTracker = make(map[string]bool)
	return d.decode("root", input, reflect.ValueOf(d.config.Result).Elem())
}

// inputWrapper holds the value and the original key to preserve case sensitivity in Remain
type inputWrapper struct {
	val         reflect.Value
	originalKey string
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
		// If input is nil, we still proceed because defaults might need to be applied
		if outVal.Kind() == reflect.Ptr {
			if !outVal.IsNil() && outVal.CanSet() {
				outVal.Set(reflect.Zero(outVal.Type()))
			}
			return nil
		}
		if outVal.Kind() != reflect.Struct {
			return nil
		}
	}

	// Weakly typed hook
	if d.config.WeaklyTypedInput && input != nil {
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

// ============================================================================
// Core Logic: Struct Decoding
// ============================================================================

func (d *Decoder) decodeStruct(name string, input interface{}, outVal reflect.Value) error {
	// 1. Prepare Input Data Map
	dataMap, err := d.prepareInputMap(input, outVal)
	if err != nil {
		if err.Error() == "assigned directly" {
			return nil
		}
		return nil
	}

	outType := outVal.Type()
	usedKeys := make(map[string]bool)
	var remainFieldVal reflect.Value

	// 2. Iterate over Struct Fields
	for i := 0; i < outType.NumField(); i++ {
		field := outType.Field(i)
		fieldVal := outVal.Field(i)

		if field.PkgPath != "" {
			continue // Skip unexported fields
		}

		// 2.1 Parse Tag
		tagVal := field.Tag.Get(d.config.TagName)
		tagName, defaultParam := d.parseTag(tagVal)

		// 2.2 Identify "Remain" field
		isRemain := (field.Name == "Remain" || tagName == "remain") &&
			fieldVal.Kind() == reflect.Map &&
			fieldVal.Type().Key().Kind() == reflect.String

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

		// 2.3 Resolve Value
		val, keyUsed, exists := d.resolveValue(tagName, defaultParam, dataMap)

		if exists {
			if keyUsed != "" {
				usedKeys[strings.ToLower(keyUsed)] = true
			}

			if err := d.decode(name+"."+field.Name, val, fieldVal); err != nil {
				return err
			}
		} else if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			// Embedded struct: pass original input down
			if err := d.decode(name, input, fieldVal); err != nil {
				return err
			}
		}
	}

	// 3. Fill Remain Field
	if remainFieldVal.IsValid() {
		d.fillRemain(remainFieldVal, dataMap, usedKeys)
	}

	return nil
}

func (d *Decoder) prepareInputMap(input interface{}, outVal reflect.Value) (map[string]inputWrapper, error) {
	var dataVal reflect.Value
	if input != nil {
		dataVal = reflect.ValueOf(input)
		if dataVal.Kind() == reflect.Ptr {
			if dataVal.IsNil() {
				return make(map[string]inputWrapper), nil
			}
			dataVal = dataVal.Elem()
		}
	}

	if !dataVal.IsValid() {
		return make(map[string]inputWrapper), nil
	}

	if dataVal.Kind() == reflect.Struct && dataVal.Type().AssignableTo(outVal.Type()) {
		outVal.Set(dataVal)
		return nil, errors.New("assigned directly")
	}

	if dataVal.Kind() != reflect.Map {
		return make(map[string]inputWrapper), nil
	}

	dataMap := make(map[string]inputWrapper, dataVal.Len())
	iter := dataVal.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		if k.Kind() == reflect.String {
			kStr := k.String()
			dataMap[strings.ToLower(kStr)] = inputWrapper{
				val:         v,
				originalKey: kStr,
			}
		}
	}
	return dataMap, nil
}

// resolveValue determines the value to use (Input -> Ref Default -> Literal Default)
func (d *Decoder) resolveValue(tagName, defaultParam string, dataMap map[string]inputWrapper) (interface{}, string, bool) {
	lowerTagName := strings.ToLower(tagName)

	// 1. Direct lookup
	if wrapper, ok := dataMap[lowerTagName]; ok {
		return wrapper.val.Interface(), lowerTagName, true
	}

	// 2. Check Defaults
	if defaultParam == "" {
		return nil, "", false
	}

	// 2a. Reference Default (@ref)
	if strings.HasPrefix(defaultParam, "@") {
		refKey := strings.ToLower(defaultParam[1:])

		// Cycle detection (Defensive)
		if d.refTracker[refKey] {
			return nil, "", false // Cycle detected or already visited in a chain
		}
		d.refTracker[refKey] = true
		defer func() { delete(d.refTracker, refKey) }()

		if wrapper, ok := dataMap[refKey]; ok {
			return wrapper.val.Interface(), refKey, true
		}
		return nil, "", false
	}

	// 2b. Literal Default
	// Improvement: Try to parse literal into basic types (int, bool, float)
	// This helps when the target field is interface{}
	parsedVal := d.parseDefaultLiteral(defaultParam)
	return parsedVal, "", true
}

// parseDefaultLiteral attempts to convert string defaults into specific types
func (d *Decoder) parseDefaultLiteral(s string) interface{} {
	// Try Bool
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	// Try Int
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i // JSON usually prefers float64, but Int is safer for strict mapping
	}
	// Try Float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Fallback to String
	return s
}

func (d *Decoder) fillRemain(remainField reflect.Value, dataMap map[string]inputWrapper, usedKeys map[string]bool) {
	if remainField.IsNil() {
		remainField.Set(reflect.MakeMap(remainField.Type()))
	}

	remainElemType := remainField.Type().Elem()

	for kLower, wrapper := range dataMap {
		if usedKeys[kLower] {
			continue
		}

		keyVal := reflect.ValueOf(wrapper.originalKey)

		if !keyVal.Type().AssignableTo(remainField.Type().Key()) {
			continue
		}

		newVal := reflect.New(remainElemType).Elem()
		if err := d.decode("remainVal", wrapper.val.Interface(), newVal); err == nil {
			remainField.SetMapIndex(keyVal, newVal)
		}
	}
}

func (d *Decoder) parseTag(tag string) (name string, defaultVal string) {
	if tag == "" {
		return "", ""
	}
	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "default=") {
			defaultVal = strings.TrimPrefix(part, "default=")
			// Note: This simple parser does not support commas inside the default value.
		}
	}
	return name, defaultVal
}

// ============================================================================
// Type Decoding Handlers (Standard)
// ============================================================================

func (d *Decoder) decodeBasic(data interface{}, val reflect.Value) error {
	dataVal := reflect.ValueOf(data)
	if !dataVal.Type().AssignableTo(val.Type()) {
		if dataVal.Type().ConvertibleTo(val.Type()) {
			val.Set(dataVal.Convert(val.Type()))
			return nil
		}
		// Special handling for numeric conversions if types differ (e.g. int -> float)
		// This happens often with default values (parsed as int) assigned to float fields
		return d.decodeNumericBridge(dataVal, val)
	}
	val.Set(dataVal)
	return nil
}

// decodeNumericBridge attempts to bridge int/float/uint mismatches for basic interface decoding
func (d *Decoder) decodeNumericBridge(dataVal reflect.Value, val reflect.Value) error {
	kind := val.Kind()
	switch {
	case isInt(kind) && isFloat(dataVal.Kind()):
		val.SetInt(int64(dataVal.Float()))
		return nil
	case isFloat(kind) && isInt(dataVal.Kind()):
		val.SetFloat(float64(dataVal.Int()))
		return nil
	}
	return nil
}

func isInt(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}
func isFloat(k reflect.Kind) bool {
	return k == reflect.Float32 || k == reflect.Float64
}

func (d *Decoder) decodeString(data interface{}, val reflect.Value) error {
	val.SetString(fmt.Sprintf("%v", data))
	return nil
}

func (d *Decoder) decodeInt(data interface{}, val reflect.Value) error {
	if data == nil {
		return nil
	}
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
		if s == "" {
			return nil
		}
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
	if data == nil {
		return nil
	}
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
		if s == "" {
			return nil
		}
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
	if data == nil {
		return nil
	}
	dataVal := reflect.ValueOf(data)
	switch dataVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val.SetFloat(float64(dataVal.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val.SetFloat(float64(dataVal.Uint()))
	case reflect.Float32, reflect.Float64:
		val.SetFloat(dataVal.Float())
	case reflect.String:
		s := dataVal.String()
		if s == "" {
			return nil
		}
		f, err := strconv.ParseFloat(s, val.Type().Bits())
		if err == nil {
			val.SetFloat(f)
		}
	}
	return nil
}

func (d *Decoder) decodeBool(data interface{}, val reflect.Value) error {
	if data == nil {
		return nil
	}
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

// ============================================================================
// Hooks and Utilities
// ============================================================================

func (d *Decoder) weaklyTypedHook(data interface{}, targetType reflect.Type) interface{} {
	dataVal := reflect.ValueOf(data)

	// JSON Number Support
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

	// Map Key String Compatibility
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
	return string(jn)
}
