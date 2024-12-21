---
slug: functionList
title: 前端 - 函数列表
type:  docs
sidebar_position: 21
---
:::tip  有那些doptime-client的函数可以使用？
## doptime-client 的函数列表
:::
  
```typescript title="doptime-client.d.ts"
declare class OptionClass {
    private UrlItems;
    Header: {
        [key: string]: string;
    };
    ThrowPromiseError: boolean;
    Urlbase: string;
    private optionObject;
    private withUrlField;
    rspTypeJson: () => OptionClass;
    rspTypeJpeg: () => OptionClass;
    rspTypeOgg: () => OptionClass;
    rspTypeMpeg: () => OptionClass;
    rspTypeMp4: () => OptionClass;
    rspTypeText: () => OptionClass;
    rspTypeStream: () => OptionClass;
    rspTypeAny: (anyType: string) => OptionClass;
    WithRds: (dataSourceName: string) => OptionClass;
    withUrlbase: (urlbase: string) => OptionClass;
    ThrowSecondaryPromiseError: (allowed: boolean) => OptionClass;
    paramString: () => string;
    constructor();
}
export declare const Option: OptionClass;
export declare const configure: (UrlBase?: string, JWT?: string, PrimaryErrorHandler?: Function) => void;
export default configure;
export declare enum urlGetCmd {
    HEXISTS = "HEXISTS",
    GET = "GET",
    HGET = "HGET",
    HGETALL = "HGETALL",
    HMGET = "HMGET"
}
export declare const urlGet: (cmd: urlGetCmd | undefined, Key: string, Field?: string, opt?: OptionClass) => string;
export declare const time: (opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hExists: (Key: string, Field?: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hset: (Key: string, Field: string | undefined, data: any, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const get: (Key: string, Field?: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hGet: (Key: string, Field?: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hDel: (Key: string, Field?: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hGetAll: (Key: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hVals: (Key: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hKeys: (Key: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hRandField: (Key: string, Count: number, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const hMGet: (Key: string, Fields?: any[], opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRange: (Key: string, Start: number, Stop: number, WITHSCORES?: boolean, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRevRange: (Key: string, Start: number, Stop: number, WITHSCORES: boolean, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRank: (Key: string, Member: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zScore: (Key: string, Member: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRangeByScore: (Key: string, Min: number | string, Max: number | string, WITHSCORES: boolean, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRevRangeByScore: (Key: string, Max: number | string, Min: number | string, WITHSCORES: boolean, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zAdd: (Key: string, Score: number, Member: any, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRem: (Key: string, Member: any, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zRemRangeByScore: (Key: string, Min: number, Max: number, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zCount: (Key: string, Min: number, Max: number, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const zCard: (Key: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const sIsMember: (Key: string, Member: string, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;
export declare const api: (serviceName: string, data?: any, opt?: OptionClass) => Promise<import("axios").AxiosResponse<any, any>>;

```

### doptime-client 的github地址是：[**doptime-client**](https://github.com/doptime/doptime-client)
查看doptime-client [**TypeScript源码**](https://github.com/doptime/doptime-client/blob/main/src/index.ts)  

## dopitme-client 重要说明
1. 后端返回的Content-Type 是在前端指定的。这通过指定opt?: OptionClass 来实现。  
    默认的返回类型是"application/json"，你可以不用指定。  
    可以像这样指定返回的Content-Type。比如 Option.rspTypeJpeg()。  
    如果需要指定任意的其它的类型， 你可以使用 Option.rspTypeAny(...)  
2. doptime-client 提供了一个函数urlGet，用于生成url。你可以用这个url来访问诸如图片、音频等资源。
