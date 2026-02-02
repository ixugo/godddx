<p align="center">
    <img src="./logo.png#gh-light-mode-only" alt="GoDDD Logo" width="550"/>
    <img src="./logo_dark.png#gh-dark-mode-only" alt="GoDDD Logo" width="550"/>
</p>

<p align="center">
    <a href="https://github.com/ixugo/goddd/releases"><img src="https://img.shields.io/github/v/release/ixugo/goddd?include_prereleases" alt="Version"/></a>
    <a href="https://github.com/ixugo/goddd/blob/master/LICENSE.txt"><img src="https://img.shields.io/dub/l/vibe-d.svg" alt="License"/></a>
	<a href="https://gin-gonic.com"><img width=30px  src="https://avatars.githubusercontent.com/u/7894478?s=48&v=4" alt="GIN"/></a>
    <a href="https://gorm.io"><img width=70px src="https://gorm.io/gorm.svg" alt="GORM"/></a>
</p>

<p align="center">
    <a href="./README.md">English</a> | <a href="./README_CN.md">中文</a>
</p>

# GoDDD — Domain-Driven Design Framework for Go

**GoDDD** = **Go** + **DDD** (Domain-Driven Design)

A REST API framework and code generation tool based on Clean Architecture principles, designed specifically for small to medium-sized Go projects.

---

## 🎯 Design Philosophy

### Clean Architecture in Go Practice

GoDDD is deeply inspired by [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html). After understanding its core principles, we designed a **pragmatic layered architecture** tailored for small to medium projects — preserving the core value of dependency inversion while avoiding the development burden of over-abstraction.

<p align="center">
    <img src="./docs/ddd.jpg" alt="Clean Architecture" width="800"/>
</p>

### Core Principles

**Dependency Rule**: Outer layers depend on inner layers; inner layers reverse dependencies through interfaces.

```
┌─────────────────────────────────────────────────────────────────────┐
│                      API Layer (Primary Adapter)                    │
│  internal/web/api/                                                  │
│  Responsibility: HTTP protocol conversion → Call Core → Response    │
└────────────────────────────────┬────────────────────────────────────┘
                                 │ depends on
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                   Core Layer (Domain/Business Core)                 │
│  internal/core/<domain>/                                            │
│  Responsibility: Business logic, domain models, port interfaces     │
│                                                                     │
│  ├─ core.go           # Core struct + Storer interface definition   │
│  ├─ port.go           # Secondary adapter interfaces (cross-domain) │
│  ├─ model.go          # Domain type definitions (non-GORM mapped)   │
│  ├─ <entity>.go       # Business methods + EntityStorer interface   │
│  ├─ <entity>.model.go # Domain model (GORM mapped)                  │
│  ├─ <entity>.param.go # Input parameter definitions                 │
│  ├─ adapter/          # Cross-domain adapter implementations        │
│  └─ store/<domain>db/ # Database implementation (secondary adapter) │
└───────────────┬─────────────────────────────────────────────────────┘
                │ implements
                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                  Store/Adapter Layer (Secondary Adapter)            │
│              Implements Storer interfaces defined in Core           │
└─────────────────────────────────────────────────────────────────────┘
```

### Design Principles

| Traditional View | GoDDD Position |
|-----------------|----------------|
| "Small projects don't need layering" | ❌ Layering is automated by tools, cost approaches zero |
| "Direct calls between domains are convenient" | ❌ Convenient now, technical debt forever |
| "Too many interfaces are troublesome" | ✅ godddx auto-generates, no boilerplate needed |

### Domain Isolation Principle

**Domains must be decoupled through interfaces**, even if two domains are closely related (e.g., message and user), they cannot directly import each other.

```go
// ❌ Wrong: message directly depends on user package
import "myapp/internal/core/user"

func (c *Core) AddMessage(ctx context.Context, in AddMessageInput) error {
    u := user.NewService().Get(in.UserID) // Breaks cohesion!
}

// ✅ Correct: Define interface in port.go, inject via adapter
type UserProvider interface {
    GetUserBrief(ctx context.Context, userID string) (*UserBrief, error)
}

func (c *Core) AddMessage(ctx context.Context, in AddMessageInput, provider UserProvider) error {
    user, _ := provider.GetUserBrief(ctx, in.To)
    // Use UserBrief type defined in this domain
}
```

**Adapter Implementation** (located in `adapter/` directory):
```go
// adapter/user.go
type UserAdapter struct {
    userCore user.Core  // Adapter can directly depend on other domain's Core
}

func (a *UserAdapter) GetUserBrief(ctx context.Context, userID string) (*message.UserBrief, error) {
    u, err := a.userCore.GetUser(ctx, userID)
    if err != nil {
        return nil, nil
    }
    return &message.UserBrief{ID: u.ID, Name: u.Name}, nil  // Convert to this domain's type
}
```

### Layer Responsibilities

