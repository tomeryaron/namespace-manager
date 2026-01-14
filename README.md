# Namespace Manager

A REST API service for managing Kubernetes namespaces with TTL (Time To Live) functionality. Allows users to create, list, and delete namespaces with automatic expiration tracking.

## Architecture Overview

```
┌─────────────────┐
│   HTTP Client   │
│  (curl/browser) │
└────────┬─────────┘
         │ HTTP Requests
         │ (POST/GET/DELETE)
         ▼
┌─────────────────────────────────┐
│      HTTP Server Layer          │
│  (internal/httpserver)          │
│  - Routes requests              │
│  - Validates input              │
│  - Handles HTTP responses       │
└────────┬────────────────────────┘
         │
         │ Calls methods
         ▼
┌─────────────────────────────────┐
│    Kubernetes Client Layer       │
│    (internal/kube)               │
│  - Creates namespaces            │
│  - Lists namespaces             │
│  - Deletes namespaces           │
│  - Manages TTL annotations      │
└────────┬────────────────────────┘
         │
         │ Kubernetes API calls
         ▼
┌─────────────────────────────────┐
│    Kubernetes Cluster            │
│  - Namespace resources           │
│  - Annotations (TTL, owner)     │
└─────────────────────────────────┘
```

## Go Application Architecture

### Package Structure

```
namespace-manager/
├── cmd/app/
│   └── main.go              # Application entry point
├── internal/
│   ├── httpserver/          # HTTP server and handlers
│   │   ├── server.go        # Server struct and routing
│   │   └── handlers.go      # HTTP request handlers
│   └── kube/                # Kubernetes client operations
│       ├── client.go        # Kubernetes client initialization
│       └── namespaces.go    # Namespace CRUD operations
└── charts/                   # Helm chart for deployment
```

### Core Structs

#### 1. `httpserver.Server` (internal/httpserver/server.go)

The main HTTP server wrapper that handles routing and request handling.

```go
type Server struct {
    server    *http.Server   // Standard Go HTTP server
    mux       *http.ServeMux // Request router (maps URLs to handlers)
    kubeClient *kube.Client  // Kubernetes client for operations
}
```

**Responsibilities:**
- Manages HTTP server lifecycle
- Routes incoming requests to appropriate handlers
- Provides Kubernetes client to handlers

**Key Methods:**
- `NewServer(addr, kubeClient)` - Creates a new server instance
- `RegisterRoute(path, handler)` - Maps URL paths to handler functions
- `ListenAndServe()` - Starts the HTTP server

#### 2. `kube.Client` (internal/kube/client.go)

Wraps the Kubernetes clientset to interact with the Kubernetes API.

```go
type Client struct {
    clientset *kubernetes.Clientset // Kubernetes API client
    config    *rest.Config           // Cluster connection config
}
```

**Responsibilities:**
- Manages connection to Kubernetes cluster
- Auto-detects cluster context (in-cluster vs local)
- Provides methods for namespace operations

**Key Methods:**
- `NewClient()` - Creates client, auto-detects cluster config
  - If running in pod: uses service account (in-cluster config)
  - If running locally: uses `~/.kube/config`

#### 3. `kube.NamespaceInfo` (internal/kube/namespaces.go)

Represents namespace information with TTL details.

```go
type NamespaceInfo struct {
    Name      string    // Namespace name
    Owner     string    // Owner name (from annotation)
    Team      string    // Team name (from annotation)
    CreatedAt time.Time // Creation timestamp
    ExpiresAt time.Time // Expiration timestamp
    TTL       int       // Remaining TTL in hours
}
```

**Used for:**
- API responses when listing namespaces
- Converting Kubernetes namespace objects to API format

#### 4. Request/Response Structs (internal/httpserver/handlers.go)

**CreateNamespaceRequest:**
```go
type CreateNamespaceRequest struct {
    Name  string `json:"name"`  // Namespace name
    TTL   int    `json:"ttl"`   // Time to live in hours
    Owner string `json:"owner"` // Owner name
    Team  string `json:"team"`  // Team name
}
```

**DeleteNamespaceRequest:**
```go
type DeleteNamespaceRequest struct {
    Name string `json:"name"` // Namespace name to delete
}
```

**NamespaceResponse:**
```go
type NamespaceResponse struct {
    Message string `json:"message"` // Success/error message
    Name    string `json:"name"`   // Namespace name
}
```

### Request Flow

#### Creating a Namespace

```
1. Client sends POST /api/namespaces/create
   {
     "name": "my-namespace",
     "ttl": 24,
     "owner": "john",
     "team": "devops"
   }

2. HandleCreateNamespaceRequest receives request
   ├─ Validates HTTP method (must be POST)
   ├─ Parses JSON body into CreateNamespaceRequest struct
   ├─ Validates required fields (name, owner, team, ttl)
   └─ Calls s.kubeClient.CreateNamespace(...)

3. kube.Client.CreateNamespace executes
   ├─ Creates Kubernetes Namespace object
   ├─ Adds annotations: owner, team, expires_at
   └─ Calls Kubernetes API to create namespace

4. Response sent back to client
   {
     "message": "Namespace created successfully",
     "name": "my-namespace"
   }
```

