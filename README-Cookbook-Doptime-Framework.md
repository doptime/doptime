# Doptime Framework: Master Developer Context 

**Core Philosophy:**

1. **Frontend-Driven Data:** The frontend (`doptime-client`) acts as the "Controller," defining data access paths using **Key Classes** (`hashKey`, `listKey`) and Context Placeholders (`@sub`).
2. **Dragonfly-Native:** The backend is stateless, relying on DragonflyDB (Redis-compatible) for high-performance storage.
3. **String Keys Only:** Integers are **strictly forbidden** as keys to prevent JavaScript precision loss. Always cast IDs to Strings.
4. **Implicit & Context-Aware:** Go structs use **Mapper v2** for input binding and automatic context injection.

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
const lb = new zSetKey<string>("game:leaderboard");
await lb.zRevRange(0, 9, true); 

// System Queue (List)
const queue = new listKey<string>("system:tasks");
await queue.lPush(JSON.stringify({ task: "cleanup" }));

```

### 2.3 RPC (Remote Procedure Call)

Use `createApi` only when pure CRUD is insufficient (e.g., complex validation, 3rd-party API calls).

```typescript
// 1. Define Types
type AuthSyncReq = { email: string };
type AuthSyncRes = { status: string };

// 2. Create Caller (matches Backend API name)
const callAuthSync = createApi<AuthSyncReq, AuthSyncRes>("api:auth:sync");

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

```go
import "github.com/doptime/redisdb"

// Define Key Accessor Globally
// "profiles" is the Key Scope prefix.
var ProfilesKey = redisdb.NewHashKey[string, *Profile](
    redisdb.WithKey("profiles"), 
    redisdb.WithRds("secondary-dragonfly-db"),
).HttpOn(redisdb.HashAll) // Enable CRUD from Frontend

```

### 3.3 Defining API (RPC)

Use `api.Api` to define logic callable by `createApi` in Frontend.

```go
import "github.com/doptime/doptime/api"

// Logic exposed as "api:auth:sync"
var AuthSyncApi = api.Api(func(req *AuthSyncReq) (*AuthSyncRes, error) {
    // req is auto-filled using Mapper v2
    return &AuthSyncRes{Status: "ok"}, nil
})

```

---

## 4. Security & Architecture Constraints

### 4.1 The "String Key" Rule

**Constraint:** JavaScript destroys large Integers (scientific notation).

* ❌ `new hashKey("order").hGet(1234567890123456789)` -> Fails.
* ✅ `new hashKey("order").hGet("1234567890123456789")` -> Safe.
* **Instruction:** Always cast IDs to String in both Frontend and Backend.

### 4.2 Context Injection Pattern

The "Zero-API" security model relies on the Framework and Mapper working together:

1. **Tamper-Proofing:** `httpContext.go` **removes** any user-provided keys starting with `@` from the input map.
2. **Injection:** The framework injects verified claims (e.g., `sub`) as keys with an `@` prefix (e.g., `{"@sub": "user_123"}`).
3. **Binding:** The Go struct uses `json:"@@sub"` to bind this secure value.

---

## 5. Meta-Instructions for AI Code Generation

**When generating code, strictly follow these rules:**

1. **Frontend:**
* Always generate `new [Type]Key("name")` instances.
* Never generate global `hGet`/`hSet` calls.
* Use `urlGet` for image sources.


2. **Backend:**
* Use `redisdb.NewHashKey` (or `NewListKey`, etc.).
* **Struct Tags:** Include `json` (Mapper v2 syntax), `msgpack` (Storage), and `validate`/`mod` as needed.
* **Syntax:** Use space separators for `json` tags. **Never use commas**.
* **Context:** Use `@@` tags only for context injection (e.g., `@@sub`).


3. **Imports:** Ensure `doptime-client` imports match exports (`hashKey`, `Opt`, `createApi`, `urlGet`).