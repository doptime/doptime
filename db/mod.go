package db

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

// GenerateNanoid creates a unique identifier using the specified size.
func GenerateNanoid(size int) string {
	alphabet := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	if size <= 0 || size > 21 {
		size = 21
	}
	id, b := make([]byte, size), make([]byte, 1)

	for i := range id {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err == nil {
			id[i] = alphabet[index.Int64()]
		} else {
			rand.Read(b)
			ind := int(b[0]) % len(alphabet)
			id[i] = alphabet[ind]
		}
	}
	return string(id)
}

// ModifierFunc is the function signature for all field modifiers.
type ModifierFunc func(fieldValue interface{}, tagParam string) (interface{}, error)

// FieldModifier stores metadata for a struct field's modifier.
type FieldModifier struct {
	FieldIndex int
	FieldName  string
	Modifier   ModifierFunc
	TagParam   string
	ForceApply bool
}

// StructModifiers holds a collection of registered modifiers for a specific struct type and cached tag info.
type StructModifiers struct {
	modifierRegistry map[string]ModifierFunc
	fieldModifiers   []*FieldModifier
	ValType          reflect.Type
}

// TrimSpaces removes leading and trailing white spaces from the string.
func TrimSpaces(fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.TrimSpace(str), nil
	}
	return fieldValue, nil
}

// ToLowercase converts the string to lowercase.
func ToLowercase(fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.ToLower(str), nil
	}
	return fieldValue, nil
}

// ToUppercase converts the string to uppercase.
func ToUppercase(fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.ToUpper(str), nil
	}
	return fieldValue, nil
}

// ToTitleCase converts the string to title case.
func ToTitleCase(fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.Title(strings.ToLower(str)), nil
	}
	return fieldValue, nil
}

// FormatDate formats a time.Time value according to the provided format.
func FormatDate(fieldValue interface{}, tagParam string) (interface{}, error) {
	if t, ok := fieldValue.(time.Time); ok {
		return t.Format(tagParam), nil
	}
	return fieldValue, nil
}
func ApplyCounter(fieldValue interface{}, tagParam string) (interface{}, error) {
	if v, ok := fieldValue.(int); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(int8); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(int16); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(int32); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(int64); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(uint); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(uint8); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(uint16); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(uint32); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(uint64); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(float32); ok {
		return v + 1, nil
	} else if v, ok := fieldValue.(float64); ok {
		return v + 1, nil
	}
	return fieldValue, nil
}

var ModerMap = cmap.New[*StructModifiers]()

// RegisterStructModifiers initializes the StructModifiers for a specific struct type with optional extra modifiers.
func RegisterStructModifiers(extraModifiers map[string]ModifierFunc, structType reflect.Type) bool {
	if structType == nil {
		return false
	}
	for k := structType.Kind(); k == reflect.Pointer; k = structType.Kind() {
		structType = structType.Elem()
	}
	// Ensure we have a struct type.
	if structType.Kind() != reflect.Struct {
		return false
	}

	modifiers := &StructModifiers{
		modifierRegistry: map[string]ModifierFunc{
			"default":    ApplyDefault,
			"unixtime":   ApplyUnixTime,
			"counter":    ApplyCounter,
			"nanoid":     GenerateNanoidFunc,
			"trim":       TrimSpaces,
			"lowercase":  ToLowercase,
			"uppercase":  ToUppercase,
			"title":      ToTitleCase,
			"dateFormat": FormatDate,
		},
		fieldModifiers: []*FieldModifier{},
		ValType:        structType,
	}
	for name, modifier := range extraModifiers {
		modifiers.modifierRegistry[name] = modifier
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("mod")
		if tag != "" {
			forceApply := false
			if forceApply = strings.Contains(tag, ",force"); forceApply {
				tag = strings.Replace(tag, ",force", "", -1)
			}
			cmd_Param := strings.SplitN(tag, "=", 2)
			modifierName := cmd_Param[0]
			tagParam := ""
			if len(cmd_Param) == 2 {
				tagParam = cmd_Param[1]
			}
			modifierFunc, exists := modifiers.modifierRegistry[modifierName]
			if !exists {
				continue // Skip unregistered modifiers
			}

			fieldModifier := &FieldModifier{
				FieldIndex: i,
				FieldName:  field.Name,
				Modifier:   modifierFunc,
				TagParam:   tagParam,
				ForceApply: forceApply,
			}
			modifiers.fieldModifiers = append(modifiers.fieldModifiers, fieldModifier)
		}
	}
	if len(modifiers.fieldModifiers) == 0 {
		return false
	}

	_typeName := structType.String()
	ModerMap.Set(_typeName, modifiers)
	return true
}

