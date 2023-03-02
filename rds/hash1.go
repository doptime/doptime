package rds

import (
	"context"
	"errors"
	"reflect"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/yangkequn/saavuu/logger"
)

func HGet1(ctx context.Context, rc *redis.Client, key string, field interface{}, value *interface{}) (err error) {
	fieldBytes, err := msgpack.Marshal(field)
	if err != nil {
		return err
	}
	cmd := rc.HGet(ctx, key, string(fieldBytes))
	data, err := cmd.Bytes()
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(data, value)
}

func HSet1(ctx context.Context, rc *redis.Client, key string, field interface{}, value interface{}) (err error) {
	fieldBytes, err := msgpack.Marshal(field)
	if err != nil {
		return err
	}
	valueBytes, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}
	status := rc.HSet(ctx, key, fieldBytes, valueBytes)
	return status.Err()
}

func HGetMap(ctx context.Context, rc *redis.Client, key string, mapOut interface{}) (err error) {
	mapElem := reflect.TypeOf(mapOut)
	if mapElem.Kind() != reflect.Map {
		logger.Lshortfile.Println("mapOut must be a map[interface{}] struct/interface{}")
		return errors.New("mapOut must be a map[interface{}] struct/interface{}")
	}
	cmd := rc.HGetAll(ctx, key)
	data, err := cmd.Result()
	if err != nil {
		return err
	}
	//append all data to mapOut
	KeyStructSupposed := mapElem.Key()
	valueStructSupposed := mapElem.Elem()
	for k, v := range data {
		//make a copy of stru , to valObj
		keyObj := reflect.New(KeyStructSupposed).Interface()
		if err = msgpack.Unmarshal([]byte(k), &keyObj); err != nil {
			logger.Lshortfile.Println("HGetAll1: key unmarshal error:", err)
			continue
		}
		valObj := reflect.New(valueStructSupposed).Interface()
		if err = msgpack.Unmarshal([]byte(v), &valObj); err == nil {
			logger.Lshortfile.Println("HGetAll1: value unmarshal error:", err)
		}
		reflect.ValueOf(mapOut).SetMapIndex(reflect.ValueOf(keyObj).Elem(), reflect.ValueOf(valObj).Elem())
	}
	return err
}
func HSetMap(ctx context.Context, rc *redis.Client, key string, mapIn interface{}) (err error) {
	mapElem := reflect.TypeOf(mapIn)
	if mapElem.Kind() != reflect.Map {
		logger.Lshortfile.Println("mapIn must be a map[interface{}] struct/interface{}")
		return errors.New("mapIn must be a map[interface{}] struct/interface{}")
	}
	mapOut := make(map[string]interface{})
	//append all key value of mapIn to mapOut, using msgpack
	for _, k := range reflect.ValueOf(mapIn).MapKeys() {
		//marshal key to bytes
		keyBytes, err := msgpack.Marshal(k.Interface())
		if err != nil {
			logger.Lshortfile.Println("HSetMap: key marshal error:", err)
			continue
		}
		//marshal value to bytes
		valueBytes, err := msgpack.Marshal(reflect.ValueOf(mapIn).MapIndex(k).Interface())
		if err != nil {
			logger.Lshortfile.Println("HSetMap: value marshal error:", err)
			continue
		}
		mapOut[string(keyBytes)] = valueBytes
	}
	//hset to redis
	return rc.HSet(ctx, key, mapOut).Err()
}
func HMGET1(ctx context.Context, rc *redis.Client, key string, fields []interface{}, values *[]interface{}) (err error) {
	fieldBytes := make([]string, 0, len(fields))
	for _, v := range fields {
		b, err := msgpack.Marshal(v)
		if err != nil {
			logger.Lshortfile.Println("HMGET1: field marshal error:", err)
			continue
		}
		fieldBytes = append(fieldBytes, string(b))
	}
	cmd := rc.HMGet(ctx, key, fieldBytes...)
	data := cmd.Val()
	*values = make([]interface{}, 0, len(data))
	valueStruct := reflect.TypeOf(values).Elem().Elem()
	//unmarshal each value of cmd.Val() to interface{}, using msgpack
	for _, v := range data {
		obj := reflect.New(valueStruct).Interface()
		if err = msgpack.Unmarshal([]byte(v.(string)), &obj); err != nil {
			logger.Lshortfile.Println("HMGET1: value unmarshal error:", err)
			continue
		}
		*values = append(*values, obj)
	}
	return cmd.Err()
}
func isPointerToSlice(obj interface{}) (ok bool) {
	objType := reflect.TypeOf(obj)
	if objType.Kind() != reflect.Ptr {
		return false
	}
	if objType.Elem().Kind() != reflect.Slice {
		return false
	}
	return true
}

func HKeysMK(ctx context.Context, rc *redis.Client, key string, fields interface{}) (err error) {
	if !isPointerToSlice(fields) {
		logger.Lshortfile.Println("fields must be a pointer to slice")
		return errors.New("fields must be a pointer to slice")
	}
	cmd := rc.HKeys(ctx, key)
	// structFields := reflect.TypeOf(fields).Elem()
	// *fields = make([]interface{}, 0, len(cmd.Val()))
	structFields := reflect.TypeOf(fields).Elem().Elem()
	slice := reflect.MakeSlice(reflect.TypeOf(fields).Elem(), 0, len(cmd.Val()))
	reflect.ValueOf(fields).Elem().Set(slice)
	for _, v := range cmd.Val() {
		field := reflect.New(structFields).Interface()
		if err = msgpack.Unmarshal([]byte(v), &field); err != nil {
			logger.Lshortfile.Println("HKeys1: field unmarshal error:", err)
			continue
		}
		//*fields = append(*fields, field)
		reflect.ValueOf(fields).Elem().Set(reflect.Append(reflect.ValueOf(fields).Elem(), reflect.ValueOf(field).Elem()))
	}
	return cmd.Err()
}

func HVals(ctx context.Context, rc *redis.Client, key string, values *[]interface{}) (err error) {
	cmd := rc.HVals(ctx, key)
	data := cmd.Val()
	*values = make([]interface{}, 0, len(data))
	valueStruct := reflect.TypeOf(values).Elem().Elem()
	//unmarshal each value of cmd.Val() to interface{}, using msgpack
	for _, v := range data {
		obj := reflect.New(valueStruct).Interface()
		if err = msgpack.Unmarshal([]byte(v), &obj); err != nil {
			logger.Lshortfile.Println("HVals: value unmarshal error:", err)
			continue
		}
		*values = append(*values, obj)
	}
	return cmd.Err()
}
