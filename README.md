# Layer 8 Web Services

Advanced Web Server, Client & Reverse Proxy for the Layer 8 Framework

## Overview

**Layer 8 Web Services** (l8web) is a comprehensive Go-based web infrastructure library designed specifically for the Layer 8 distributed systems framework. It provides RESTful HTTP and GraphQL endpoints, reverse proxy capabilities, and client libraries that seamlessly integrate with the Layer 8 network overlay, enabling secure web-based access to distributed services.

## Recent Updates (2025)

### Latest Features
- **GraphQL Client Support**: Full-featured GraphQL client with query/mutation operations and Protocol Buffer mapping
- **API Key Authentication**: Support for API key-based authentication via custom headers (X-USER-ID, X-API-KEY)
- **Two-Factor Authentication (TFA)**: TOTP-based two-factor authentication with QR code setup
- **Enhanced Authentication System**: Flexible authentication configuration for both REST and GraphQL clients
- **User Registration with CAPTCHA**: Built-in user registration flow with CAPTCHA verification

### Recent Improvements
- Fixed loading sequence issues
- Enhanced REST client authentication flow with automatic token refresh
- Added support for multiple authentication methods (Bearer, API Key, TFA)
- Improved proxy configuration with multi-port SNI-based routing
- Better retry logic and timeout handling for network requests

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

### Reverse Proxy
- **SNI-Based Routing**: TLS Server Name Indication for multi-domain certificate selection
- **Multi-Domain Support**: Handle multiple domains with per-domain configuration
- **Multi-Port Support**: Route traffic on multiple ports (443, 14443, 9092, 9094, etc.)
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
    │  │ WebService   │  │          │  │ WebService   │  │
    │  │ (TFA, Reg)   │  │          │  │ (TFA, Reg)   │  │
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
├── README.md                           # Main documentation
├── web.html                            # Project website
├── go/
│   ├── go.mod                          # Go module definition
│   ├── test.sh                         # Test runner script
│   ├── web/
│   │   ├── server/                     # REST Server implementation
│   │   │   ├── RestServer.go           # Core HTTP/HTTPS server
│   │   │   ├── ServiceHandler.go       # HTTP request handler with routing
│   │   │   ├── WebService.go           # Service manager (auth, TFA, registration)
│   │   │   ├── LoadWebUI.go            # Web UI file serving
│   │   │   ├── CoockieToken.go         # Token extraction (header/cookie/query)
│   │   │   ├── TFA.go                  # Two-Factor Authentication
│   │   │   └── BodyToProto.go          # HTTP body to Protocol Buffer parsing
│   │   ├── client/                     # REST Client implementation
│   │   │   └── RestClient.go           # REST client with auth & retry
│   │   ├── gclient/                    # GraphQL Client
│   │   │   └── GraphQLClient.go        # GraphQL client implementation
│   │   └── proxy/                      # Reverse Proxy
│   │       ├── reverse_proxy.go        # SNI-based routing proxy
│   │       ├── main/main.go            # Proxy executable entry point
│   │       ├── proxy.yaml              # Kubernetes DaemonSet config
│   │       └── README.md               # Proxy documentation
│   └── tests/                          # Test suite
│       ├── TestRestServer_test.go      # REST server tests
│       ├── TestAuth_test.go            # Authentication tests
│       ├── TestWeb_test.go             # Web integration tests
│       ├── TestUtils.go                # Test utilities
│       └── TestInit.go                 # Test initialization
```

## Dependencies

- **Go 1.25.4+**
- **Protocol Buffers 3.0+** - google.golang.org/protobuf
- **Layer 8 Framework** - Core distributed systems framework

### Layer 8 Dependencies
- `github.com/saichler/l8bus` - Layer 8 bus/overlay networking
- `github.com/saichler/l8types` - Type definitions and interfaces
- `github.com/saichler/l8utils` - Utility functions (certs, IP, maps)
- `github.com/saichler/l8srlz` - Serialization utilities
- `github.com/saichler/l8services` - Service management

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
            posts {
                title
                content
            }
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
            createdAt
        }
    }
`
variables = map[string]interface{}{
    "input": map[string]interface{}{
        "title":    "New Post",
        "content":  "Post content here",
        "authorId": "user123",
    },
}
response, err = gqlClient.Mutate(mutation, variables, "PostResponse", "createPost")
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

## Security Features

- **TLS/HTTPS**: Full SSL/TLS encryption with auto-generated certificates
- **Certificate Pinning**: Custom CA certificate support for enhanced security
- **Bearer Token Auth**: Secure token-based authentication with refresh capability
- **API Key Auth**: Machine-to-machine authentication via custom headers
- **Two-Factor Auth**: TOTP-based second factor authentication
- **CAPTCHA Support**: Bot protection for registration flows
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
