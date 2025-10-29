# Layer 8 Web Services

Advanced Web Server, Client & Reverse Proxy for the Layer 8 Framework

## Recent Updates

### Latest Features (2025)
- **GraphQL Client Support**: New GraphQL client with full query and mutation support
- **API Key Authentication**: Added support for API key-based authentication via custom headers (X-USER-ID, X-API-KEY)
- **Enhanced Authentication System**: Flexible authentication configuration for both REST and GraphQL clients
- **Improved Error Handling**: Better retry logic and timeout handling for network requests

### Recent Improvements
- Fixed loading sequence issues
- Enhanced REST client authentication flow
- Added support for multiple authentication methods
- Improved proxy configuration handling

## Overview

**Layer 8 Web Services** (l8web) is a comprehensive Go-based web infrastructure library designed specifically for the Layer 8 distributed systems framework. It provides RESTful HTTP and GraphQL endpoints, reverse proxy capabilities, and client libraries that seamlessly integrate with the Layer 8 network overlay, enabling secure web-based access to distributed services.

## Features

### Web Server
- **RESTful API Server**: Full HTTP REST server supporting GET, POST, PUT, PATCH, and DELETE methods
- **Protocol Buffers Integration**: Native support for Protocol Buffers serialization/deserialization
- **TLS/HTTPS Support**: Built-in SSL/TLS support with automatic certificate generation and management
- **Authentication System**: Enhanced bearer token authentication with configurable auth paths
- **Service Discovery**: Automatic registration and discovery of web services
- **Plugin System**: Dynamic loading of service plugins with hot-reload capability
- **Multi-cast Communication**: Integration with Layer 8's proximity-based routing
- **Development Mode**: Built-in development server with hot-reload and debug features

### Reverse Proxy (New)
- **Multi-Domain Support**: Handle multiple domains with per-domain configuration
- **SSL Termination**: Automatic SSL certificate management for proxied domains
- **Load Balancing**: Distribute requests across multiple backend servers
- **Header Manipulation**: Add, modify, or remove headers in transit
- **Path Rewriting**: Flexible URL path transformation rules
- **WebSocket Support**: Full WebSocket protocol proxying
- **Caching Layer**: Optional response caching for improved performance

### Web Client
- **HTTP/HTTPS Client**: Full-featured REST client with timeout handling and retry logic
- **Authentication**: Enhanced bearer token authentication with automatic token refresh
- **API Key Authentication**: Support for API key-based authentication via custom headers
- **Certificate Management**: Custom CA certificate support with certificate pinning
- **Compression**: GZIP compression support with automatic content negotiation
- **Configurable Endpoints**: Flexible URL construction and endpoint management
- **Connection Pooling**: Efficient connection reuse for improved performance
- **Request/Response Interceptors**: Middleware support for request/response modification

### GraphQL Client (New)
- **GraphQL Support**: Full-featured GraphQL client for query and mutation operations
- **Variable Support**: Support for GraphQL variables in queries and mutations
- **Error Handling**: Comprehensive GraphQL error parsing and reporting
- **Authentication**: Bearer token and API key authentication support
- **SSL/TLS Support**: Secure connections with custom certificate support
- **Response Mapping**: Automatic mapping of GraphQL responses to Protocol Buffer messages
- **Retry Logic**: Built-in retry mechanism for timeout and connection issues

## Architecture

The project follows a modular architecture with clear separation between server, proxy, and client components:

```
go/
├── web/
│   ├── server/          # Web server implementation
│   │   ├── RestServer.go       # Main REST server with enhanced auth
│   │   ├── ServiceHandler.go   # HTTP request handler
│   │   ├── WebService.go       # Service registration
│   │   └── LoadWebUI.go        # Web UI loader
│   ├── proxy/           # Reverse proxy implementation
│   │   ├── reverse_proxy.go    # Core proxy server
│   │   └── main/               # Proxy executable
│   ├── client/          # REST client implementation
│   │   └── RestClient.go       # Enhanced REST client with API key support
│   └── gclient/         # GraphQL client implementation (New)
│       └── GraphQLClient.go    # GraphQL client with query/mutation support
└── tests/              # Comprehensive test suite
    ├── TestWeb_test.go        # Web integration tests
    ├── TestAuth_test.go       # Authentication tests
    └── TestRestServer_test.go # REST server tests
```

## Dependencies

- **Go 1.23.8+**
- **Layer 8 Framework**: Core distributed systems framework
- **Protocol Buffers**: Google's language-neutral data serialization
- **Google UUID**: UUID generation and manipulation

Key Layer 8 dependencies:
- `github.com/saichler/layer8`: Core Layer 8 framework
- `github.com/saichler/l8types`: Type definitions and interfaces
- `github.com/saichler/l8utils`: Utility functions
- `github.com/saichler/l8services`: Service management
- `github.com/saichler/l8srlz`: Serialization utilities

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
    Authentication: false,
    CertName:       "server",  // For HTTPS
    Prefix:         "/api/",
}

