package zk
import (
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	spath := "/test"
	r := NewZKRegister(testServers, time.Second*5, spath)
	services := []string{"service1", "service2"}
	for _, serv := range services {
		err := r.Register(serv)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}
	}
	nodes, err := r.cli.GetChildren(spath)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(nodes) != len(services) {
		t.FailNow()
	}
	t.Log(nodes)
}