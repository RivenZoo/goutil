package zk

import (
	"testing"
	"time"
)

var (
	testServers = []string{"127.0.0.1:2181"}
)

func TestZkCliConnect(t *testing.T) {
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	err := cli.Connect()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestCreatePersistNode(t *testing.T) {
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	persistNode := "/test/persist"

	err := cli.CreatePersistNode(persistNode)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Second)

	ok, err := cli.NodeExist(persistNode)
	if err != nil || !ok {
		t.Log(err)
		t.Fail()
	}

	cli.Close()

	ok, err = cli.NodeExist(persistNode)
	if err != nil || !ok {
		t.Log(err)
		t.Fail()
	}

	err = cli.DeleteNode(persistNode)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Second)

	// after delete
	ok, err = cli.NodeExist(persistNode)
	if err != nil || ok {
		t.Log(err)
		t.Fail()
	}
}

func TestCreateEphemeralNode(t *testing.T) {
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	ephemeralNode := "/test/ephem"
	err := cli.CreateEphemeralNode(ephemeralNode)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Second)

	ok, err := cli.NodeExist(ephemeralNode)
	if err != nil || !ok {
		t.Log(err)
		t.Fail()
	}

	cli.Close()

	// after disconnect ephem node should disappear
	ok, err = cli.NodeExist(ephemeralNode)
	if err != nil || ok {
		t.Log(err)
		t.Fail()
	}
}

func findString(str []string, key string) bool {
	for _, s := range str {
		if s == key {
			return true
		}
	}
	return false
}

func TestGetChildren(t *testing.T) {
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	ephemeralNode := "/test/ephem"
	err := cli.CreateEphemeralNode(ephemeralNode)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Second)

	children, err := cli.GetChildren("/test")
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	ok := findString(children, "ephem")
	if !ok {
		t.Fail()
	}
	t.Log(children)
}

func TestWatchDir(t *testing.T) {
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	ephemeralNode := "/test/ephem"
	err := cli.CreateEphemeralNode(ephemeralNode)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	time.Sleep(time.Second)

	w := cli.WatchDir("/test")
	if w == nil {
		t.Fail()
	}
	children := <-w.Children
	t.Log(children)
	ok := findString(children, "ephem")
	if !ok {
		t.Fail()
	}
	err = cli.DeleteNode(ephemeralNode)
	if err != nil {
		t.Fail()
	}
	children = <-w.Children
	if len(children) != 0 {
		t.Fail()
	}
}
