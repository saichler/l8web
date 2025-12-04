package server

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func bodyToProto(w http.ResponseWriter, r *http.Request, method string, body proto.Message) bool {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read body for method " + method + "\n"))
		w.Write([]byte(err.Error()))
		fmt.Println("Failed to read body for method " + method + "\n")
		return false
	}

	if data != nil && len(data) > 0 {
		err = protojson.Unmarshal(data, body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to unmarshal body for method " + method + " element " + reflect.ValueOf(body).Elem().Type().Name() + "\n"))
			w.Write([]byte("body for method " + method + string(data) + "\n"))
			w.Write([]byte(err.Error()))
			fmt.Println("Failed to unmarshal body for method " + method + " element " + reflect.ValueOf(body).Elem().Type().Name() + "\n")
			fmt.Println("body for method " + method + string(data) + "\n")
			return false
		}
	}
	return true
}
