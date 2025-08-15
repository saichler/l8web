package server

import (
	"errors"
	"github.com/saichler/l8types/go/ifs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"reflect"
)

type ServiceHandler struct {
	serviceName string
	serviceArea byte
	vnic        ifs.IVNic
	method2Body map[string]proto.Message
	method2Resp map[string]proto.Message
}

func (this *ServiceHandler) addEndPoint(method, body, resp string) {
	if body != "" {
		info, err := this.vnic.Resources().Registry().Info(body)
		if err != nil {
			this.vnic.Resources().Logger().Error(err)
			return
		}
		ins, err := info.NewInstance()
		if err != nil {
			this.vnic.Resources().Logger().Error(err)
			return
		}
		this.vnic.Resources().Registry().Register(ins)
		this.method2Body[method] = ins.(proto.Message)
	}
	if resp != "" {
		info, err := this.vnic.Resources().Registry().Info(resp)
		if err != nil {
			this.vnic.Resources().Logger().Error(err)
			return
		}
		ins, err := info.NewInstance()
		if err != nil {
			this.vnic.Resources().Logger().Error(err)
			return
		}
		this.vnic.Resources().Registry().Register(ins)
		this.method2Resp[method] = ins.(proto.Message)
	}
}

func (this *ServiceHandler) ServiceName() string {
	return this.serviceName
}

func (this *ServiceHandler) ServiceArea() byte {
	return this.serviceArea
}

func (this *ServiceHandler) serveHttp(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	body, err := this.newBody(method)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot find pb for method " + method + "\n"))
		w.Write([]byte(err.Error()))
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read body for method " + method + "\n"))
		w.Write([]byte(err.Error()))
		return
	}
	err = protojson.Unmarshal(data, body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to unmarshal body for method " + method + " element " + reflect.ValueOf(body).Elem().Type().Name() + "\n"))
		w.Write([]byte("body for method " + method + string(data) + "\n"))
		w.Write([]byte(err.Error()))
		return
	}
	resp := this.vnic.ProximityRequest(this.serviceName, this.serviceArea, methodToAction(method), body)
	if resp.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error from single request:\n"))
		w.Write([]byte(resp.Error().Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{list: ["))
	first := true
	for _, ei := range resp.Elements() {
		elem, ok := ei.(proto.Message)
		if ok {
			if !first {
				w.Write([]byte(","))
			}
			first = false
			j, e := protojson.Marshal(elem)
			if e != nil {
				w.Write([]byte("Erorr marshaling:" + reflect.ValueOf(elem).Elem().Type().Name()))
				w.Write([]byte(e.Error()))
			} else {
				w.Write(j)
			}
		}
	}
	w.Write([]byte("]}"))
}

func (this *ServiceHandler) newBody(method string) (proto.Message, error) {
	pb, ok := this.method2Body[method]
	if !ok {
		return nil, errors.New("Method does not have any protobuf registered")
	}
	return reflect.New(reflect.ValueOf(pb).Elem().Type()).Interface().(proto.Message), nil
}

func (this *ServiceHandler) newResp(method string) (proto.Message, error) {
	pb, ok := this.method2Resp[method]
	if !ok {
		return nil, errors.New("Method does not have any protobuf registered")
	}
	return reflect.New(reflect.ValueOf(pb).Elem().Type()).Interface().(proto.Message), nil
}

func methodToAction(method string) ifs.Action {
	switch method {
	case http.MethodPost:
		return ifs.POST
	case http.MethodGet:
		return ifs.GET
	case http.MethodDelete:
		return ifs.DELETE
	case http.MethodPut:
		return ifs.PUT
	case http.MethodPatch:
		return ifs.PATCH
	}
	return ifs.GET
}
