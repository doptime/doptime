package rdsdb

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	cmap "github.com/orcaman/concurrent-map/v2"
)

var mapWebDataDocs cmap.ConcurrentMap[string, *WebDataDocs] = cmap.New[*WebDataDocs]()

var NextSaveWebDataDocTime int64 = 0
var KeyWebDataDocs = HashKey[string, *WebDataDocs](Option.WithKey("WebDataDocs"))

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

type WebDataDocs struct {
	KeyName string
	// string, hash, list, set, zset, stream
	KeyType  string
	Instance interface{}
}

func (ctx *Ctx[k, v]) RegisterWebData(keyType string) {
	var validRdsKeyTypes = map[string]bool{"string": true, "list": true, "set": true, "hash": true, "zset": true, "stream": true}
	if _, ok := validRdsKeyTypes[keyType]; !ok {
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
	dataSchema := &WebDataDocs{KeyName: rootKey, KeyType: keyType, Instance: value}
	mapWebDataDocs.Set(ctx.Key, dataSchema)
	SyncWebDataWithRedis()
}

func SyncWebDataWithRedis() {
	now := time.Now().Unix()
	NextSaveWebDataDocTime = now
	time.Sleep(time.Second * 3)
	//another thread is trying to save DataSchema to redis
	if now != NextSaveWebDataDocTime {
		return
	}
	var localStructuredDataMap = make(map[string]*WebDataDocs)
	mapWebDataDocs.IterCb(func(key string, value *WebDataDocs) {
		localStructuredDataMap[key] = value
	})

	KeyWebDataDocs.HSet(localStructuredDataMap)

	if vals, err := KeyWebDataDocs.HGetAll(); err == nil {
		for k, v := range vals {
			mapWebDataDocs.Set(k, v)
		}
	}
}
