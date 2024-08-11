package rdsdb

import (
	"context"
	"crypto/rand"
	"math/big"
	"reflect"
	"strings"
	"time"
)

func Nanoid(size int) string {
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

// UnixTime sets the value to the current Unix timestamp based on provided unit.
var nanoidFunc = func(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	return Nanoid(21), nil
}

// Modifier is the function signature for all modifiers.
type Modifier func(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error)

// TagCache stores metadata for a struct field's modifier.
type TagCache struct {
	Index      int
	FieldName  string
	ModFunc    Modifier
	TagParam   string
	SkipIsZero bool
}

// Modifiers holds a collection of registered modifiers for a specific struct type and cached tag info.
type Modifiers[T any] struct {
	registry map[string]Modifier
	tagCache []*TagCache
}

// NewModifiers creates a new instance of Modifiers for a specific struct type.
func NewModifiers[T any](modMap map[string]Modifier) *Modifiers[T] {
	modifiers := &Modifiers[T]{
		registry: map[string]Modifier{
			"default":  Default,
			"unixtime": unixtime,
			"nanoid":   nanoidFunc,
		},
		tagCache: []*TagCache{},
	}

	t := reflect.TypeOf((*T)(nil)).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mod")
		if tag != "" {
			keyValue := strings.SplitN(tag, "=", 2)
			modName := keyValue[0]
			tagParam := ""
			if len(keyValue) == 2 {
				tagParam = keyValue[1]
			}

			skipIsZero := false
			if strings.Contains(tagParam, "force") {
				skipIsZero = true
				tagParam = strings.Replace(tagParam, "force", "", -1)
				tagParam = strings.Trim(tagParam, ",")
			}

			modFunc, exists := modifiers.registry[modName]
			if !exists {
				continue // Skip unregistered modifiers
			}

			modifiers.tagCache = append(modifiers.tagCache, &TagCache{
				Index:      i,
				FieldName:  field.Name,
				ModFunc:    modFunc,
				TagParam:   tagParam,
				SkipIsZero: skipIsZero,
			})
		}
	}
	if len(modifiers.tagCache) == 0 {
		return nil
	}

	return modifiers
}

// Mod applies the registered modifiers to an instance of the struct.
func (m *Modifiers[T]) Mod(ctx context.Context, s *T) error {
	if m == nil {
		return nil
	}

	v := reflect.ValueOf(s).Elem()

	for _, cache := range m.tagCache {
		field := v.Field(cache.Index)
		if cache.SkipIsZero || isZero(field) {
			newValue, err := cache.ModFunc(ctx, field.Interface(), cache.TagParam)
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

// Default sets a default value if the current value is nil or the zero value for its type.
func Default(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	return tagParam, nil
}

// UnixTime sets the value to the current Unix timestamp based on provided unit.
func unixtime(ctx context.Context, fieldValue interface{}, tagParam string) (interface{}, error) {
	switch tagParam {
	case "ms":
		return time.Now().UnixMilli(), nil
	case "s":
		return time.Now().Unix(), nil
	default:
		return time.Now().UnixMilli(), nil
	}
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
	Name     string `mod:"trim"`
	Age      int    `mod:"default=18"`
	UnixTime int64  `mod:"unixtime=ms,force"`
}
