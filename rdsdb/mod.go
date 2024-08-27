package rdsdb

import (
	"crypto/rand"
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
func RegisterStructModifiers(extraModifiers map[string]ModifierFunc, structType reflect.Type) *StructModifiers {
	//structType := reflect.TypeOf((*T)(nil)).Elem()
	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if kv := structType.Kind().String(); kv != "struct" {
		return nil
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
	}
	for name, modifier := range extraModifiers {
		modifiers.modifierRegistry[name] = modifier
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tag := field.Tag.Get("mod")
		if tag != "" {
			tagParts := strings.SplitN(tag, "=", 2)
			modifierName := tagParts[0]
			tagParam := ""
			if len(tagParts) == 2 {
				tagParam = tagParts[1]
			}
			modifierFunc, exists := modifiers.modifierRegistry[modifierName]
			if !exists {
				continue // Skip unregistered modifiers
			}

			forceApply := false
			if strings.Contains(tagParam, "force") {
				forceApply = true
				tagParam = strings.Replace(tagParam, "force", "", -1)
				tagParam = strings.Trim(tagParam, ",")
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
		return nil
	}
	_typeName := structType.String()
	ModerMap.Set(_typeName, modifiers)

	return modifiers
}

var nonModifiers *StructModifiers = &StructModifiers{modifierRegistry: nil, fieldModifiers: nil}

func ApplyModifiers(modifiers *StructModifiers, val interface{}) error {
	var ok bool

	// load modifier
	if modifiers == nonModifiers {
		return nil
	} else {
		structType := reflect.TypeOf(val)
		if structType.Kind() == reflect.Pointer {
			structType = structType.Elem()
		}
		_typeName := structType.String()
		modifiers, ok = ModerMap.Get(_typeName)
		if !ok {
			*modifiers = *nonModifiers
			return nil
		}
	}

	structValue := reflect.ValueOf(val)
	for structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	for _, fieldModifier := range modifiers.fieldModifiers {
		field := structValue.Field(fieldModifier.FieldIndex)
		if fieldModifier.ForceApply || isZero(field) {
			newValue, err := fieldModifier.Modifier(field.Interface(), fieldModifier.TagParam)
			if err != nil {
				return err
			}

			// Setting the new value back to the struct field.
			if field.CanSet() {
				field.Set(reflect.ValueOf(newValue))
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
