# Layer 8 Web Services

Advanced Web Server, Client, Webhook & Reverse Proxy for the Layer 8 Framework

## Overview

**Layer 8 Web Services** (l8web) is a comprehensive Go-based web infrastructure library designed specifically for the Layer 8 distributed systems framework. It provides RESTful HTTP and GraphQL endpoints, webhook handling, reverse proxy capabilities, and client libraries that seamlessly integrate with the Layer 8 network overlay, enabling secure web-based access to distributed services.

## Features

### Web Server
- **RESTful API Server**: Full HTTP REST server supporting GET, POST, PUT, PATCH, and DELETE methods
- **Protocol Buffers Integration**: Native support for Protocol Buffers serialization/deserialization via protojson
- **TLS/HTTPS Support**: Built-in SSL/TLS support with automatic certificate generation using Layer 8 utilities
- **Multi-Layer Authentication**: Bearer token, API key, and Two-Factor Authentication support
- **Service Discovery**: Automatic registration and discovery of web services via Layer 8 VNet
- **Plugin System**: Dynamic loading of service plugins with hot-reload capability
- **Multi-cast Communication**: Integration with Layer 8's proximity-based routing
- **Web UI Serving**: Dynamic file serving with SPA support and directory-level routing

### Webhook Handler
- **Provider Interface**: Pluggable webhook provider system for different VCS platforms
- **GitHub Provider**: Event detection via `X-GitHub-Event` header, HMAC-SHA256 signature verification via `X-Hub-Signature-256`
- **GitLab Provider**: Event detection via `X-Gitlab-Event` header, secret token verification via `X-Gitlab-Token`
- **Signature Verification**: HMAC-SHA256 signature validation for payload integrity
- **Issue Reference Extraction**: Parses commit messages for issue refs (`Fixes #42`, `Closes L8B-123`, `Resolves <uuid>`)
- **POST-Only Enforcement**: Rejects non-POST requests automatically

### Reverse Proxy
- **SNI-Based Routing**: TLS Server Name Indication for multi-domain certificate selection
- **Multi-Domain Support**: Handle multiple domains with per-domain configuration (layer8vibe.dev, probler.dev, layer-8.dev, layer8-book.help, l8erp.one)
- **Multi-Port Support**: Route traffic on multiple ports (443, 14443, 9092, 9094, 5444, 3114, 2883)
- **SSL Termination**: Automatic SSL certificate management for proxied domains
- **WebSocket Support**: Full WebSocket protocol proxying via Go's http.ReverseProxy
- **Environment Configuration**: NODE_IP environment variable support for dynamic backend routing

### REST Client
- **HTTP/HTTPS Client**: Full-featured REST client with connection pooling
- **Bearer Token Authentication**: Automatic token refresh with configurable auth endpoints
- **API Key Authentication**: Custom header-based auth (X-USER-ID, X-API-KEY)
- **Certificate Management**: Custom CA certificate support with certificate pinning
- **Compression**: GZIP compression with automatic content negotiation
- **Retry Logic**: Automatic retry on timeout with 5-second backoff (up to 5 attempts)
- **Configurable Endpoints**: Flexible URL construction with prefix support

### GraphQL Client
- **Full GraphQL Support**: Query and mutation operations with variable support
- **Error Handling**: Comprehensive GraphQL error parsing and reporting
- **Authentication**: Both Bearer token and API key authentication methods
- **SSL/TLS Support**: Secure connections with custom certificate support
- **Response Mapping**: Automatic mapping of GraphQL responses to Protocol Buffer messages
- **Retry Logic**: Built-in retry mechanism for timeout and connection issues

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Reverse Proxy (SNI-based)                 │
│              Multi-domain SSL termination & routing          │
└─────────────┬───────────────────────────────┬───────────────┘
              │                               │
    ┌─────────▼──────────┐          ┌────────▼──────────┐
    │   REST Server      │          │   REST Server      │
    │  ┌──────────────┐  │          │  ┌──────────────┐  │
    │  │ Service      │  │          │  │ Service      │  │
    │  │ Handler      │  │          │  │ Handler      │  │
    │  ├──────────────┤  │          │  ├──────────────┤  │
    │  │ Auth Layer   │  │          │  │ Auth Layer   │  │
    │  │ (Bearer/API) │  │          │  │ (Bearer/API) │  │
    │  ├──────────────┤  │          │  ├──────────────┤  │
    │  │ Webhook      │  │          │  │ WebService   │  │
    │  │ Handler      │  │          │  │ (TFA, Reg)   │  │
    │  ├──────────────┤  │          │  ├──────────────┤  │
    │  │ Layer 8 VNic │  │          │  │ Layer 8 VNic │  │
    │  └──────────────┘  │          │  └──────────────┘  │
    └────────────────────┘          └────────────────────┘
              │                               │
    ┌─────────▼───────────────────────────────▼───────────┐
    │              Layer 8 Network Overlay                 │
    │        (Service Discovery & Message Routing)         │
    └──────────────────────────────────────────────────────┘
              │                               │
    ┌─────────▼──────────┐          ┌────────▼──────────┐
    │   REST Client      │          │  GraphQL Client    │
    │  • Bearer Auth     │          │  • Query/Mutation  │
    │  • API Key Auth    │          │  • Variables       │
    │  • GZIP Compress   │          │  • Error Handling  │
    │  • Retry Logic     │          │  • Proto Mapping   │
    └────────────────────┘          └────────────────────┘
