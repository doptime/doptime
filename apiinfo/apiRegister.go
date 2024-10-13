package apiinfo

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/doptime/config/cfgredis"
	"github.com/doptime/doptime/specification"
	"github.com/doptime/redisdb"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

type DocsOfApi struct {
	KeyName  string
	ParamIn  interface{}
	ParamOut interface{}
	UpdateAt int64
}

var KeyApiDataDocs = redisdb.HashKey[string, *DocsOfApi](redisdb.WithKey("Docs:Api"))

var ApiDocsMap cmap.ConcurrentMap[string, *DocsOfApi] = cmap.New[*DocsOfApi]()

var SynAPIRunOnce = sync.Mutex{}

func RegisterApi(Name string, paramInType reflect.Type, paramOutType reflect.Type) (err error) {
	_, ok := ApiDocsMap.Get(Name)
	if ok {
		return nil
	}
	webdata := &DocsOfApi{
		KeyName:  Name,
		UpdateAt: time.Now().Unix(),
	}

	//vType := reflect.TypeOf((*i)(nil)).Elem()
	if webdata.ParamIn, err = specification.InstantiateType(paramInType); err != nil {
		return err
	}
	//oType := reflect.TypeOf((*o)(nil)).Elem()
	if webdata.ParamOut, err = specification.InstantiateType(paramOutType); err != nil {
		return err
	}
	ApiDocsMap.Set(Name, webdata)
	if SynAPIRunOnce.TryLock() {
		go syncWithRedis()
	}
	return nil
}

func syncWithRedis() {
	//wait arrival of other schema to be store in map
	time.Sleep(time.Second)
	for {
		now := time.Now().Unix()

		client, ok := cfgredis.Servers.Get("default")
		if !ok || (ApiDocsMap.Count() == 0) {
			time.Sleep(time.Minute)
			continue
		}
		clientpiepie := client.Pipeline()
		if ok && ApiDocsMap.Count() > 0 {
			ApiDocsMap.IterCb(func(key string, value *DocsOfApi) {
				value.UpdateAt = now
				if bs, err := msgpack.Marshal(value); err == nil {
					clientpiepie.HSet(context.Background(), "Docs:Api", value.KeyName, string(bs)).Err()
				}
			})
		}
		clientpiepie.Exec(context.Background())
		time.Sleep(time.Minute * 10)
	}
}
