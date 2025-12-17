package server

import (
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
	webService  ifs.IWebService
	authEnabled bool
}

type ServiceAction struct {
	body proto.Message
	resp proto.Message
}

var Timeout = 30
var Target = ""
var Method = ifs.M_Leader

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
		aToken := ""
		if !ok && (id == "Token Setup TFA" || id == "Token Need TFA Verification") {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(id))
			return
		}
		if !ok {
			//This might be a request for the adjacent
			if len(bearer) > 7 && (strings.HasPrefix(bearer, "bearer") || strings.HasPrefix(bearer, "Bearer")) {
				bearer = bearer[7:]
			}

			mtx.Lock()
			aToken, ok = adjacentTokens[bearer]
			mtx.Unlock()

			if aToken != "" {
				id, ok = this.vnic.Resources().Security().ValidateToken(aToken)
			}
		}
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		aaaid = id
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Failed to read body for method " + r.Method + "\n"))
		w.Write([]byte(err.Error()))
		fmt.Println("Failed to read body for method " + r.Method + "\n")
		return
	}

	if strings.ToLower(r.Method) == "get" && (data == nil || len(data) == 0) {
		qData := r.URL.Query().Get("body")
		data = []byte(qData)
	}

	action := methodToAction(r.Method, nil)
	body, _, err := this.webService.Protos(string(data), action)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot find pb for method " + r.Method + "\n"))
		w.Write([]byte(err.Error()))
		fmt.Println("Cannot find pb for method " + r.Method + "\n")
		return
	}

	action = methodToAction(r.Method, body)
	var elems ifs.IElements

	dest := this.vnic.Resources().SysConfig().RemoteUuid
	if this.serviceName == health.ServiceName {
		this.vnic.Resources().Logger().Info("Sending to vnet")
		elems = this.vnic.Request(dest, this.serviceName, this.serviceArea, action, body, Timeout)
	} else {
		if Target != "" {
			elems = this.vnic.Request(Target, this.serviceName, this.serviceArea, action, body, Timeout, aaaid)
		} else {
			if Method == ifs.M_Leader {
				elems = this.vnic.LeaderRequest(this.serviceName, this.serviceArea, action, body, Timeout, aaaid)
			} else if Method == ifs.M_Local {
				elems = this.vnic.LocalRequest(this.serviceName, this.serviceArea, action, body, Timeout, aaaid)
			} else {
				elems = this.vnic.ProximityRequest(this.serviceName, this.serviceArea, action, body, Timeout, aaaid)
			}
		}
	}

	if elems.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error from single request:"))
		w.Write([]byte(elems.Error().Error()))
		fmt.Println("Error from single request:")
		fmt.Println(elems.Error().Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	response, e := elems.AsList(this.vnic.Resources().Registry())
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