```

### Project Structure

```
l8web/
├── README.md
├── go/
│   ├── go.mod                          # Go module (Go 1.26.1)
│   ├── test.sh                         # Test runner script
│   ├── web/
│   │   ├── server/                     # REST Server implementation
│   │   │   ├── RestServer.go           # Core HTTP/HTTPS server
│   │   │   ├── ServiceHandler.go       # HTTP request handler with routing
│   │   │   ├── WebService.go           # Service manager (auth, TFA, registration)
│   │   │   ├── LoadWebUI.go            # Web UI file serving with SPA support
│   │   │   ├── CoockieToken.go         # Token extraction (header/cookie/query)
│   │   │   ├── TFA.go                  # Two-Factor Authentication (TOTP)
│   │   │   └── BodyToProto.go          # HTTP body to Protocol Buffer parsing
│   │   ├── client/                     # REST Client implementation
│   │   │   └── RestClient.go           # REST client with auth & retry
│   │   ├── gclient/                    # GraphQL Client
│   │   │   └── GraphQLClient.go        # GraphQL client implementation
│   │   ├── webhook/                    # Webhook handling
│   │   │   ├── webhook.go              # Core handler, Provider interface, EventHandler
│   │   │   ├── signature.go            # HMAC-SHA256 signature verification
│   │   │   ├── refs.go                 # Issue reference extraction from commit messages
│   │   │   ├── github/                 # GitHub webhook provider
│   │   │   │   └── github.go           # X-GitHub-Event, X-Hub-Signature-256 support
│   │   │   └── gitlab/                 # GitLab webhook provider
│   │   │       └── gitlab.go           # X-Gitlab-Event, X-Gitlab-Token support
│   │   └── proxy/                      # Reverse Proxy
│   │       ├── reverse_proxy.go        # SNI-based multi-domain routing proxy
│   │       ├── main/main.go            # Proxy executable entry point
│   │       ├── proxy.yaml              # Kubernetes DaemonSet config
│   │       ├── build.sh                # Docker build script
│   │       ├── Dockerfile              # Multi-stage Docker build
│   │       └── README.md               # Proxy documentation
│   └── tests/                          # Test suite
│       ├── TestRestServer_test.go      # REST server tests
│       ├── TestAuth_test.go            # Authentication tests
│       ├── TestWeb_test.go             # Web integration tests
│       ├── TestWebhook_test.go         # Webhook handler tests
│       ├── TestGitHub_test.go          # GitHub webhook provider tests
│       ├── TestSignature_test.go       # Signature verification tests
│       ├── TestRefs_test.go            # Issue reference extraction tests
│       ├── TestUtils.go                # Test utilities
│       └── TestInit.go                 # Test initialization
```

## Dependencies

- **Go 1.26.1+**
- **Protocol Buffers 3.0+** - google.golang.org/protobuf
- **Layer 8 Framework** - Core distributed systems framework

### Layer 8 Dependencies
- `github.com/saichler/l8bus` - Layer 8 bus/overlay networking
- `github.com/saichler/l8types` - Type definitions and interfaces
- `github.com/saichler/l8utils` - Utility functions (certs, IP, maps)
- `github.com/saichler/l8srlz` - Serialization utilities
- `github.com/saichler/l8services` - Service management (indirect)
- `github.com/saichler/l8test` - Testing infrastructure

## Installation

```bash
go get github.com/saichler/l8web/go
```

## Usage

### Creating a Web Server

```go
import (
    "github.com/saichler/l8web/go/web/server"
)

// Configure the server
config := &server.RestServerConfig{
    Host:           "localhost",
    Port:           8080,
    Authentication: true,        // Enable authentication
    CertName:       "server",    // Auto-generate SSL cert if missing
    Prefix:         "/api/v1/",
}

// Create the server
srv, err := server.NewRestServer(config)
if err != nil {
    log.Fatal(err)
}

// Register with Layer 8 service manager
srv.RegisterWebService(webService, vnic)

