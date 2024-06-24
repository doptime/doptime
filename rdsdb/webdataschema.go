package rdsdb

import (
	"reflect"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var mapWebDataSchema cmap.ConcurrentMap[string, *WebDataSchema] = cmap.New[*WebDataSchema]()

var NextSaveWebDataSchemaTime int64 = 0
var KeyWebDataSchema = HashKey[string, *WebDataSchema](Option.WithKey("WebDataSchema"))

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

			if (field.Kind() == reflect.Struct || field.Kind() == reflect.Ptr) && !field.IsNil() {
				initializeFields(field)
			}
		}
	}
}

type WebDataSchema struct {
	KeyName string
	// string, hash, list, set, zset, stream
	KeyType  string
	Instance interface{}
}

func (ctx *Ctx[k, v]) RegisterWebDataSchema(keyType string) {
	var validRdsKeyTypes = map[string]bool{"string": true, "list": true, "set": true, "hash": true, "zset": true, "stream": true}
	if _, ok := validRdsKeyTypes[keyType]; !ok {
		return
	}
	// 获取 v 的类型
	vType := reflect.TypeOf((*v)(nil)).Elem()

	// 检查 vType 是否可以实例化
	if vType.Kind() == reflect.Interface || vType.Kind() == reflect.Ptr || vType.Kind() == reflect.Invalid {
		return
	}

	// 创建 v 的实例
	valueElem := reflect.New(vType).Elem()
	value := valueElem.Interface()

	// 确认实例化是否成功
	if reflect.ValueOf(value).IsNil() {
		return
	}
	initializeFields(valueElem)
	dataSchema := &WebDataSchema{
		KeyName:  ctx.Key,
		KeyType:  "string",
		Instance: value,
	}
	mapWebDataSchema.Set(ctx.Key, dataSchema)
	SyncStructuresWithRedis := func() {
		now := time.Now().Unix()
		NextSaveWebDataSchemaTime = now
		time.Sleep(time.Second * 3)
		//another thread is trying to save DataSchema to redis
		if now != NextSaveWebDataSchemaTime {
			return
		}
		var localStructuredDataMap = make(map[string]*WebDataSchema)
		mapWebDataSchema.IterCb(func(key string, value *WebDataSchema) {
			localStructuredDataMap[key] = value
		})

		KeyWebDataSchema.HSet(localStructuredDataMap)

		if vals, err := KeyWebDataSchema.HGetAll(); err == nil {
			for k, v := range vals {
				mapWebDataSchema.Set(k, v)
			}
		}
	}
	SyncStructuresWithRedis()
}
