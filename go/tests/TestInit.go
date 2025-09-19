package tests

import (
	. "github.com/saichler/l8test/go/infra/t_resources"
	. "github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8bus/go/overlay/protocol"
)

//var topo *TestTopology

func init() {
	Log.SetLogLevel(Trace_Level)
}

func setup() {
	protocol.Discovery_Enabled = false
	setupTopology()
}

func tear() {
	shutdownTopology()
}

func reset(name string) {
	Log.Info("*** ", name, " end ***")
	//topo.ResetHandlers()
}

func setupTopology() {
	//topo = NewTestTopology(4, []int{20000, 30000, 40000}, Trace_Level)
}

func shutdownTopology() {
	//topo.Shutdown()
}
