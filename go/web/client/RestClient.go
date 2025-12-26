/*
 * Copyright (c) 2025 Sharon Aicler (saichler@gmail.com)
 *
 * Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package client provides a REST HTTP client for communicating with Layer 8 web services.
// It supports both HTTP and HTTPS with custom CA certificates, bearer token authentication,
// API key authentication, GZIP compression, and automatic retry on timeout.
//
// Features:
//   - HTTP/HTTPS with TLS certificate verification or InsecureSkipVerify
//   - Bearer token authentication with automatic token refresh via Auth()
//   - API key authentication via custom headers (X-USER-ID, X-API-KEY)
//   - GZIP response decompression
//   - Automatic retry on timeout (up to 5 attempts with 5-second backoff)
//   - Protocol Buffer serialization via protojson
//
// Example usage:
//
//	config := &RestClientConfig{
//	    Host:  "api.example.com",
//	    Port:  443,
//	    Https: true,
//	    TokenRequired: true,
//	}
//	client, _ := NewRestClient(config, resources)
//	client.Auth("user", "pass")
//	response, _ := client.GET("/users", "UserList", "", "", nil)
package client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// RestClient is an HTTP client for communicating with Layer 8 REST services.
// It handles authentication, request building, and response parsing with
// Protocol Buffer support.
type RestClient struct {
	RestClientConfig               // Embedded configuration
	httpClient       *nethttp.Client // Underlying HTTP client with TLS config
	resources        ifs.IResources  // Layer 8 resources for type registry access
}

// RestClientConfig contains configuration options for creating a REST client.
type RestClientConfig struct {
	Host          string        // Target server hostname (e.g., "api.example.com")
	Prefix        string        // URL prefix for all requests (e.g., "/api/v1/")
	Port          int           // Target server port
	Https         bool          // Enable HTTPS connections
	TokenRequired bool          // Require bearer token for requests
	Token         string        // Current bearer token (set by Auth() or manually)
	CertFileName  string        // Path to CA certificate file for TLS verification
	AuthInfo      *RestAuthInfo // Authentication configuration
}

// RestAuthInfo contains authentication configuration for the REST client.
// Supports two modes: bearer token authentication and API key authentication.
type RestAuthInfo struct {
	NeedAuth   bool   // Enable bearer token authentication flow
	BodyType   string // Protocol Buffer type name for auth request body
	UserField  string // Field name for username in auth request
	PassField  string // Field name for password in auth request
	RespType   string // Protocol Buffer type name for auth response
	TokenField string // Field name containing token in auth response
	AuthPath   string // Endpoint path for authentication (e.g., "/auth")
	IsAPIKey   bool   // Use API key authentication instead of bearer token
	ApiUser    string // API user ID (sent as X-USER-ID header)
	ApiKey     string // API key (sent as X-API-KEY header)
}

// NewRestClient creates a new REST client with the provided configuration.
// For HTTPS connections, it configures TLS:
//   - If CertFileName is provided, it uses that CA certificate for verification
//   - Otherwise, it uses InsecureSkipVerify (suitable for self-signed certs)
//
// Returns an error if the certificate file cannot be read.
func NewRestClient(config *RestClientConfig, resources ifs.IResources) (*RestClient, error) {
	rc := &RestClient{}
	rc.CertFileName = config.CertFileName
	rc.Host = config.Host
	rc.Https = config.Https
	rc.AuthInfo = config.AuthInfo
	rc.Prefix = config.Prefix
	rc.Port = config.Port
	rc.TokenRequired = config.TokenRequired
	rc.Token = config.Token
	rc.resources = resources

	if !rc.Https {
		rc.httpClient = &nethttp.Client{}
	} else {
		if rc.CertFileName != "" {
			caCert, err := os.ReadFile(rc.CertFileName)
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			rc.httpClient = &nethttp.Client{
				Transport: &nethttp.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: caCertPool,
						//InsecureSkipVerify: true,
						ClientAuth: tls.NoClientCert,
						ServerName: rc.Host,
					},
				},
			}
		} else {
			rc.httpClient = &nethttp.Client{
				Transport: &nethttp.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
						ServerName:         rc.Host,
					},
				},
			}
		}

	}

	return rc, nil
}

// buildURL constructs the full URL for a request, combining the host, port,
// prefix, endpoint, and any query variables. The prefix is not added for
// the /auth endpoint to support authentication at different paths.
func (rc *RestClient) buildURL(end, vars string) string {
	url := bytes.Buffer{}
	url.WriteString("http")
	if rc.Https {
		url.WriteString("s")
	}
	url.WriteString("://")
	url.WriteString(rc.Host)
	url.WriteString(":")
	url.WriteString(strconv.Itoa(rc.Port))
	if rc.Prefix != "" && end != "/auth" {
		url.WriteString(rc.Prefix)
	}
	url.WriteString(end)
	url.WriteString(vars)
	fmt.Println("Client URL:", url.String())
	return url.String()
}

// request creates an HTTP request with proper headers and authentication.
// It marshals the Protocol Buffer body to JSON, sets Authorization header
// if a token is available, and adds API key headers if configured.
// Panics if TokenRequired is true but no token is available for non-auth endpoints.
func (rc *RestClient) request(method, end, vars string, pbBody proto.Message) (*nethttp.Request, error) {
	var body []byte
	var err error
	if pbBody != nil && vars == "" {
		body, err = protojson.Marshal(pbBody)
		if err != nil {
			return nil, err
		}
	}
	url := rc.buildURL(end, vars)
	request, err := nethttp.NewRequest(method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}

	if rc.TokenRequired && rc.Token == "" && rc.Https && !rc.isAuthPath(end) {
		panic("No token with secure connection!")
	}

	if rc.TokenRequired && rc.Token != "" {
		request.Header.Set("Authorization", "Bearer "+rc.Token)
	}
	request.Header.Add("content-type", "application/json")
	request.Header.Add("Accept", "application/json, text/plain, */*")
	request.Header.Add("Access-Control-Allow-Origin", "*")
	if rc.AuthInfo.IsAPIKey {
		request.Header.Add("X-USER-ID", rc.AuthInfo.ApiUser)
		request.Header.Add("X-API-KEY", rc.AuthInfo.ApiKey)
	}
	return request, nil
}

