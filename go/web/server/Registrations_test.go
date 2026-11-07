package server

import (
	"testing"

	"github.com/saichler/l8types/go/types/l8api"
)

func regQuery(action l8api.L8Register, id string) *l8api.L8Query {
	return &l8api.L8Query{Text: "select * from Employee", Register: action, RegisterId: id}
}

func TestRegistrationsTrackAndTake(t *testing.T) {
	r := newRegistrations()
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "table"))
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "chart"))
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "table")) // re-fetch, same component
	r.track("aaa-2", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "table"))

	// Not tracked: a plain query, a query with no register id, no user.
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_INVALID, "other"))
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, ""))
	r.track("", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "x"))

	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_UNREGISTER, "chart"))

	got := r.take("aaa-1")
	if len(got) != 1 || got[0].registerId != "table" || got[0].serviceName != "Employee" ||
		got[0].serviceArea != 30 || got[0].text != "select * from Employee" {
		t.Fatalf("aaa-1: expected only the 'table' registration, got %+v", got)
	}
	if len(r.take("aaa-1")) != 0 {
		t.Fatal("take must remove what it returns")
	}
	if len(r.take("aaa-2")) != 1 {
		t.Fatal("another user's registrations must be untouched")
	}
}

func TestRegistrationsKeyIncludesService(t *testing.T) {
	r := newRegistrations()
	// The dashboard registers one query per KPI model under one register id.
	r.track("aaa-1", "Employee", 30, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "dashboard"))
	r.track("aaa-1", "SalesOrder", 60, regQuery(l8api.L8Register_L8_REGISTER_REGISTER, "dashboard"))
	if n := len(r.take("aaa-1")); n != 2 {
		t.Fatalf("expected one registration per service under the same id, got %d", n)
	}
}
