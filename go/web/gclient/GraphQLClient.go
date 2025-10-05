package gclient

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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

type GraphQLClient struct {
	GraphQLClientConfig
	httpClient *nethttp.Client
	resources  ifs.IResources
}

type GraphQLClientConfig struct {
	Host          string
	Prefix        string
	Port          int
	Https         bool
	TokenRequired bool
	Token         string
	CertFileName  string
	AuthInfo      *GraphQLAuthInfo
	Endpoint      string // GraphQL endpoint path (default: /graphql)
}

type GraphQLAuthInfo struct {
	NeedAuth   bool
	BodyType   string
	UserField  string
	PassField  string
	RespType   string
	TokenField string
	AuthPath   string
	IsAPIKey   bool
	ApiUser    string
	ApiKey     string
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage          `json:"data,omitempty"`
	Errors []GraphQLError           `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func NewGraphQLClient(config *GraphQLClientConfig, resources ifs.IResources) (*GraphQLClient, error) {
	gc := &GraphQLClient{}
	gc.CertFileName = config.CertFileName
	gc.Host = config.Host
	gc.Https = config.Https
	gc.AuthInfo = config.AuthInfo
	gc.Prefix = config.Prefix
	gc.Port = config.Port
	gc.TokenRequired = config.TokenRequired
	gc.Token = config.Token
	gc.resources = resources
	gc.Endpoint = config.Endpoint
	if gc.Endpoint == "" {
		gc.Endpoint = "/graphql"
	}

	if !gc.Https {
		gc.httpClient = &nethttp.Client{}
	} else {
		if gc.CertFileName != "" {
			caCert, err := os.ReadFile(gc.CertFileName)
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			gc.httpClient = &nethttp.Client{
				Transport: &nethttp.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs:    caCertPool,
						ClientAuth: tls.NoClientCert,
						ServerName: gc.Host,
					},
				},
			}
		} else {
			gc.httpClient = &nethttp.Client{
				Transport: &nethttp.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
						ServerName:         gc.Host,
					},
				},
			}
		}
	}

	return gc, nil
}

func (gc *GraphQLClient) buildURL(end string) string {
	url := bytes.Buffer{}
	url.WriteString("http")
	if gc.Https {
		url.WriteString("s")
	}
	url.WriteString("://")
	url.WriteString(gc.Host)
	url.WriteString(":")
	url.WriteString(strconv.Itoa(gc.Port))
	if gc.Prefix != "" && end != "/auth" {
		url.WriteString(gc.Prefix)
	}
	url.WriteString(end)
	fmt.Println("GraphQL Client URL:", url.String())
	return url.String()
}

func (gc *GraphQLClient) request(end string, gqlRequest *GraphQLRequest) (*nethttp.Request, error) {
	body, err := json.Marshal(gqlRequest)
	if err != nil {
		return nil, err
	}

	url := gc.buildURL(end)
	request, err := nethttp.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if gc.TokenRequired && gc.Token == "" && gc.Https && !gc.isAuthPath(end) {
		panic("No token with secure connection!")
	}

	if gc.TokenRequired && gc.Token != "" {
		request.Header.Set("Authorization", "Bearer "+gc.Token)
	}
	request.Header.Add("content-type", "application/json")
	request.Header.Add("Accept", "application/json, text/plain, */*")
	request.Header.Add("Access-Control-Allow-Origin", "*")
	if gc.AuthInfo != nil && gc.AuthInfo.IsAPIKey {
		request.Header.Add("X-USER-ID", gc.AuthInfo.ApiUser)
		request.Header.Add("X-API-KEY", gc.AuthInfo.ApiKey)
	}
	return request, nil
}

func (gc *GraphQLClient) isAuthPath(end string) bool {
	if gc.AuthInfo == nil {
		return false
	}
	if strings.HasSuffix(gc.AuthInfo.AuthPath, end) {
		return true
	}
	return false
}

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

func isTimeout(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection timed out") {
		time.Sleep(time.Second * 5)
		return true
	}
	return false
}

func (gc *GraphQLClient) Auth(user, pass string) error {
	if gc.AuthInfo == nil || !gc.AuthInfo.NeedAuth {
		return nil
	}

	info, err := gc.resources.Registry().Info(gc.AuthInfo.BodyType)
	if err != nil {
		return err
	}

	creds, _ := info.NewInstance()
	credsVal := reflect.ValueOf(creds).Elem()
	if !credsVal.FieldByName(gc.AuthInfo.UserField).CanSet() || !credsVal.FieldByName(gc.AuthInfo.PassField).CanSet() {
		return errors.New("invalid credential field names")
	}

	credsVal.FieldByName(gc.AuthInfo.UserField).Set(reflect.ValueOf(user))
	credsVal.FieldByName(gc.AuthInfo.PassField).Set(reflect.ValueOf(pass))

	// For GraphQL auth, we need to construct a mutation or query
	// This is a simplified version - you may need to customize based on your auth schema
	authQuery := fmt.Sprintf(`mutation { login(input: { %s: "%s", %s: "%s" }) { %s } }`,
		strings.ToLower(gc.AuthInfo.UserField[:1])+gc.AuthInfo.UserField[1:],
		user,
		strings.ToLower(gc.AuthInfo.PassField[:1])+gc.AuthInfo.PassField[1:],
		pass,
		strings.ToLower(gc.AuthInfo.TokenField[:1])+gc.AuthInfo.TokenField[1:])

	token, err := gc.Execute(authQuery, nil, gc.AuthInfo.RespType, gc.AuthInfo.TokenField, 5)
	if err != nil {
		return err
	}

	tokenVal := reflect.ValueOf(token).Elem()
	if !tokenVal.FieldByName(gc.AuthInfo.TokenField).CanSet() {
		return errors.New("invalid token field name")
	}

	t, ok := tokenVal.FieldByName(gc.AuthInfo.TokenField).Interface().(string)
	if !ok {
		return errors.New("invalid token field value, should be string")
	}

	gc.Token = t
	return nil
}

func (gc *GraphQLClient) Execute(query string, variables map[string]interface{}, responseType, responseAttribute string, tryCount int) (proto.Message, error) {
	gqlRequest := &GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	request, err := gc.request(gc.Endpoint, gqlRequest)
	if err != nil {
		return nil, err
	}

	// Execute the request
	response, err := gc.httpClient.Do(request)
	if err != nil {
		if isTimeout(err) {
			if tryCount <= 5 {
				return gc.Execute(query, variables, responseType, responseAttribute, tryCount+1)
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
		return nil, errors.New("GraphQL request failed with status " + response.Status + ":" + string(jsonBytes))
	}

	// Parse GraphQL response
	var gqlResponse GraphQLResponse
	err = json.Unmarshal(jsonBytes, &gqlResponse)
	if err != nil {
		return nil, err
	}

	// Check for GraphQL errors
	if len(gqlResponse.Errors) > 0 {
		errMsg := "GraphQL errors: "
		for i, gqlErr := range gqlResponse.Errors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += gqlErr.Message
		}
		return nil, errors.New(errMsg)
	}

	if responseType == "" {
		return nil, nil
	}

	info, err := gc.resources.Registry().Info(responseType)
	if err != nil {
		return nil, err
	}
	_interface, err := info.NewInstance()
	if err != nil {
		return nil, err
	}

	responsePb := _interface.(proto.Message)

	// Extract the data field
	dataBytes := gqlResponse.Data
	if responseAttribute != "" {
		// Extract nested field from data
		var dataMap map[string]json.RawMessage
		err = json.Unmarshal(dataBytes, &dataMap)
		if err != nil {
			return nil, err
		}
		if attrData, ok := dataMap[responseAttribute]; ok {
			dataBytes = attrData
		} else {
			return nil, errors.New("response attribute '" + responseAttribute + "' not found in GraphQL response")
		}
	}

	err = protojson.Unmarshal(dataBytes, responsePb)
	if err != nil {
		fmt.Println(string(dataBytes))
	}
	return responsePb, err
}

func (gc *GraphQLClient) Query(query string, variables map[string]interface{}, responseType, responseAttribute string) (proto.Message, error) {
	return gc.Execute(query, variables, responseType, responseAttribute, 1)
}

func (gc *GraphQLClient) Mutate(mutation string, variables map[string]interface{}, responseType, responseAttribute string) (proto.Message, error) {
	return gc.Execute(mutation, variables, responseType, responseAttribute, 1)
}
