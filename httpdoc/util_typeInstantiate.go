package httpdoc

import (
	"fmt"
	"reflect"
)

func InstantiateType(vType reflect.Type) (value interface{}, err error) {

	// 检查 vType 是否可以实例化
	if vType.Kind() == reflect.Interface || vType.Kind() == reflect.Invalid {
		fmt.Println("vType is not valid, vType: ", vType)
		return nil, fmt.Errorf("vType is not valid in instantiatetype, vType: %v", vType)
	}

	// 创建 v 的实例
	valueElem := reflect.New(vType).Elem()
	//if vType is pointer, we need to create a new instance of the valueElem
	if vType.Kind() == reflect.Ptr {
		//ensure !reflect.ValueOf(value).IsNil()
		valueElem.Set(reflect.New(vType.Elem()))
	}

	InstantiateFields(valueElem)
	return valueElem.Interface(), nil
}
func InstantiateFields(value reflect.Value) {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		value = value.Elem()
	}

	if value.Kind() == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldType := field.Type()

			if field.Kind() == reflect.Ptr && field.IsNil() {
				field.Set(reflect.New(fieldType.Elem()))
			}

			// 其它的类型
			if field.Kind() == reflect.Map && field.IsNil() {
				field.Set(reflect.MakeMap(fieldType))
				// 如果map的key是string类型，初始化一个具体的值
				if fieldType.Key().Kind() == reflect.String {
					elemType := fieldType.Elem()
					var elemValue reflect.Value
					switch elemType.Kind() {
					case reflect.String:
						elemValue = reflect.ValueOf("string")
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						elemValue = reflect.ValueOf(0)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						elemValue = reflect.ValueOf(0)
					case reflect.Float32, reflect.Float64:
						elemValue = reflect.ValueOf(0.0)
					case reflect.Bool:
						elemValue = reflect.ValueOf(false)
					case reflect.Ptr:
						elemValue = reflect.New(elemType.Elem())
					case reflect.Struct:
						elemValue = reflect.New(elemType).Elem()
						InstantiateFields(elemValue)
					default:
						elemValue = reflect.Zero(elemType)
					}
					field.SetMapIndex(reflect.ValueOf("exampleKey"), elemValue)
				}
			}

			// 检查并初始化切片类型字段
			if field.Kind() == reflect.Slice && field.IsNil() {
				elemType := fieldType.Elem()
				switch elemType.Kind() {
				case reflect.String:
					field.Set(reflect.MakeSlice(fieldType, 1, 1))
					field.Index(0).Set(reflect.ValueOf(""))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					field.Set(reflect.MakeSlice(fieldType, 1, 1))
					field.Index(0).Set(reflect.ValueOf(0))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					field.Set(reflect.MakeSlice(fieldType, 1, 1))
					field.Index(0).Set(reflect.ValueOf(0))
				case reflect.Float32, reflect.Float64:
					field.Set(reflect.MakeSlice(fieldType, 1, 1))
					field.Index(0).Set(reflect.ValueOf(0.0))
				case reflect.Bool:
					field.Set(reflect.MakeSlice(fieldType, 1, 1))
					field.Index(0).Set(reflect.ValueOf(false))
				default:
					field.Set(reflect.MakeSlice(fieldType, 0, 0))
				}
			}

			if (field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr) && !field.IsNil() {
				InstantiateFields(field)
			}
		}
	}
}
