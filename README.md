## doptime, the most concise, redis based web-server framework
    the name doptime borrow from "杀悟",means kill bad wisdom。 I hate bad tools.
### major advantages on API design
* APIs you defined, support both monolithic and microservice architecture. perfect!
* very simple to define and use API. just see the golang example demo.
* No API version related problem. Just see web client demo.
* Very easy to upgrade API , just change the data structure. no extra schema definition needed.
* You don't need to write any CREATE GET PUT or DELETE  Logic. Just use redis to query，modify or delete. That means most CURD can be done at frontend, needs no backend job.
* You can focus on operations with multiple data logic only.  We call it "API".
    doptime will put API data to redis stream, and the API receive and process the stream data.
* redis pipeline  brings high batch process performance.  
### major advantages on Data Op
* Using most welcomed redis compatible db. no database but redis compatible KEYDB. With flash storage supportion, KEYDB brings both memory speed and whole disk capacity
* Very Easy to define and access data. see keyInDemo.HSET(req.Id, req) in golang example.
 - Schema data is adopted to keep maintain-bility. Easy to upgrade data structure.
### other features
* Use msgpack to support structure data by default. Easily to upgrade data sturecture.
* All HTTP requests are transferd as binary msgpack data. It's compact and fast.
* allow specify Content-Type in web client.
* allow specify response fields in web client to reduce web traffic
* support JWT for authorization
* fully access control
* support CORS
  