#### Listing Namespaces

```
1. Client sends GET /api/namespaces/list?owner=john

2. HandleListNamespacesRequest receives request
   ├─ Validates HTTP method (must be GET)
   ├─ Extracts "owner" query parameter
   └─ Calls s.kubeClient.ListNamespaces(owner)

3. kube.Client.ListNamespaces executes
   ├─ Fetches all namespaces from Kubernetes
   ├─ Filters by owner (if provided)
   ├─ Extracts annotations (owner, team, expires_at)
   ├─ Calculates remaining TTL
   └─ Converts to []NamespaceInfo

4. Response sent back to client
   [
     {
       "name": "my-namespace",
       "owner": "john",
       "team": "devops",
       "created_at": "2024-01-15T10:00:00Z",
       "expires_at": "2024-01-16T10:00:00Z",
       "ttl": 20
     }
   ]
```

#### Deleting a Namespace

```
1. Client sends DELETE /api/namespaces/delete
   {
     "name": "my-namespace"
   }

2. HandleDeleteNamespaceRequest receives request
   ├─ Validates HTTP method (must be DELETE)
   ├─ Parses JSON body into DeleteNamespaceRequest struct
   └─ Calls s.kubeClient.DeleteNamespace(name)

3. kube.Client.DeleteNamespace executes
   ├─ Sends delete request to Kubernetes API
   ├─ Waits for actual deletion (polls every 500ms)
   ├─ Times out after 30 seconds if not deleted
   └─ Returns success when namespace is gone

4. Response sent back to client
   {
     "message": "Namespace deleted successfully",
     "name": "my-namespace"
   }
```

### Key Design Patterns

#### 1. **Separation of Concerns**
- **httpserver package**: Handles HTTP layer (routing, validation, responses)
- **kube package**: Handles Kubernetes API interactions
- Clear boundaries between layers

#### 2. **Dependency Injection**
- Kubernetes client is passed to Server during initialization
- Makes testing easier and dependencies explicit

#### 3. **Struct Embedding**
- Server wraps `http.Server` and `http.ServeMux`
- Provides clean API while leveraging standard library

#### 4. **Error Handling**
- Errors bubble up from kube layer to http layer
- HTTP layer converts errors to appropriate status codes

## API Endpoints

### POST `/api/namespaces/create`
Create a new namespace with TTL.

**Request Body:**
```json
{
  "name": "my-namespace",
  "ttl": 24,
  "owner": "john",
  "team": "devops"
}
```

**Response:** `201 Created`
```json
{
  "message": "Namespace created successfully",
  "name": "my-namespace"
}
```

### GET `/api/namespaces/list?owner=john`
List namespaces, optionally filtered by owner.

**Query Parameters:**
- `owner` (optional): Filter by owner name

**Response:** `200 OK`
```json
[
  {
    "name": "my-namespace",
    "owner": "john",
    "team": "devops",
    "created_at": "2024-01-15T10:00:00Z",
    "expires_at": "2024-01-16T10:00:00Z",
    "ttl": 20
  }
]
```

### DELETE `/api/namespaces/delete`
Delete a namespace (waits for actual deletion).

**Request Body:**
```json
{
  "name": "my-namespace"
}
```

**Response:** `200 OK`
```json
{
  "message": "Namespace deleted successfully",
  "name": "my-namespace"
}
```

## Building and Running

### Local Development

```bash
# Run directly with Go
go run ./cmd/app

# Or build and run macOS binary
go build -o namespace-manager-mac ./cmd/app
./namespace-manager-mac
```

### Docker Build

```bash
# Build Linux binary and Docker image
./build.sh

# Or manually:
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o namespace-manager ./cmd/app
docker build -t namespace-manager:latest .
```

### Kubernetes Deployment

```bash
# Deploy using Helm
helm install namespace-manager ./charts/namespace-manager

# Port forward to test
kubectl port-forward svc/namespace-manager 8080:8080
```

## Configuration

### Kubernetes Client

The application automatically detects the cluster connection:

- **In-cluster**: When running inside a Kubernetes pod, uses the service account token
- **Local development**: Uses `~/.kube/config` file

### Required RBAC Permissions

The service account needs these permissions:
- `namespaces`: `get`, `list`, `create`, `delete`, `watch`

See `charts/namespace-manager/templates/clusterrole.yaml` for the full RBAC configuration.

## Future Enhancements

- [ ] CronJob to automatically delete expired namespaces
- [ ] Get namespace TTL endpoint
- [ ] Update namespace TTL endpoint
- [ ] Authentication/authorization
- [ ] Namespace resource quotas management
