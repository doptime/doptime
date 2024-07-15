package api

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/doptime/doptime/config"
	"github.com/doptime/doptime/specification"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/vmihailenco/msgpack/v5"
)

type DocsOfApi struct {
	KeyName  string
	ParamIn  interface{}
	ParamOut interface{}
	UpdateAt int64
}

var ApiDocsMap cmap.ConcurrentMap[string, *DocsOfApi] = cmap.New[*DocsOfApi]()

var ApiVendorMap cmap.ConcurrentMap[string, *VendorOption] = cmap.New[*VendorOption]()
var SynWebDataRunOnce = sync.Mutex{}

func (a *Context[i, o]) RegisterApiDoc(vendorinfo *VendorOption) (err error) {
	_, ok := ApiDocsMap.Get(a.Name)
	if ok {
		return nil
	}
	webdata := &DocsOfApi{
		KeyName:  a.Name,
		UpdateAt: time.Now().Unix(),
	}
	if vendorinfo != nil {
		webdata.KeyName = webdata.KeyName + strings.Split(vendorinfo.VendorAccountEmail, "@")[0]
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
	if vendorinfo != nil {
		ApiVendorMap.Set(a.Name+"_"+vendorinfo.VendorAccountEmail, vendorinfo)
	}
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
		if !ok || (ApiDocsMap.Count() == 0 && ApiVendorMap.Count() == 0) {
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
		if ok && ApiVendorMap.Count() > 0 {
			ApiVendorMap.IterCb(func(key string, value *VendorOption) {
				value.ActiveAt = now
				if bs, err := msgpack.Marshal(value); err == nil {
					clientpiepie.HSet(context.Background(), "Docs:ApiVendor", key, string(bs)).Err()
				}
			})
		}
		clientpiepie.Exec(context.Background())
		time.Sleep(time.Minute * 10)
	}
}
