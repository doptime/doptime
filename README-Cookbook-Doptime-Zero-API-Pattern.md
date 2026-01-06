# Doptime Zero-API Pattern: Architecture Reference
**Version:** 1.0.0
**Status:** Stable

## 1. Overview

The **Doptime Zero-API Pattern** eliminates the traditional backend service layer (Controller → Service → DAO) by exposing secure, type-safe database keys directly to the frontend client. 

This architecture leverages **Struct Tags** for logic injection and **Redis Hash Structures** for data organization, ensuring automatic multi-tenancy and zero-boilerplate CRUD operations.

### Core Value Proposition
* **Zero Glue Code:** No manual API endpoints or fetch wrappers.
* **Unified Typing:** Frontend and Backend share identical data definitions.
* **Auto-Tenancy:** Data isolation is enforced at the storage schema level via `@sub`.
* **Atomic Operations:** Direct utilization of Redis Hash capabilities.

---

## 2. Backend Specification (Go)

The backend is responsible solely for **Schema Definition** and **Access Control Configuration**.

### 2.1 Schema Definition
Logic is defined declaratively using `mod` tags.

```go
package schema

import "[github.com/doptime/redisdb](https://github.com/doptime/redisdb)"

type Profile struct {
    // @field: Auto-injects the Redis Hash Field (UUID) into this struct property.
    // When creating, the backend generates a UUID. When reading, it maps the Key.
    ID string `json:"id" mod:"default=@field" validator:"nonempty"`

    // @sub: Auto-injects the current Subject (User ID) from the Auth Token.
    // Enforces Row-Level Security; the client cannot tamper with this value.
    ParentID string `json:"pid" mod:"default=@sub"`

    // Business Data
    Name   string `json:"name"`
    Avatar string `json:"avatar"`
    Grade  int    `json:"grade"`
    Energy int    `json:"energy" mod:"default=100"`
}

```

### 2.2 Key Configuration

Expose the data structure with specific permissions.

```go
// Pattern: "Profile:%s"
// The "%s" is strictly bound to the user's ID via the client context.
// Structure: Redis Hash <Field: UUID, Value: JSON>
var HKeyProfile = redisdb.NewHashKey[string, *Profile]("Profile:%s").HttpOn(
    redisdb.HGet | redisdb.HGetAll | redisdb.HSet | redisdb.HDel,
)

```

---

## 3. Frontend Specification (TypeScript)

The frontend interacts with the database using the generated `keyProfile` client object. No HTTP fetch logic is required.

### 3.1 Client Import

```typescript
import { keyProfile } from '@/lib/doptime';
import type { Profile } from '@/lib/doptime/types'; // Auto-generated types

```

### 3.2 CRUD Implementation Reference

| Operation | Method | Key / ID Strategy | Description |
| --- | --- | --- | --- |
| **List** | `hGetAll()` | N/A | Fetches all items in the user's hash (`Profile:{uid}`). |
| **Create** | `hSet()` | `Key: "@uuid"` | Passing `"@uuid"` triggers ID generation on the backend. |
| **Update** | `hSet()` | `Key: {id}` | Updates specific field in the hash. Fails if ID doesn't exist. |
| **Delete** | `hDel()` | `Key: {id}` | Removes the item from the hash. |

### 3.3 Code Examples

#### Fetch List

```typescript
const fetchList = async (): Promise<Profile[]> => {
  // Returns Map<UUID, Profile>. Convert to Array for UI.
  const dataMap = await keyProfile.hGetAll();
  return Object.values(dataMap || {});
};

```

#### Create New Item

```typescript
const createProfile = async (data: Partial<Profile>) => {
  // Magic string "@uuid" instructs backend to generate a new UUID
  await keyProfile.hSet("@uuid", {
    name: "New Agent",
    grade: 4,
    ...data
  });
};

```

#### Update Item

```typescript
const updateProfile = async (profile: Profile) => {
  // Uses existing ID to locate and overwrite the record
  await keyProfile.hSet(profile.id, profile);
};

```

#### Delete Item

```typescript
const deleteProfile = async (id: string) => {
  await keyProfile.hDel(id);
};

```

---

## 4. Security & Architecture Constraints

1. **Context Injection (`@sub`)**:
* **Rule**: All user-specific data structs MUST include a field tagged with `mod:"default=@sub"`.
* **Effect**: This binds the data to the authenticated user's ID, preventing cross-tenant data leaks without manual SQL `WHERE` clauses.


2. **ID Management (`@field`)**:
* **Rule**: The primary identifier MUST be tagged with `mod:"default=@field"`.
* **Effect**: Ensures the JSON payload `id` matches the Redis Hash Key (Field), maintaining data integrity.


3. **Hash vs. List**:
* **Recommendation**: Use `NewHashKey` (Redis Hash) for collections of objects. It supports O(1) access for Read/Update/Delete by ID, whereas Redis Lists do not.


