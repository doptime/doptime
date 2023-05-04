## saavuu, the most concise, redis based web-server framework
    the name saavuu borrow from "杀悟",means kill bad wisdom。 I hate bad tools.
### main features
* All HTTP requests are transferd as binary msgpack data. It's compact and fast.
* No API version related problem. Just use redis api.
* Use msgpack to support structure data by default. Easily to upgrade data sturecture.
* Use no database but redis compatible KEYDB. With flash storage supportion, KEYDB brings both memory speed and whole disk capacity
* You don't need to write any CREATE GET PUT or DELETE  Logic. Just use redis to query，modify or delete. That means most CURD can be done at frontend, needs no backend job.
* You can focus on operations with multiple data logic only.  We call it "API".
    saavuu will put API data to redis stream, and the API receive and process the stream data.
* You can use any programming language you like. python or go or may be c# if you like.
* redis pipeline  brings high batch process performance.  
### other features
* allow specify Content-Type in web client.
* allow specify response fields in web client to reduce web traffic
* support JWT for authorization
* fully access control
* support CORS
### possible drawbacks
* saavuu's API has higher latency than monolithic web server or traditional RPC . 
  for saavuu, data flow in API is transfered : 
    client => saavuu => redis => api =>redis => saavuu => client. this usually takes 2ms in local network. 
  for traditional RPC, with data flow is transfered : 
    client => otherFramework => RPC => otherFramework => client
* For thoese data operations without api， saavuu is fast, with data flow is transfered :
    client => saavuu => redis => saavuu => client. this usually takes 1ms in local network.

  However, as you will find out, saavuu makes api (dynamic upgrade version/ new api) hot plugable, and bring down microservice's complexity to near zero. because saavuu is just a non editable redis proxy. so only the  API logic part is needed, and many cases, API logic is isn't needed at all, you just need to use develop nothing but using saavuu's redis api.
  
## demo usage
### server, go example:
```
package main

import (
	"github.com/yangkequn/saavuu/api"
)

type InDemo struct {
	Data   []uint16
	Id   string `msgpack:"alias:JWT_id"`
}
//define api with input/output data structure
var ApiDemo=api.Api(func(req *InDemo) (ret string, err error) {
    // your logic here
    if req.Id == "" || len(req.Data) == 0 {
        return nil, saavuu.ErrInvalidInput
    }
    return `{data:"ok"}`, nil
})
```

### server, python example:
```
class service_textToMp3(Service):
    def __init__(self):
        Service.__init__(self,"redis.vm:6379/0")
    def check_input_item(self, i):
            if "BackTo" not in i:
                return False
            return True
    def process(self,items):
        #your logic here
        for i in items:
            self.send_back(i,{"Result":input.value})
service_textToMp3().start()
```

### web client, javascript /typescript example:
```
HGET("UserInfo", id).then((data) => {
    //your logic here
})
```


## about configuration 
    configuration is read from enviroment variables. Make sure enviroment variables are added to your IDE (launch.json for vs code) or docker. 
    these are the default example:
```
    "RedisAddress_PARAM": "127.0.0.1:6379",
    "RedisPassword_PARAM": "",
    "RedisDb_PARAM": "0",
    "RedisAddress_DATA": "127.0.0.1:6379",
    "RedisPassword_DATA": "",
    "RedisDb_DATA": "0",
    "JWTSecret": "WyBJujUQzWg4YiQqLe9N36DA/7QqZcOkg2o=",
    "JWT_IGNORE_FIELDS": "iat,exp,nbf,iss,aud,sub,typ,azp,nonce,auth_time,acr,amr,at_hash,c_hash,updated_at,nonce,auth_time,acr,amr,at_hash,c_hash,updated_at",
    "CORS": "*",
    "MaxBufferSize": "3145728",
    "AutoPermission": "true",
```