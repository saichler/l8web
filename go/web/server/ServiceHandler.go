package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/saichler/l8bus/go/overlay/health"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ServiceHandler struct {
	serviceName string
	serviceArea byte
	vnic        ifs.IVNic
	method2Body map[string]proto.Message
	method2Resp map[string]proto.Message
	authEnabled bool
}

var Timeout = 30
var Target = ""
var Method = ifs.M_Leader

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
	aaaid := ""
	if this.authEnabled {
		bearer := r.Header.Get("Authorization")
		if bearer == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		id, ok := this.vnic.Resources().Security().ValidateToken(bearer)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		aaaid = id
	}

	method := r.Method
	body, err := this.newBody(method)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot find pb for method " + method + "\n"))
		w.Write([]byte(err.Error()))
		fmt.Println("Cannot find pb for method " + method + "\n")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read body for method " + method + "\n"))
		w.Write([]byte(err.Error()))
		fmt.Println("Failed to read body for method " + method + "\n")
		return
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
			return
		}
	}
	if strings.ToLower(method) == "get" {
		qData := r.URL.Query().Get("body")
		if qData != "" {
			err = protojson.Unmarshal([]byte(qData), body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to unmarshal query body for method " + method + " element " + reflect.ValueOf(body).Elem().Type().Name() + "\n"))
				w.Write([]byte("body for method " + method + string(data) + "\n"))
				w.Write([]byte(err.Error()))
				fmt.Println("Failed to unmarshal query body for method " + method + " element " + reflect.ValueOf(body).Elem().Type().Name() + "\n")
				fmt.Println("body for method " + method + string(data) + "\n")
				return
			}
		}
	}

	var resp ifs.IElements
	if this.serviceName == health.ServiceName {
		this.vnic.Resources().Logger().Info("Sending to vnet")
		resp = this.vnic.Request(this.vnic.Resources().SysConfig().RemoteUuid, this.serviceName, this.serviceArea, methodToAction(method, body), body, Timeout)
	} else {
		if Target != "" {
			resp = this.vnic.Request(Target, this.serviceName, this.serviceArea, methodToAction(method, body), body, Timeout, aaaid)
		} else {
			if Method == ifs.M_Leader {
				resp = this.vnic.LeaderRequest(this.serviceName, this.serviceArea, methodToAction(method, body), body, Timeout, aaaid)
			} else if Method == ifs.M_Local {
				resp = this.vnic.LocalRequest(this.serviceName, this.serviceArea, methodToAction(method, body), body, Timeout, aaaid)
			} else {
				resp = this.vnic.ProximityRequest(this.serviceName, this.serviceArea, methodToAction(method, body), body, Timeout, aaaid)
			}
		}
	}

	if resp.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error from single request:"))
		w.Write([]byte(resp.Error().Error()))
		fmt.Println("Error from single request:")
		fmt.Println(resp.Error().Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	response, e := resp.AsList(this.vnic.Resources().Registry())
	if e != nil {
		w.Write([]byte("{}"))
		/*
			w.Write([]byte("Erorr as list:"))
			w.Write([]byte(e.Error()))
		*/
		return
	}

	marshalOptions := protojson.MarshalOptions{
		UseEnumNumbers: true,
	}
	j, e := marshalOptions.Marshal(response.(proto.Message))
	if e != nil {
		w.Write([]byte("Erorr marshaling:" + reflect.ValueOf(response).Elem().Type().Name()))
		w.Write([]byte(e.Error()))
		fmt.Println("Erorr marshaling:" + reflect.ValueOf(response).Elem().Type().Name())
	} else {
		w.Write(j)
	}
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

func methodToAction(method string, body proto.Message) ifs.Action {
	isMapReduce := false
	q, ok := body.(*l8api.L8Query)
	if ok {
		if strings.Contains(strings.ToLower(q.Text), "mapreduce") {
			isMapReduce = true
		}
	}
	switch method {
	case http.MethodPost:
		if isMapReduce {
			return ifs.MapR_POST
		}
		return ifs.POST
	case http.MethodGet:
		if isMapReduce {
			return ifs.MapR_GET
		}
		return ifs.GET
	case http.MethodDelete:
		if isMapReduce {
			return ifs.MapR_DELETE
		}
		return ifs.DELETE
	case http.MethodPut:
		if isMapReduce {
			return ifs.MapR_PUT
		}
		return ifs.PUT
	case http.MethodPatch:
		if isMapReduce {
			return ifs.MapR_PATCH
		}
		return ifs.PATCH
	}
	return ifs.GET
}