// isAuthPath checks if the endpoint is the configured authentication path.
// Used to skip token requirements for the auth endpoint itself.
func (rc *RestClient) isAuthPath(end string) bool {
	if rc.AuthInfo == nil {
		return false
	}
	if strings.HasSuffix(rc.AuthInfo.AuthPath, end) {
		return true
	}
	return false
}

// is200 checks if an HTTP status string represents a successful response (2xx).
// Parses the numeric status code from the status line (e.g., "200 OK").
func is200(status string) (bool, error) {
	index := strings.Index(status, " ")
	stat, err := strconv.Atoi(status[0:index])
	if err != nil {
		return false, err
	}
	if stat >= 200 && stat <= 299 {
		return true, nil
	}
	return false, nil
}

// isTimeout checks if an error indicates a timeout or connection issue.
// If so, it sleeps for 5 seconds before returning true to enable retry.
// Detects: "connection reset by peer", "timeout", "connection timed out".
func isTimeout(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection timed out") {
		time.Sleep(time.Second * 5)
		return true
	}
	return false
}

// Auth performs authentication against the configured AuthPath endpoint.
// It creates a credentials message using reflection based on AuthInfo configuration,
// sends it to the server, and extracts the bearer token from the response.
// The token is stored in rc.Token for use in subsequent requests.
//
// Requires AuthInfo to be configured with: BodyType, UserField, PassField,
// RespType, TokenField, and AuthPath.
//
// Returns nil if NeedAuth is false or if authentication succeeds.
func (rc *RestClient) Auth(user, pass string) error {
	if rc.AuthInfo == nil || !rc.AuthInfo.NeedAuth {
		return nil
	}

	info, err := rc.resources.Registry().Info(rc.AuthInfo.BodyType)
	if err != nil {
		return err
	}

	creds, _ := info.NewInstance()
	credsVal := reflect.ValueOf(creds).Elem()
	if !credsVal.FieldByName(rc.AuthInfo.UserField).CanSet() || !credsVal.FieldByName(rc.AuthInfo.PassField).CanSet() {
		return errors.New("invalid credential field names")
	}

	credsVal.FieldByName(rc.AuthInfo.UserField).Set(reflect.ValueOf(user))
	credsVal.FieldByName(rc.AuthInfo.PassField).Set(reflect.ValueOf(pass))

	token, err := rc.Do("POST", rc.AuthInfo.AuthPath, rc.AuthInfo.RespType, "", "", creds.(proto.Message), 5)
	if err != nil {
		return err
	}

	tokenVal := reflect.ValueOf(token).Elem()
	if !tokenVal.FieldByName(rc.AuthInfo.TokenField).CanSet() {
		return errors.New("invalid token field name")
	}

	t, ok := tokenVal.FieldByName(rc.AuthInfo.TokenField).Interface().(string)
	if !ok {
		return errors.New("invalid token field value, should be string")
	}

	rc.Token = t
	return nil
}

