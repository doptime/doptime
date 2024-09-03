package db

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var WebDataDocsMap cmap.ConcurrentMap[string, *WebDataDocs] = cmap.New[*WebDataDocs]()

var SynWebDataRunOnce = sync.Mutex{}
var KeyWebDataDocs = HashKey[string, *WebDataDocs](WithKey("Docs:Data"))

func initializeFields(value reflect.Value) {
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
						elemValue = reflect.ValueOf("")
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
						initializeFields(elemValue)
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
				initializeFields(field)
			}
		}
	}
}

type WebDataDocs struct {
	KeyName string
	// string, hash, list, set, zset, stream
	KeyType         string
	UpdateAt        int64
	CreateFromLocal bool `msgpack:"-"`
	Instance        interface{}
}

func (ctx *Ctx[k, v]) RegisterWebData() {
	var validRdsKeyTypes = map[string]bool{"string": true, "list": true, "set": true, "hash": true, "zset": true, "stream": true}
	if _, ok := validRdsKeyTypes[ctx.KeyType]; !ok {
		return
	}
	// 获取 v 的类型
	vType := reflect.TypeOf((*v)(nil)).Elem()

	// 检查 vType 是否可以实例化
	if vType.Kind() == reflect.Interface || vType.Kind() == reflect.Invalid {
		fmt.Println("vType is not valid, vType: ", vType)
		return
	}

	// 创建 v 的实例
	valueElem := reflect.New(vType).Elem()
	//if vType is pointer, we need to create a new instance of the valueElem
	if vType.Kind() == reflect.Ptr {
		valueElem.Set(reflect.New(vType.Elem()))
	}
	value := valueElem.Interface()

	if reflect.ValueOf(value).IsNil() {
		return
	}
	initializeFields(valueElem)
	rootKey := strings.Split(ctx.Key, ":")[0]
	dataSchema := &WebDataDocs{KeyName: rootKey, KeyType: ctx.KeyType, Instance: value, UpdateAt: time.Now().Unix(), CreateFromLocal: true}
	WebDataDocsMap.Set(ctx.Key, dataSchema)
	if SynWebDataRunOnce.TryLock() {
		go syncWebDataToRedis()
	}
}

func syncWebDataToRedis() {
	//wait arrival of other schema to be store in map
	time.Sleep(time.Second)
	for {
		now := time.Now().Unix()
		//only update local defined data to redis
		WebDataDocsMap.IterCb(func(key string, value *WebDataDocs) {
			if value.CreateFromLocal {
				value.UpdateAt = now
			}
		})
		if WebDataDocsMap.Count() > 0 {
			KeyWebDataDocs.HSet(WebDataDocsMap.Items())
		}

		//for the purpose of checking the data schema
		//copy all data schema to local ,but do not cover the local data
		if vals, err := KeyWebDataDocs.HGetAll(); err == nil {
			for k, v := range vals {
				//if defined in non local, allow to cover the definition
				//that is ,only local defined data schema can not be covered
				if v, ok := WebDataDocsMap.Get(k); ok && v.CreateFromLocal {
					continue
				}
				WebDataDocsMap.Set(k, v)
			}
		}
		//sleep 10 min to save next time
		time.Sleep(time.Minute * 10)
	}
}
