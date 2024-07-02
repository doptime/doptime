package api

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

type ApiDocs struct {
	KeyName  string
	UpdateAt int64
	ParamIn  interface{}
	ParamOut interface{}
}

var ApiDocsMap cmap.ConcurrentMap[string, *ApiDocs] = cmap.New[*ApiDocs]()

var SynWebDataRunOnce = sync.Mutex{}

func (a *Context[i, o]) RegisterApiDoc() (err error) {
	_, ok := ApiDocsMap.Get(a.Name)
	if ok {
		return nil
	}
	webdata := &ApiDocs{
		KeyName:  a.Name,
		UpdateAt: time.Now().Unix(),
	}

	vType := reflect.TypeOf((*i)(nil)).Elem()
	if webdata.ParamIn, err = specification.InstantiateType(vType); err != nil {
		return err
	}
	oType := reflect.TypeOf((*o)(nil)).Elem()
	if webdata.ParamOut, err = specification.InstantiateType(oType); err != nil {
		return err
	}
	ApiDocsMap.Set(a.Name, webdata)
	if SynWebDataRunOnce.TryLock() {
		go syncWithRedis()
	}
	return nil
}

func syncWithRedis() {
	//wait arrival of other schema to be store in map
	time.Sleep(time.Second)
	for {
		now := time.Now().Unix()
		//only update local defined data to redis
		var localStructuredDataMap = make(map[string][]byte)
		ApiDocsMap.IterCb(func(key string, value *ApiDocs) {
			value.UpdateAt = now
			localStructuredDataMap[key], _ = msgpack.Marshal(value)
		})
		if len(localStructuredDataMap) > 0 {
			client, ok := config.Rds["default"]
			if ok {
				client.HSet(context.Background(), "ApiDocs", localStructuredDataMap)
			}
		}
		//sleep 10 min to save next time
		time.Sleep(time.Minute * 10)
	}
}
