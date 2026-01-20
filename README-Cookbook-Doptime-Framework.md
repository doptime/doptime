# Doptime Framework: Master Developer Context

**Core Philosophy:**

1. **Frontend-Driven Data:** The frontend (npm package `doptime-client`) acts as the "Controller," defining data access paths using **Key Classes** (`hashKey`, `listKey`) and Context Placeholders (`@sub`).
2. **Dragonfly-Native:** The backend is stateless, relying on DragonflyDB (Redis-compatible) for high-performance storage.
3. **String Keys Only:** Integers are **strictly forbidden** as keys to prevent JavaScript precision loss. Always cast IDs to Strings.
4. **Implicit & Context-Aware:** Go structs use **Mapper v2** for input binding and automatic context injection.
5. **Path-Style Keys:** **Strictly** use `/` as the separator for keys (e.g., `scope/key`). **NEVER** use `:`.

---

## 1. Infrastructure & Config

**Database:** DragonflyDB (Redis-compatible).
**Config:** `config.toml` (Local) or `CONFIG_URL` (Prod).

**Config.toml (Dev Profile):**

```toml
[[Redis]]
  Name = "default"
  Host = "127.0.0.1"
  Port = 6379
[Http]
  Port = 80

```

---

## 2. Frontend Development (`doptime-client`)

**Pattern:** Object-Oriented. Instantiate a Key Class -> Call Methods.

**0. Setup:**

```bash
npm install doptime-client

```

**Imports:** `import { configure, hashKey, listKey, zSetKey, createApi, urlGet, Opt } from "doptime-client"`.

### 2.1 Initialization (Entry Point)

Must run before any data requests (typically in `layout.tsx` or `main.ts`).

```typescript
configure({
    urlBase: "https://api.myapp.com",
    // Token: Static string OR Async Function (resolved once at init)
    token: async () => await fetchClerkToken(), 
    // Global Error Handler (e.g., 401 Redirect)
    primaryErrorHandler: (err) => { 
        if (err.response?.status === 401) window.location.href = "/login"; 
    },
    allowThrowError: false
});

```

### 2.2 Data Access: The Class-Based Pattern

**Strict Rule:** Do **NOT** use global `hGet`/`hSet` functions. You must instantiate a specific Key class.

#### A. Context-Aware Keys (Multi-Tenancy)

Use `@sub` in the key name. The backend automatically replaces it with the verified UserID.

```typescript
// Definition
export interface Profile {
    id: string;
    name: string;
    avatar: string;
}

// 1. Instantiate Key
// The frontend sends "profiles:@sub". 
// The backend replaces "@sub" with the UserID from the JWT Token.
// CORRECTION:always Use ':' before '@'.
export const keyProfile = new hashKey<Profile>("profiles:@sub");

// 2. Usage
// Create (Backend generates UUID if key is "@uuid")
const createProfile = async (name: string) => {
    // "@uuid" is a magic string that triggers backend ID generation
    await keyProfile.hSet("@uuid", { name, avatar: "default.png" });
}

// List (Get all profiles for current user)
const listProfiles = async () => {
    const map = await keyProfile.hGetAll();
    return Object.values(map || {}); // Convert Map to Array
}

```

#### B. Standard Keys (Shared/Static)

```typescript
// Shared Leaderboard (Sorted Set)
// CORRECTION: Use '/' separator.
const lb = new zSetKey<string>("game/leaderboard");
await lb.zRevRange(0, 9, true); 

// System Queue (List)
const queue = new listKey<string>("system/tasks");
await queue.lPush(JSON.stringify({ task: "cleanup" }));

```

### 2.3 RPC (Remote Procedure Call)

Use `createApi` only when pure CRUD is insufficient (e.g., complex validation, 3rd-party API calls).

```typescript
// 1. Define Types
type AuthSyncReq = { email: string };
type AuthSyncRes = { status: string };

// 2. Create Caller (matches Backend API name)
// CORRECTION: Use '/' separator.
const callAuthSync = createApi<AuthSyncReq, AuthSyncRes>("api/auth/sync");

// 3. Invoke
await callAuthSync({ email: "user@example.com" });

```

### 2.4 Assets (Images/Media)

**Strict Rule:** NEVER download binary blobs via `hGet`. Generate Direct URLs for `<img>`/`<video>` tags to leverage browser caching.

```typescript
// Generate: https://api.site/HGET-profiles:user_123?f=avatar&rt=image/jpeg
const getAvatarUrl = (userId: string) => {
    return urlGet(
        undefined,               // Default Op (HGET)
        `profiles:${userId}`,    // Key
        "avatar",                // Field
        Opt.WithResponseAsJpeg() // Return Type header
    );
};

```

---

## 3. Backend Development (Go)

**Lang:** Go 1.24+
**Package:** `github.com/doptime/doptime`
**Mapping Library:** `github.com/doptime/mapper` (v2)
**DB Library:** `github.com/doptime/redisdb`