// Create and start the server
srv, err := server.NewRestServer(config)
if err != nil {
    log.Fatal(err)
}

// Register with Layer 8 service manager
// ... service registration code ...

// Start the server
go srv.Start()
```

### Creating a Web Client

```go
import (
    "github.com/saichler/l8web/go/web/client"
)

// Configure the client
config := &client.RestClientConfig{
    Host:          "localhost",
    Port:          8080,
    Https:         true,
    CertFileName:  "ca.crt",
    TokenRequired: false,
    Prefix:        "/api/",
}

// Create the client
restClient, err := client.NewRestClient(config, resources)
if err != nil {
    log.Fatal(err)
}

// Make requests
response, err := restClient.GET("/endpoint", "ResponseType", "", "", nil)
```

### Using API Key Authentication

```go
import (
    "github.com/saichler/l8web/go/web/client"
)

// Configure client with API key authentication
config := &client.RestClientConfig{
    Host:  "localhost",
    Port:  8080,
    Https: true,
    AuthInfo: &client.RestAuthInfo{
        IsAPIKey: true,
        ApiUser:  "your-user-id",
        ApiKey:   "your-api-key",
    },
}

// Create the client
restClient, err := client.NewRestClient(config, resources)
if err != nil {
    log.Fatal(err)
}
```

### Using GraphQL Client

```go
import (
    "github.com/saichler/l8web/go/web/gclient"
)

// Configure GraphQL client
config := &gclient.GraphQLClientConfig{
    Host:     "localhost",
    Port:     8080,
    Https:    true,
    Endpoint: "/graphql",
    AuthInfo: &gclient.GraphQLAuthInfo{
        IsAPIKey: true,
        ApiUser:  "your-user-id",
        ApiKey:   "your-api-key",
    },
}

// Create the client
graphQLClient, err := gclient.NewGraphQLClient(config, resources)
if err != nil {
    log.Fatal(err)
}

// Execute a GraphQL query
query := `
    query GetUser($id: ID!) {
        user(id: $id) {
            name
            email
        }
    }
`
variables := map[string]interface{}{
    "id": "user123",
}
response, err := graphQLClient.Query(query, variables, "UserResponse", "user")

// Execute a GraphQL mutation
mutation := `
    mutation UpdateUser($id: ID!, $name: String!) {
        updateUser(id: $id, name: $name) {
            id
            name
        }
    }
`
response, err := graphQLClient.Mutate(mutation, variables, "UserResponse", "updateUser")
```

### Service Registration

Services can be registered with the web server through the Layer 8 framework:

```go
// Register a web service handler
webNic.Resources().Services().RegisterServiceHandlerType(&server.WebService{})

// Activate the web service
_, err = webNic.Resources().Services().Activate(
    server.ServiceTypeName, 
    ifs.WebService,
    0, 
    webNic.Resources(), 
    webNic, 
    srv,
)
```

## Configuration

### Server Configuration

- **Host**: Server bind address
- **Port**: Server port number
- **Authentication**: Enable/disable authentication
- **CertName**: Certificate name for HTTPS (generates if not exists)
- **Prefix**: URL prefix for all endpoints

### Client Configuration

- **Host**: Target server hostname
- **Port**: Target server port
- **Https**: Enable HTTPS connections
- **TokenRequired**: Require bearer token authentication
- **Token**: Authentication token
- **CertFileName**: CA certificate file for verification
- **AuthInfo**: Authentication configuration including:
  - **IsAPIKey**: Enable API key authentication
  - **ApiUser**: API user ID for X-USER-ID header
  - **ApiKey**: API key for X-API-KEY header
  - **AuthPath**: Authentication endpoint path

## Testing

The project includes comprehensive tests that demonstrate client-server interactions:

```bash
# Run tests with coverage
./test.sh
```

The test suite includes:
- REST server creation and configuration
- Service registration and discovery
- Client-server communication
- Protocol buffer serialization/deserialization
- Plugin loading and activation

## Security Features

- **TLS/HTTPS Support**: Full SSL/TLS encryption for secure communications
- **Certificate Management**: Automatic certificate generation and validation
- **Bearer Token Authentication**: Token-based authentication system
- **API Key Authentication**: Support for API key-based authentication with custom headers
- **Custom CA Support**: Support for custom certificate authorities
- **Flexible Auth Paths**: Configurable authentication endpoints and paths that bypass auth

## Integration with Layer 8

l8web is tightly integrated with the Layer 8 framework:

- **VNic Integration**: Uses Virtual Network Interfaces for communication
- **Service Discovery**: Automatic discovery of available services
- **Proximity Routing**: Leverages Layer 8's proximity-based message routing
- **Plugin System**: Dynamic loading of service implementations
- **Resource Management**: Integrated with Layer 8's resource management system

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Contributing

This project is part of the Layer 8 framework ecosystem. Please refer to the main Layer 8 project for contribution guidelines.