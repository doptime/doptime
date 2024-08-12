package rdsdb

import cmap "github.com/orcaman/concurrent-map/v2"

type HashInterface interface {
	HSet(values ...interface{}) error
	HMGETInterface(fields ...interface{}) ([]interface{}, error)
}

var hashKeyMap cmap.ConcurrentMap[string, HashInterface] = cmap.New[HashInterface]()

var hashKey = HashKey[string, interface{}](WithKey("_"))

// var zsetKey = ZSetKey[string, interface{}](WithKey("_"))
// var listKey = ListKey[string, interface{}](WithKey("_"))
// var setKey = SetKey[string, interface{}](WithKey("_"))
// var stringKey = StringKey[string, interface{}](WithKey("_"))

func HKeyInterface(key string, RedisDataSource string) (db HashInterface, err error) {
	hashInterface, ok := hashKeyMap.Get(key + ":" + RedisDataSource)
	if ok {
		return hashInterface, nil
	}
	db = &CtxHash[string, interface{}]{Ctx: hashKey.Duplicate(key, RedisDataSource)}
	return db, nil
}