// Start the server (blocking)
go srv.Start()
```

### Creating a REST Client with Bearer Token

```go
import (
    "github.com/saichler/l8web/go/web/client"
)

// Configure client with bearer token authentication
config := &client.RestClientConfig{
    Host:          "localhost",
    Port:          8080,
    Https:         true,
    CertFileName:  "ca.crt",
    TokenRequired: true,
    Prefix:        "/api/v1/",
    AuthInfo: &client.RestAuthInfo{
        NeedAuth:   true,
        BodyType:   "AuthRequest",
        UserField:  "username",
        PassField:  "password",
        RespType:   "AuthResponse",
        TokenField: "token",
        AuthPath:   "/auth",
    },
}

// Create the client
restClient, err := client.NewRestClient(config, resources)
if err != nil {
    log.Fatal(err)
}

// Authenticate (obtains and stores bearer token)
err = restClient.Auth("myuser", "mypassword")

// Make authenticated requests
response, err := restClient.GET("/users", "UserList", "", "", nil)
response, err = restClient.POST("/users", "User", "", "", newUser)
response, err = restClient.DELETE("/users/123", "", "", "", nil)
```

### Using API Key Authentication

```go
import (
    "github.com/saichler/l8web/go/web/client"
)

// Configure client with API key authentication
config := &client.RestClientConfig{
    Host:  "api.example.com",
    Port:  443,
    Https: true,
    AuthInfo: &client.RestAuthInfo{
        IsAPIKey: true,
        ApiUser:  "app-client-001",      // Sent as X-USER-ID header
        ApiKey:   "sk-1234567890abcdef", // Sent as X-API-KEY header
    },
}

restClient, err := client.NewRestClient(config, resources)
// All requests automatically include API key headers
response, err := restClient.GET("/protected/data", "DataResponse", "", "", nil)
```

### Using GraphQL Client

```go
import (
    "github.com/saichler/l8web/go/web/gclient"
)

// Configure GraphQL client
config := &gclient.GraphQLClientConfig{
    Host:     "api.example.com",
    Port:     443,
    Https:    true,
    Endpoint: "/graphql",
    AuthInfo: &gclient.GraphQLAuthInfo{
        IsAPIKey: true,
        ApiUser:  "app-client-001",
        ApiKey:   "sk-1234567890abcdef",
    },
}

// Create the client
gqlClient, err := gclient.NewGraphQLClient(config, resources)
if err != nil {
    log.Fatal(err)
}

// Execute a GraphQL query
query := `
    query GetUser($id: ID!) {
        user(id: $id) {
            id
            name
            email
        }
    }
`
variables := map[string]interface{}{
    "id": "user123",
}
response, err := gqlClient.Query(query, variables, "UserResponse", "user")

// Execute a GraphQL mutation
mutation := `
    mutation CreatePost($input: PostInput!) {
        createPost(input: $input) {
            id
            title
        }
    }
`
variables = map[string]interface{}{
    "input": map[string]interface{}{
        "title":   "New Post",
        "content": "Post content here",
    },
}
response, err = gqlClient.Mutate(mutation, variables, "PostResponse", "createPost")
```

### Using Webhook Handlers

```go
import (
    "github.com/saichler/l8web/go/web/webhook"
    "github.com/saichler/l8web/go/web/webhook/github"
)

// Create a GitHub webhook handler
handler := webhook.NewHandler(
    &github.Provider{},
    func(eventType string, payload []byte) int {
        switch eventType {
        case "push":
            var push github.PushEvent
            json.Unmarshal(payload, &push)
            // Extract issue references from commit messages
            for _, commit := range push.Commits {
                refs := webhook.ExtractIssueRefs(commit.Message)
                // Process refs (e.g., close issues)
            }
            return http.StatusOK
        case "pull_request":
            var pr github.PullRequestEvent
            json.Unmarshal(payload, &pr)
            // Handle PR events
            return http.StatusOK
        default:
            return http.StatusOK
        }
    },
    func(payload []byte) string {
        // Look up webhook secret for the repo
        repoURL := github.RepoURL(payload)
        return lookupSecret(repoURL)
    },
)

// Register the handler on your HTTP mux
mux.Handle("/webhooks/github", handler)
```

### Service Registration

```go
// Register a web service handler type
webNic.Resources().Services().RegisterServiceHandlerType(&server.WebService{})

// Activate the web service
service, err := webNic.Resources().Services().Activate(
    server.ServiceTypeName,
    ifs.WebService,
    0,  // Auto-assign port
    webNic.Resources(),
    webNic,
    srv,
)

