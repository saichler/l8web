package server

import (
	"net/http"

	"github.com/saichler/l8types/go/types/l8api"
	"google.golang.org/protobuf/encoding/protojson"
)

func (this *WebService) TFASetup(w http.ResponseWriter, r *http.Request) {
	body := &l8api.L8TFASetup{}
	if !bodyToProto(w, r, "POST", body) {
		return
	}

	secret, qr, err := this.vnic.Resources().Security().TFASetup(body.UserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	resp := &l8api.L8TFASetupR{}
	resp.Secret = secret
	resp.Qr = qr
	respData, err := protojson.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

func (this *WebService) TFAVerify(w http.ResponseWriter, r *http.Request) {
	body := &l8api.L8TFAVerify{}
	if !bodyToProto(w, r, "POST", body) {
		return
	}
	err := this.vnic.Resources().Security().TFAVerify(body.UserId, body.Code, body.Bearer, this.vnic)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	resp := &l8api.L8TFAVerifyR{}
	resp.Ok = true
	respData, err := protojson.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}