func ApplyModifiers(val interface{}) error {
	// Check if the provided value is nil
	if val == nil {
		return fmt.Errorf("nil value passed to ApplyModifiers")
	}

	// Load the modifiers
	structValue := reflect.ValueOf(val)
	structType := structValue.Type()
	// If it's a pointer, dereference it until we get the underlying non-pointer type
	for structType.Kind() == reflect.Pointer || structType.Kind() == reflect.Interface {
		structValue = structValue.Elem()
		structType = structValue.Type()
	}
	_typeName := structType.String()
	modifiers, ok := ModerMap.Get(_typeName)
	if !ok || modifiers == nil {
		return nil // If no modifiers are found, simply return
	}
	// Ensure the final type is a struct
	if structType != modifiers.ValType {
		return fmt.Errorf("ApplyModifiers expects a struct type but got %v", structType.Kind())
	}

	// Apply the modifiers to each field
	for _, fieldModifier := range modifiers.fieldModifiers {
		field := structValue.Field(fieldModifier.FieldIndex)

		// Apply the modifier only if the field can be set
		if fieldModifier.ForceApply || isZero(field) {
			newValue, err := fieldModifier.Modifier(field.Interface(), fieldModifier.TagParam)
			if err != nil {
				return err
			}

			// Set the new value back to the struct field
			if field.CanSet() {
				// Handle type conversion, in case the new value's type doesn't match the field's type
				newValueReflect := reflect.ValueOf(newValue)
				if newValueReflect.Type().ConvertibleTo(field.Type()) {
					field.Set(newValueReflect.Convert(field.Type()))
				} else {
					return fmt.Errorf("cannot set field %s: incompatible types", fieldModifier.FieldName)
				}
			} else {
				return fmt.Errorf("cannot set field %s: field is unexported or otherwise not settable", fieldModifier.FieldName)
			}
		}
	}
	return nil
}

// ApplyDefault sets a default value if the current value is nil or the zero value for its type.
func ApplyDefault(fieldValue interface{}, tagParam string) (interface{}, error) {
	return tagParam, nil
}

// ApplyUnixTime sets the value to the current Unix timestamp based on provided unit.
func ApplyUnixTime(fieldValue interface{}, tagParam string) (interface{}, error) {
	switch tagParam {
	case "ms":
		return time.Now().UnixMilli(), nil
	default:
		return time.Now().Unix(), nil
	}
}

// GenerateNanoidFunc generates a Nanoid and returns it as a string.
func GenerateNanoidFunc(fieldValue interface{}, tagParam string) (interface{}, error) {
	size := 21
	if tagParam != "" {
		size, _ = strconv.Atoi(tagParam)
	}
	return GenerateNanoid(size), nil
}

// isZero checks if a reflect.Value is zero for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr, reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// Example usage
type ExampleStruct struct {
	Name     string    `mod:"trim,lowercase"`
	Age      int       `mod:"default=18"`
	UnixTime int64     `mod:"unixtime=ms,force"`
	Counter  int64     `mod:"counter,force"`
	Email    string    `mod:"lowercase,trim"`
	Created  time.Time `mod:"dateFormat=2006-01-02T15:04:05Z07:00"`
}
