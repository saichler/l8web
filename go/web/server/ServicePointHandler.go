package server

import (
	"errors"
	"github.com/saichler/types/go/common"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"reflect"
)

type ServicePointHandler struct {
	serviceName   string
	serviceArea   uint16
	methodToProto map[string]proto.Message
	vnic          common.IVirtualNetworkInterface
}

func NewServicePointHandler(serviceName string, serviceArea uint16, vnic common.IVirtualNetworkInterface) *ServicePointHandler {
	sph := &ServicePointHandler{}
	sph.serviceName = serviceName
	sph.serviceArea = serviceArea
	sph.vnic = vnic
	sph.methodToProto = make(map[string]proto.Message)
	return sph
}

func (this *ServicePointHandler) AddMethodType(method string, pb proto.Message) {
	this.vnic.Resources().Registry().Register(pb)
	this.methodToProto[method] = pb
}

func (this *ServicePointHandler) serveHttp(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	pb, err := this.newPb(method)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	err = protojson.Unmarshal(data, pb)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	resp := this.vnic.SingleRequest(this.serviceName, this.serviceArea, methodToAction(method), pb)
	if resp.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(resp.Error().Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	elem, ok := resp.Element().(proto.Message)
	if ok {
		j, e := protojson.Marshal(elem)
		if e != nil {
			w.Write([]byte(e.Error()))
		} else {
			w.Write(j)
		}
	}
}

func (this *ServicePointHandler) newPb(method string) (proto.Message, error) {
	pb, ok := this.methodToProto[method]
	if !ok {
		return nil, errors.New("Method does not have any protobuf registered")
	}
	return reflect.New(reflect.ValueOf(pb).Elem().Type()).Interface().(proto.Message), nil
}

func methodToAction(method string) common.Action {
	switch method {
	case http.MethodPost:
		return common.POST
	case http.MethodGet:
		return common.GET
	case http.MethodDelete:
		return common.DELETE
	case http.MethodPut:
		return common.PUT
	case http.MethodPatch:
		return common.PATCH
	}
	return common.GET
}