### 3.1 Data Modeling (Struct Definition)

Structs serve three purposes: **Input Binding** (Mapper), **Storage** (Msgpack), and **Validation**.

**Tag Reference:**

| Tag | Context | Description | Example |
| --- | --- | --- | --- |
| `json` | Input | **Mapper v2**: Binds input JSON/Map. Supports defaults & context. | `json:"name @default"` |
| `msgpack` | Storage | **RedisDB**: Defines field name for storage. | `msgpack:"uid"` |
| `mod` | Pre-Save | **RedisDB**: Modifiers applied before saving. | `mod:"trim,lowercase"` |
| `validate` | Check | **Validator**: Rules via `go-playground/validator`. | `validate:"required,email"` |

**Example Struct:**

```go
type Profile struct {
    // [Context Injection & Storage]
    // json: binds to injected "@sub" (UserID).
    // msgpack: stores as "id".
    // validate: ensures it's not empty.
    ID string `json:"@@sub" msgpack:"id" validate:"required"` 
    
    // [Request Info Injection]
    // Binds to the client's IP address injected by the framework
    ClientIP string `json:"@@remoteAddr" msgpack:"-"`

    // [Implicit Mapping]
    // json: maps "Name" -> Name (implicit).
    // mod: trims whitespace before save.
    Name string `msgpack:"name" mod:"trim"`

    // [Default Values]
    // json: defaults to 4 if missing.
    Grade int `json:"@4" msgpack:"grade"` 
    
    // [System Fields]
    // Auto-handled by RedisDB (no tags needed)
    CreatedAt time.Time
    UpdatedAt time.Time
}

```

### 3.2 Data Access (RedisDB)

**Factory Pattern:** Use `redisdb.New{Type}Key` to define accessors.

**CRITICAL RULE - Generics:**
Most Key types (Hash, Set, ZSet, String, VectorSet, Stream) require **TWO** type parameters `[k, v]`.
Only List requires **ONE** type parameter `[v]`.

```go
import "github.com/doptime/redisdb"

// 1. HashKey [k, v]
var ProfilesKey = redisdb.NewHashKey[string, *Profile](
    redisdb.WithKey("profiles"), 
    redisdb.WithRds("secondary-dragonfly-db"),
).HttpOn(redisdb.HashAll)

// 2. SetKey [k, v] - SUPPLEMENTED
// Note: Must explicitly state [string, string] even if both are strings.
var DirtyIndex = redisdb.NewSetKey[string, string](
    redisdb.WithKey("sys/idx/sym"),
)

// 3. ZSetKey [k, v] - SUPPLEMENTED
var LeaderboardKey = redisdb.NewZSetKey[string, string](
    redisdb.WithKey("game/leaderboard"),
)

// 4. ListKey [v] - SUPPLEMENTED
// Note: ListKey only takes one type parameter [v].
var TaskQueue = redisdb.NewListKey[string](
    redisdb.WithKey("sys/tasks"),
)

```

### 3.3 Defining API (RPC)

Use `api.Api` to define logic callable by `createApi` in Frontend.

```go
import "github.com/doptime/doptime/api"

// Logic exposed as "/authsync", lowercase AuthSyncReq automatically with Req removed automatically.

var AuthSyncApi = api.Api(func(req *AuthSyncReq) (*AuthSyncRes, error) {
    // req is auto-filled using Mapper v2
    return &AuthSyncRes{Status: "ok"}, nil
})

```

### 3.4 Key Methods Reference

Complete list of methods for each Key type. **No `context.Context` required.**

#### `HashKey[k, v]`

* `HGet(field k) (v, error)`
* `HSet(values ...interface{}) (int64, error)` // Supports `map[k]v` or `k, v, k, v`
* `Save(value v) (int64, error)` // Auto-detects PK from struct
* `HMSet(kvMap map[k]v) (int64, error)`
* `HExists(field k) (bool, error)`
* `HGetAll() (map[k]v, error)`
* `HRandField(count int) ([]k, error)`
* `HRandFieldWithValues(count int) ([]k, []v, error)`
* `HMGET(fields ...interface{}) ([]v, error)`
* `HLen() (int64, error)`
* `HDel(fields ...k) error`
* `HKeys() ([]k, error)`
* `HVals() ([]v, error)`
* `HIncrBy(field k, increment int64) error`
* `HIncrByFloat(field k, increment float64) error`
* `HSetNX(field k, value v) error`
* `HScan(cursor uint64, match string, count int64) ([]k, []v, uint64, error)`

#### `ListKey[v]`

* `RPush(param ...v) error`
* `LPush(param ...v) error`
* `RPushX(param ...v) error`
* `LPushX(param ...v) error`
* `RPop() (v, error)`
* `LPop() (v, error)`
* `LRange(start, stop int64) ([]v, error)`
* `LRem(count int64, param v) error`
* `LSet(index int64, param v) error`
* `LIndex(ind int64) (v, error)`
* `BLPop(timeout time.Duration) (v, error)`
* `BRPop(timeout time.Duration) (v, error)`
* `BRPopLPush(dest string, timeout time.Duration) (v, error)`
* `LInsertBefore(pivot, param v) error`
* `LInsertAfter(pivot, param v) error`
* `LTrim(start, stop int64) error`
* `LLen() (int64, error)`

