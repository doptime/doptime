---
slug: http-api
title: 后端 - http - api
type:  docs
sidebar_position: 6
---


:::tip 入参映射 - 从http、rpc参数到api入参
## 使用 mapstructure tag 定义api入参
::: 
doptime 采用 **mapstructure** 用于定义http的入参，它可以实现入参的别名和类型转化。   
1. mapstructure可帮助你免于类型转换和变量名大小写转换等繁琐工作。    
2. mapstructure 可以将string 自动转化为int*,uint*, float*, bool等类型。同样，也可以把int*, float*, bool*等类型转化为string。  
3. 注意，从string 转化 int*时，采用10进制转化。如果你期望不同进制，请使用string类型来接受数据。然后再自行转化。
4. mapstructure 的匹配和转化适用用 http请求和rpc请求。  
> **[[查看mapstructure文档]](https://pkg.go.dev/github.com/mitchellh/mapstructure?utm_source=godoc#hdr-Field_Tags)**  




:::tip 入参验证 - 确保你的参数符合要求:
## 使用 validator tag 验证入参
::: 
doptime 采用 **validator** 用于验证http的入参。  
    1. validator 是一个强大的数据验证库，它可以验证结构体中的字段。
> **[[查看validator文档]](https://pkg.go.dev/github.com/go-playground/validator/v10?utm_source=godoc)**  &nbsp;&nbsp;&nbsp;&nbsp;  [**[validator 官方example]**](https://github.com/go-playground/validator/blob/master/_examples/struct-level/main.go)



:::tip 使用mapstructure 和validator的示例代码:
## 使用 mapstructure validator的 示例代码
::: 

```go   title="main.go"
type Person struct {
    Name string
}
type InDemo struct {
    //要求不能是零值,否则返回错误
    Id   string `mapstructure:"JwtId" validate:"required,min=16,max=64"`
    //相当于 `mapstructure:"HeaderRemoteAddr"`   
    HeaderRemoteAddr        string  
    HeaderUserAgent string
    Username  string `validate:"required,min=5,max=20"`
    Email     string `mapstructure:"email" validate:"required,email"`
    FavouriteColor string     `validate:"hexcolor|rgb|rgba"`
    Addresses      []*Address `validate:"required,dive,required"` 
    //squash 嵌入结构
    Person `mapstructure:",squash"`
    //omitempty. if 
    Age int `mapstructure:""  validate:"gte=18,lte=120`
    //相当于 `mapstructure:"Text"` 或是 `mapstructure:"text"`
    //mapstructure tag is not always needed
    Text string 
    Discount  float64 `json:"discount" validate:"percentage"`
    //other param will be kept as map
    Other map[string]interface{} `mapstructure:",remain"`
}
```


:::tip 特殊约定
## http 入参的特殊约定
::: 


### **特殊约定之一: Header 前缀**
- 要引用http header 信息，入参的变量名应该添加 Header 前缀  
- Header前缀的参数包括HeaderRemoteAddr,HeaderHost,HeaderMethod,HeaderPath,HeaderQuery
- 入参如无Header打头的字段，不会在Other字段中看到header信息。  
- 入参含有Header打头的字段，则会在Other中看到其它未使用的header信息。

### **特殊的约定之二: Jwt 前缀**
- **JwtXxx**  
  JWT中的每一个属性名xxx, 会被修改为JwtXxx(CamelCase格式)作为新的属性名称。例如 id 修改为JwtId

-  **阻止Jwt伪造**   
    在客户端直接传递Jwt*参数（比如 Jwtid）是无效的，会被认为是试图伪造Jwt.请换成其它名称。

    
:::tip 从api 返回值到http响应
## http 响应代码 与 返回内容
::: 
api 包括两个返回参数，一个是任意类型的返回值，一个是错误 err。http 的响应是对这个返回值的处理。
```go   title="main.go"
//define api with input/output data structure
ApiDemo:=api.Api(func(req *InMyStruct) (ret interface{}, err error) {
 ..
}).Func
```
### http status code
- 如果err是JWT相关的错误。就会返回401错误  
- 其它的错误，会返回500错误。  
- 如果没有错误。则返回200。 
### 返回值的处理
- 同时任意的返回值如果是string 或者是byte[],则直接返相应的数据。
- 其它类型会被转化为json格式返回。



