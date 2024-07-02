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

type DocsOfApi struct {
	KeyName  string
	UpdateAt int64
	ParamIn  interface{}
	ParamOut interface{}
}

var ApiDocsMap cmap.ConcurrentMap[string, *DocsOfApi] = cmap.New[*DocsOfApi]()

var SynWebDataRunOnce = sync.Mutex{}

func (a *Context[i, o]) RegisterApiDoc() (err error) {
	_, ok := ApiDocsMap.Get(a.Name)
	if ok {
		return nil
	}
	webdata := &DocsOfApi{
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
		client, ok := config.Rds["default"]
		if ok && ApiDocsMap.Count() > 0 {
			clientpiepie := client.Pipeline()
			ApiDocsMap.IterCb(func(key string, value *DocsOfApi) {
				value.UpdateAt = now
				if bs, err := msgpack.Marshal(value); err == nil {
					clientpiepie.HSet(context.Background(), "Docs:Api", value.KeyName, string(bs)).Err()
				}
				clientpiepie.Exec(context.Background())
			})
		}
		time.Sleep(time.Minute * 10)
	}
}
