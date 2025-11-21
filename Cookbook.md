这是一份经过深度修正、基于您提供的源码（`Option.ts`, `hashkey.ts` 等）重构的 **Doptime Framework Master Context (v4.0)**。

此文档纠正了之前关于 `data.New`（后端）和全局函数（前端）的错误描述，严格遵循\*\*面向对象（Class-based）\*\*的数据操作模式。请将此文档作为 System Prompt 发送给 LLM。

-----

# Doptime Framework: Master Developer Context (v4.0)

**Core Philosophy:**

1.  **Frontend-Driven Data:** Frontend (`doptime-client`) accesses DB directly via **Key Classes** (`hashKey`, `listKey`).
2.  **Dragonfly-Native:** Backend logic is minimized; storage relies on DragonflyDB (Redis-compatible).
3.  **String Keys Only:** Integers are **forbidden** as keys. Use Strings.

-----

## 1\. Infrastructure & Config

**Database:** DragonflyDB.
**Config:** `config.toml` (Local) or `CONFIG_URL` (Prod).

**Config.toml (Dev Profile):**

```toml
[[Redis]]
  Name = "default"
  Host = "127.0.0.1"
  Port = 6379
[Http]
  Port = 80
  AutoAuth = true # CRITICAL: Auto-whitelists Frontend operations in Dev
```

-----

## 2\. Frontend Development (`doptime-client`)

**Pattern:** Object-Oriented. Instantiate a Key Class -\> Call Methods.
**Imports:** `import { OptDefaults, hashKey, listKey, createApi } from "doptime-client"`.

### 2.1 Initialization

Configure global defaults at app startup.

```typescript
import { OptDefaults } from "doptime-client";

// Set BaseURL and Error Handler (e.g., for 401 redirects)
OptDefaults({
    urlBase: "https://api.myapp.com",
    token: "jwt_string", // Optional: Pre-configured JWT
    primaryErrorHandler: (err) => {
        if (err.response?.status === 401) location.href = "/login";
    }
});
```

### 2.2 Data Access: The Class-Based Pattern

**Rule:** Do **NOT** use global `hGet`/`hSet` functions. You must `new` a specific Key class.

#### A. Hash (Object/Map)

```typescript
import { hashKey } from "doptime-client";

// Define schema type for TypeScript safety
type UserProfile = { theme: string, name: string };

// 1. Instantiate Key (Supports @id binding for JWT user)
// Key "u:@id" -> "u:1001" automatically by Backend
const userStore = new hashKey<UserProfile>("u:@id");

// 2. Operate
await userStore.hSet("theme", { theme: "dark", name: "Alex" }); // Set field
const val = await userStore.hGet("theme"); // Get field
const all = await userStore.hGetAll(); // Get all fields
```

#### B. List (Queue/Stack)

```typescript
import { listKey } from "doptime-client";

const taskQueue = new listKey<string>("tasks:pending");
await taskQueue.rPush("job_1");
const job = await taskQueue.lPop();
```

#### C. Other Key Types

  * `stringKey<T>`: Simple Key-Value (`get`, `set`).
  * `zSetKey<T>`: Sorted Sets (`zAdd`, `zRange`).
  * `streamKey<T>`: Redis Streams (`xAdd`, `xRead`).

### 2.3 Options & Content Types

Use `Opt` chaining to modify requests per call.

```typescript
import { Opt } from "doptime-client";

// Example: Get binary data (msgpack) or plain text
await userStore.hGet("avatar", Opt.WithResponseAsJson());
// Example: Use a different Redis Datasource
await userStore.hGetAll(Opt.WithDataSource("cache_db"));
```

-----

## 3\. Backend Development (Go)

**Lang:** Go 1.24+
**Package:** `github.com/doptime/doptime`

### 3.1 Data Access (Specific Key Types)

**Correction:** `data.New` does **not** exist. Use specific factory methods from `redisdb`.

```go
import "github.com/doptime/redisdb"

// 1. Define Key Accessor Globally
// Syntax: redisdb.New[Type]Key[KeyType, ValType](Options...)
var UserStore = redisdb.NewHashKey[string, *User](
    redisdb.WithKey("u"), // Base key name
    redisdb.WithRds("default"),
)

var TaskQueue = redisdb.NewListKey[string, string](
    redisdb.WithKey("tasks"),
)
```

### 3.2 Defining API (Business Logic)

Use `createApi` in Frontend to call these.

```go
import "github.com/doptime/doptime/api"

// Logic exposed as "api:process"
var ProcessApi = api.Api(func(req *In) (*Out, error) {
    // Complex logic here...
    return &Out{}, nil
}).Func
```

-----

## 4\. Security & Best Practices

### 4.1 The "String Key" Rule

**Constraint:** JavaScript destroys large Integers (scientific notation).

  * ❌ `new hashKey("order").hGet(1234567890123456789)` -\> Fails.
  * ✅ `new hashKey("order").hGet("1234567890123456789")` -\> Safe.
  * **Instruction:** Always cast IDs to String in both Frontend and Backend.

### 4.2 Context Binding

  * **@id**: Backend replaces this with JWT `UserID`.
  * **@role**: Backend replaces this with JWT `Role`.
  * *Frontend:* `new hashKey("user:@id").hGetAll()` is secure by default.

### 4.3 Content Addressing

For caching computable content (e.g., TTS, Files), use **xxHash (64-bit)** to generate keys.

-----

## 5\. Meta-Instructions for AI Code Generation

1.  **Frontend Code Gen:** Always generate `new [Type]Key("name")` instances. Never generate global `hGet(key, ...)` calls.
2.  **Backend Code Gen:** Use `redisdb.NewHashKey` (or `NewListKey`, `NewStringKey`), **NOT** `data.New`.
3.  **Imports:** Ensure `doptime-client` imports match the `index.ts` exports (`hashKey`, `Opt`, `createApi`).
4.  **Config:** If the user asks "My permissions are denied", tell them to check `AutoAuth = true` in `config.toml`.