---
slug: frontend-Usage
title: 前端 - 如何使用
type:  docs
sidebar_position: 20
---

:::tip 第一步，安装
## 安装doptime-client
::: 
doptime-client 是一个用于前端的客户端库。它提供了一组函数，用于与doptime后端进行通信。

### 通过node 使用doptime-client
```bash
npm install doptime-client
```

### 通过拷贝typescript源码 使用doptime-client  
如果你仅使用typescript,而不使用 nodejs, 你可以直接拷贝doptime-client的源码到你的项目中。   
查看doptime-client [**TypeScript源码**](https://github.com/doptime/doptime-client/blob/main/src/index.ts)  | [**TypeScript源码 Raw**](https://raw.githubusercontent.com/doptime/doptime-client/main/src/index.ts)


### 通过拷贝js源码使用doptime-client  
通过js,而不使用 nodejs, 你可以直接拷贝doptime-client的源码到你的项目中。  
查看doptime-client [**JS源码**](https://github.com/doptime/doptime-client/blob/main/lib/index.js)   |   [**JS源码 Raw**](https://raw.githubusercontent.com/doptime/doptime-client/main/lib/index.js)  


:::tip 第二步，配置
## 通过 configure 初始化配置
::: 
你需要先初始化doptime-client,以确保doptime-client 能够正确的与doptime后端通信。   
> export declare const configure: (UrlBase?: string, JWT?: string, PrimaryErrorHandler?: Function) => void; 
  
参数说明：
- UrlBase: api server的地址。
- JWT: JWT token。  
- PrimaryErrorHandler: 默认错误处理函数，用来处理自动登录重定向等。  


```javascript   title="test.js"
const PrimaryErrorHandler = (err) => {
    if (err.response.status === 401) {
        window.location.href = "/login"
    }
}
configure("http://localhost:80", "jwtstring", PrimaryErrorHandler)   
```



:::tip 第三步，使用
## 使用doptime-client
::: 
这是使用doptime-client的示例代码。  
```javascript   title="test.js"

import { api, configure, time } from "doptime-client"
import React, { useEffect } from 'react';
import { useState } from 'react';

const Test = () => {
    const [urlBase, setUrlbase] = useState("https://api.doptime.com")
    const [tm, setTM] = useState(0)
    const [apistr, setAPI] = useState("")
    configure(urlBase)
    useEffect(() => {
        time().then((res) => {
            //convert unix timestamp to human readable time
            var time = new Date(res ).toLocaleString();
            setTM(time);
        })
        api("hello", { text: "world" }).then((res) => {
            setAPI(res)
        })
    }, [])
    return (
        <div class="flex-col ml-10">
            <div class="flex-col space-x-6 text-gray-400 ">
                <div class="text-lg">Test current api server (you can modify it to test your api server):</div>
                {/* modifiable urlBase:  */}
                <div class="text-2xl  text-yellow-600 mb-7 ring-1">
                    <input type="text" value={urlBase} onChange={(e) => setUrlbase(e.target.value)} />
                </div>
            </div> 
            <div class="flex-col space-x-4 text-gray-400">
                <h2>result of server Time</h2>
                <p class="text-xl  text-red-600">{tm}</p>
            </div>
            <div class="flex-col space-x-4 ">
                <h2 class="text-xl text-gray-400">result of calling demoapi</h2>
                <div class="text-xl text-blue-600">{apistr}</div>
            </div>


        </div>
    )
}
export default Test
```
[**查看运行结果**](/test)