| Layer | Responsibility | Guidelines |
|-------|---------------|------------|
| **API Layer** | HTTP protocol conversion | Only parameter binding, auth checks, calling Core, returning response |
| **Core Layer** | Business logic orchestration | Define all interfaces, contain domain models and business methods |
| **Store Layer** | Data persistence | Implement Storer interfaces defined in Core |
| **Adapter Layer** | Cross-domain collaboration | Implement Provider interfaces defined in Core |

**Core Mantra:**
```
Core imports none, interfaces define contracts;
Generation ensures consistency, assembly at the edge;
Types guard boundaries, tests need no dependencies;
Small steps run steady, refactoring becomes easy.
```

---

## 📚 Design References

- [Google API Design Guide](https://google-cloud.gitbook.io/api-design-guide)
- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design - Eric Evans](https://www.domainlanguage.com/ddd/)

---

## 🚀 Installation

```bash
go install github.com/ixugo/godddx@latest
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
```

---

## 🤖 MCP Integration (Cursor/Claude)

godddx supports running as an MCP (Model Context Protocol) server, allowing AI assistants to directly invoke code generation capabilities.

### Configure Cursor

Add MCP server configuration in Cursor settings (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "godddx": {
      "command": "godddx",
      "args": ["-mcp"]
    }
  }
}
```


### MCP Tool Description

Once configured, the AI assistant can use the `generate_ddd_code` tool:

| Parameter | Required | Description |
|-----------|----------|-------------|
| `content` | ✅ | Go domain model file content, must include package declaration and type struct definitions |
| `module` | ✅ | Go module name, e.g., `github.com/yourname/project` |
| `output_dir` | ✅ | Project root directory path (directory containing go.mod) |

### Usage Example

In Cursor, you can simply tell the AI:

> "Help me generate DDD code for a User domain model in the /path/to/myproject directory"

The AI will automatically call the godddx MCP tool to generate the code.

---

## 📖 Quick Start

### 1. Initialize Project

Clone the [goddd](https://github.com/ixugo/goddd) template, or initialize a new project:

```bash
go mod init myproject
```

### 2. Define Domain Model

Create `tables/<domain>/<entity>.go` file and define the entity structure:

```go
// Package name becomes the generated module directory name
package user

import "github.com/ixugo/goddd/pkg/orm"

type User struct {
    // Must include ID, CreatedAt, UpdatedAt
    ID        int      // Integer ID
    // ID     string   // Or string ID (use uniqueid.Core to generate)
    CreatedAt orm.Time
    UpdatedAt orm.Time

    Name string // Nickname (field comments become GORM comments)
    Age  int64  // Age
}
```

### 3. Generate Code

```bash
godddx -f tables/user/user.go
```

### 4. Register Routes

Call the generated `RegisterUser` function in your project to register the code with gin router.

---

## 💡 Tips

### String IDs

Use short IDs instead of UUIDs:

```go
import "github.com/ixugo/goddd/domain/uniqueid"

type User struct {
    ID uniqueid.Core
}
```

The generator handles this automatically, defaulting to 6-character random IDs. To modify the length, search for `NewUniqueID` function, or call `uni.UniqueIDWithCustomLen()` to specify the length.

It's recommended to define prefix constants for string IDs to distinguish entity types.

### Automatic Dependency Injection

When using this tool in the goddd project root directory, it will automatically update:
- `internal/web/api/provider.go`
- `internal/web/api/api.go`

### Model Requirements

Structs must include `ID`, `CreatedAt`, `UpdatedAt` fields, otherwise the generated code will need minor adjustments.

### Caching Recommendations

Caching code may not be necessary in early stages. Consider deleting the generated `store/cache` directory and enabling it later when performance optimization is needed.

### API Layer Parameter Filling

If an Input parameter is filled by the API layer (e.g., current logged-in user), its tag should use `json:"-"` with a comment:

```go
type FindMessageInput struct {
    web.PagerFilter
    ReceiverID string `json:"-"`    // Receiver ID (filled by API layer with current user)
    Type       string `form:"type"` // Message type
}
```

---

## ✅ Feature List

- [x] Generate 5 common CRUD operations (Create, Read, Update, Delete, Paginated Search)
- [x] Generate 5 common CRUD cache operations (Redis supported)
- [ ] Generate test functions for 5 common CRUD operations
- [ ] Generate API documentation for 5 common CRUD operations
- [ ] Support frontend-specified sorting in paginated queries

---

## ❓ FAQ

### Why not generate code from the database?

Tables often use JSON types, and reading from the database cannot generate JSON structs. Starting from Go structs better expresses domain models.

### What is the `CacheKey` method in the model for?

Generated cache code needs a key to get values. The `CacheKey` method determines the model's unique identifier:

- Defaults to `ID`
- Can be modified to other keys if ID is not the frequent query field
- Example: In the goddd project's token implementation, hash is used as the key

---

## 📜 License

[MIT License](./LICENSE.txt)