#### `SetKey[k, v]`

* `SAdd(members ...v) error` // Returns error only
* `SCard() (int64, error)`
* `SRem(members ...v) error`
* `SIsMember(param v) (bool, error)`
* `SMembers() ([]v, error)`
* `SScan(cursor uint64, match string, count int64) ([]v, uint64, error)`

#### `ZSetKey[k, v]`

* `ZAdd(members ...redis.Z) error`
* `ZRem(members ...interface{}) error`
* `ZRange(start, stop int64) ([]v, error)`
* `ZRangeWithScores(start, stop int64) ([]v, []float64, error)`
* `ZRevRange(start, stop int64) ([]v, error)`
* `ZRank(member interface{}) (int64, error)`
* `ZRevRank(member interface{}) (int64, error)`
* `ZScore(member interface{}) (float64, error)`
* `ZCard() (int64, error)`
* `ZCount(min, max string) (int64, error)`
* `ZIncrBy(increment float64, member interface{}) (float64, error)`
* `ZPopMax(count int64) ([]v, []float64, error)`
* `ZPopMin(count int64) ([]v, []float64, error)`

#### `StringKey[k, v]`

* `Get(Field k) (v, error)`
* `Set(key k, value v, expiration time.Duration) error`
* `Del(key k) error`
* `GetAll(match string) (map[k]v, error)`
* `SetAll(_map map[k]v) error`

#### `StreamKey[k, v]`

* `XAdd(args *redis.XAddArgs) (string, error)`
* `XDel(ids ...string) (int64, error)`
* `XLen() (int64, error)`
* `XRange(start, stop string) ([]redis.XMessage, error)`
* `XRead(args *redis.XReadArgs) ([]redis.XStream, error)`

#### `VectorSetKey[k, v]`

* `Search(query string, params ...interface{}) (int64, []v, error)`
* `Create(args ...interface{}) error`
* `DropIndex(deleteDocs bool) error`
* `Info() (map[string]interface{}, error)`

---

## 4. Security & Architecture Constraints

### 4.1 The "String Key" Rule

**Constraint:** JavaScript destroys large Integers (scientific notation).

* ❌ `new hashKey("order").hGet(1234567890123456789)` -> Fails.
* ✅ `new hashKey("order").hGet("1234567890123456789")` -> Safe.
* **Instruction:** Always cast IDs to String in both Frontend and Backend.

### 4.2 Context Injection Pattern

The "Zero-API" security model relies on the Framework (specifically `httpContext.go`) and Mapper working together.

1. **Tamper-Proofing:** The framework **removes** any user-provided keys starting with `@` from the input parameters.
2. **Injection:** The framework injects system context variables prefixed with `@`.

* **Auth Context:** `@sub` (UserID), `@email`, `@role`, etc. (from JWT).
* **Request Info:** `@remoteAddr` (IP), `@host`, `@method`, `@path`, `@rawQuery`.
* **Target Metadata:** `@key` (Redis Key), `@field` (Hash/List Field).

3. **Binding:** The Go struct uses `json:"@@variable"` (e.g., `json:"@@sub"`, `json:"@@remoteAddr"`) to safely bind these values.

### 4.3 No Context in Data Operations (SUPPLEMENTED)

**Rule:** `redisdb` functions (`HSet`, `Get`, `Set`, etc.) do **NOT** accept a `context.Context` argument. The framework handles timeouts and context internally.

* ❌ Incorrect: `key.HSet(ctx, k, v)`
* ✅ Correct: `key.HSet(k, v)`

---

## 5. Meta-Instructions for AI Code Generation

**When generating code, strictly follow these rules:**

1. **Frontend:**

* Ensure `npm install doptime-client` is assumed.
* Always generate `new [Type]Key("scope/key")`. **Separator is `/`.**
* Never generate global `hGet`/`hSet` calls.
* Use `urlGet` for image sources.

2. **Backend:**

* Use `redisdb.NewHashKey[k, v]`, `NewSetKey[k, v]`, `NewZSetKey[k, v]`, `NewListKey[v]`.
* **Generics:** Be precise. `SetKey` and `ZSetKey` need `[k, v]`. `ListKey` needs `[v]`.
* **Struct Tags:** Include `json` (Mapper v2 syntax), `msgpack` (Storage), and `validate`/`mod`.
* **Syntax:** Use space separators for `json` tags. **Never use commas**.
* **Context:** Use `@@` tags for context injection. **Do not pass `ctx` variable to DB methods.**

3. **Imports:** Ensure `doptime-client` imports match exports (`hashKey`, `Opt`, `createApi`, `urlGet`).

```

```