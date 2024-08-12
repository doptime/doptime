package rdsdb

import (
	"context"
	"crypto/rand"
	"math/big"
	"reflect"
	"strings"
	"time"
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
type ModifierFunc func(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error)

// FieldModifier stores metadata for a struct field's modifier.
type FieldModifier struct {
	FieldIndex int
	FieldName  string
	Modifier   ModifierFunc
	TagParam   string
	ForceApply bool
}

// StructModifiers holds a collection of registered modifiers for a specific struct type and cached tag info.
type StructModifiers[T any] struct {
	modifierRegistry map[string]ModifierFunc
	fieldModifiers   []*FieldModifier
}

// TrimSpaces removes leading and trailing white spaces from the string.
func TrimSpaces(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.TrimSpace(str), nil
	}
	return fieldValue, nil
}

// ToLowercase converts the string to lowercase.
func ToLowercase(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.ToLower(str), nil
	}
	return fieldValue, nil
}

// ToUppercase converts the string to uppercase.
func ToUppercase(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.ToUpper(str), nil
	}
	return fieldValue, nil
}

// ToTitleCase converts the string to title case.
func ToTitleCase(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	if str, ok := fieldValue.(string); ok {
		return strings.Title(strings.ToLower(str)), nil
	}
	return fieldValue, nil
}

// FormatDate formats a time.Time value according to the provided format.
func FormatDate(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	if t, ok := fieldValue.(time.Time); ok {
		return t.Format(tagParam), nil
	}
	return fieldValue, nil
}

// RegisterStructModifiers initializes the StructModifiers for a specific struct type with optional extra modifiers.
func RegisterStructModifiers[T any](extraModifiers map[string]ModifierFunc) *StructModifiers[T] {
	structType := reflect.TypeOf((*T)(nil)).Elem()
	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	if kv := structType.Kind().String(); kv != "struct" {
		return nil
	}

	modifiers := &StructModifiers[T]{
		modifierRegistry: map[string]ModifierFunc{
			"default":    ApplyDefault,
			"unixtime":   ApplyUnixTime,
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

	return modifiers
}

// ApplyModifiers applies the registered modifiers to an instance of the struct.
func (m *StructModifiers[T]) ApplyModifiers(ctx context.Context, s *T) error {
	structValue := reflect.ValueOf(s).Elem()

	for _, fieldModifier := range m.fieldModifiers {
		field := structValue.Field(fieldModifier.FieldIndex)
		if fieldModifier.ForceApply || isZero(field) {
			newValue, err := fieldModifier.Modifier(ctx, field.Interface(), fieldModifier.TagParam)
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
func ApplyDefault(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	return tagParam, nil
}

// ApplyUnixTime sets the value to the current Unix timestamp based on provided unit.
func ApplyUnixTime(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	switch tagParam {
	case "ms":
		return time.Now().UnixMilli(), nil
	case "s":
		return time.Now().Unix(), nil
	default:
		return time.Now().UnixMilli(), nil
	}
}

// GenerateNanoidFunc generates a Nanoid and returns it as a string.
func GenerateNanoidFunc(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	return GenerateNanoid(21), nil
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
	Email    string    `mod:"lowercase,trim"`
	Created  time.Time `mod:"dateFormat=2006-01-02T15:04:05Z07:00"`
}