// Service is now discoverable via Layer 8
log.Printf("Service registered: %s", service.ServiceName())
```

## Authentication

### Token Extraction Priority

The server extracts authentication tokens in this order:
1. **Authorization Header**: `Authorization: Bearer {token}`
2. **Cookie**: `bToken` cookie value
3. **Query Parameter**: `?token={token}`

### Built-in Endpoints

The WebService component provides these endpoints:
- `/auth` - User authentication (returns bearer token)
- `/register` - User registration with CAPTCHA
- `/captcha` - CAPTCHA challenge generation
- `/tfaSetup` - Two-Factor Authentication setup (returns QR code)
- `/tfaSetupVerify` - TFA verification
- `/registry` - Type registry access

### Two-Factor Authentication Flow

```go
// 1. Setup TFA (returns QR code URL and secret)
POST /tfaSetup

// 2. Verify TFA setup with TOTP code
POST /tfaSetupVerify
Body: { "code": "123456" }

// 3. Login with TFA
POST /auth
Body: { "username": "...", "password": "...", "tfaCode": "123456" }
```

## Configuration

### Server Configuration

| Field | Type | Description |
|-------|------|-------------|
| Host | string | Server bind address |
| Port | int | Server port number |
| Authentication | bool | Enable/disable authentication |
| CertName | string | Certificate name for HTTPS (auto-generates if missing) |
| Prefix | string | URL prefix for all endpoints |

### Client Configuration

| Field | Type | Description |
|-------|------|-------------|
| Host | string | Target server hostname |
| Port | int | Target server port |
| Https | bool | Enable HTTPS connections |
| TokenRequired | bool | Require bearer token authentication |
| Token | string | Pre-configured authentication token |
| CertFileName | string | CA certificate file for verification |
| Prefix | string | URL prefix for requests |

### Authentication Info

| Field | Type | Description |
|-------|------|-------------|
| NeedAuth | bool | Enable authentication flow |
| IsAPIKey | bool | Use API key instead of bearer token |
| ApiUser | string | API user ID (X-USER-ID header) |
| ApiKey | string | API key (X-API-KEY header) |
| AuthPath | string | Authentication endpoint path |
| BodyType | string | Auth request message type |
| RespType | string | Auth response message type |
| TokenField | string | Field containing token in response |

## Testing

```bash
# Run tests with coverage
./go/test.sh

# Or run directly
cd go
go test -v -coverprofile=cover.html ./...
go tool cover -html=cover.html
```

The test suite includes:
- REST server creation and configuration
- Service registration and discovery
- Client-server communication
- Protocol Buffer serialization/deserialization
- Authentication flows (Bearer, API Key)
- Adjacent VNet token mapping
- Webhook handler (POST enforcement, event dispatch, error handling)
- GitHub webhook provider (event type, signature verification)
- HMAC-SHA256 signature verification
- Issue reference extraction from commit messages

## Security Features

- **TLS/HTTPS**: Full SSL/TLS encryption with auto-generated certificates
- **Certificate Pinning**: Custom CA certificate support for enhanced security
- **Bearer Token Auth**: Secure token-based authentication with refresh capability
- **API Key Auth**: Machine-to-machine authentication via custom headers
- **Two-Factor Auth**: TOTP-based second factor authentication
- **CAPTCHA Support**: Bot protection for registration flows
- **Webhook Signature Verification**: HMAC-SHA256 payload validation for GitHub; token verification for GitLab
- **Adjacent Token Mapping**: Cross-VNet authentication support

## Integration with Layer 8

l8web is deeply integrated with the Layer 8 framework:

- **VNic Integration**: All communication routes through Layer 8's Virtual Network Interface
- **Service Discovery**: Automatic discovery of available services via multicast
- **Proximity Routing**: Leverages Layer 8's proximity-based message routing
- **Plugin System**: Dynamic loading of service implementations
- **Resource Management**: Integrated with Layer 8's registry and resource system
- **Security Integration**: Token validation via Layer 8's security layer

## License

Copyright (c) 2025 Sharon Aicler (saichler@gmail.com)

Licensed under the Apache License, Version 2.0. See [LICENSE](http://www.apache.org/licenses/LICENSE-2.0) for details.

## Contributing

This project is part of the Layer 8 framework ecosystem. Contributions are welcome via pull requests.

## Related Projects

- [layer8](https://github.com/saichler/layer8) - Core Layer 8 Framework
- [l8types](https://github.com/saichler/l8types) - Type definitions and interfaces
- [l8bus](https://github.com/saichler/l8bus) - Layer 8 bus/overlay networking
- [l8utils](https://github.com/saichler/l8utils) - Utility functions
- [l8services](https://github.com/saichler/l8services) - Service management
- [l8test](https://github.com/saichler/l8test) - Testing infrastructure