// Do executes an HTTP request and returns the response as a Protocol Buffer message.
//
// Parameters:
//   - method: HTTP method (GET, POST, PUT, PATCH, DELETE)
//   - end: Endpoint path (e.g., "/users")
//   - responseType: Protocol Buffer type name for deserializing response
//   - responseAttribute: Optional attribute name to wrap response JSON (for nested responses)
//   - vars: Query string to append to URL
//   - pbBody: Request body as Protocol Buffer (marshaled to JSON)
//   - tryCount: Current retry attempt (starts at 1, max 5)
//
// Handles GZIP response decompression automatically. Retries on timeout errors
// up to 5 times with 5-second backoff. Returns error for non-2xx responses.
func (rc *RestClient) Do(method, end, responseType, responseAttribute, vars string, pbBody proto.Message, tryCount int) (proto.Message, error) {

	request, err := rc.request(method, end, vars, pbBody)
	if err != nil {
		return nil, err
	}

	//Execute the request
	response, err := rc.httpClient.Do(request)
	if err != nil {
		if isTimeout(err) {
			if tryCount <= 5 {
				return rc.Do(method, end, responseType, responseAttribute, vars, pbBody, tryCount+1)
			}
		}
		return nil, err
	}

	var jsonBytes []byte

	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ := gzip.NewReader(response.Body)
		jsonBytes, _ = io.ReadAll(reader)
		defer reader.Close()
	default:
		jsonBytes, _ = io.ReadAll(response.Body)
	}
	ok, err := is200(response.Status)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New(method + " failed with status " + response.Status + ":" + string(jsonBytes))
	}

	if responseType == "" {
		return nil, err
	}

	info, err := rc.resources.Registry().Info(responseType)
	if err != nil {
		return nil, err
	}
	_interface, err := info.NewInstance()
	if err != nil {
		return nil, err
	}

	responsePb := _interface.(proto.Message)
	if responseAttribute != "" {
		buff := bytes.Buffer{}
		buff.WriteString("{\"")
		buff.WriteString(responseAttribute)
		buff.WriteString("\": ")
		buff.Write(jsonBytes)
		buff.WriteString("}")
		jsonBytes = buff.Bytes()
	}
	err = protojson.Unmarshal(jsonBytes, responsePb)
	if err != nil {
		fmt.Println(string(jsonBytes))
	}
	return responsePb, err
}

// GET performs an HTTP GET request. Convenience wrapper for Do().
// For GET requests, pbBody can be used to send a request body, though this
// is non-standard HTTP. Use vars for query parameters instead.
func (rc *RestClient) GET(end, responseType, responseAttribute, vars string, pbBody proto.Message) (proto.Message, error) {
	return rc.Do("GET", end, responseType, responseAttribute, vars, pbBody, 1)
}

// POST performs an HTTP POST request. Convenience wrapper for Do().
// Used for creating new resources. The pbBody is serialized as JSON in the request body.
func (rc *RestClient) POST(end, responseType, responseAttribute, vars string, pbBody proto.Message) (proto.Message, error) {
	return rc.Do("POST", end, responseType, responseAttribute, vars, pbBody, 1)
}

// PUT performs an HTTP PUT request. Convenience wrapper for Do().
// Used for full resource replacement. The pbBody is serialized as JSON in the request body.
func (rc *RestClient) PUT(end, responseType, responseAttribute, vars string, pbBody proto.Message) (proto.Message, error) {
	return rc.Do("PUT", end, responseType, responseAttribute, vars, pbBody, 1)
}

// PATCH performs an HTTP PATCH request. Convenience wrapper for Do().
// Used for partial resource updates. The pbBody is serialized as JSON in the request body.
func (rc *RestClient) PATCH(end, responseType, responseAttribute, vars string, pbBody proto.Message) (proto.Message, error) {
	return rc.Do("PATCH", end, responseType, responseAttribute, vars, pbBody, 1)
}

// DELETE performs an HTTP DELETE request. Convenience wrapper for Do().
// Used for resource deletion. pbBody is typically nil for DELETE requests.
func (rc *RestClient) DELETE(end, responseType, responseAttribute, vars string, pbBody proto.Message) (proto.Message, error) {
	return rc.Do("DELETE", end, responseType, responseAttribute, vars, pbBody, 1)
}
