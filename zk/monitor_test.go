package zk

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testService struct {
	count int32
}
type testConn int32

func (t *testService) InitCli(addr string, arg interface{}) ServiceCli {
	return &testService{}
}
func (t *testService) GetConn() ServiceConn {
	n := atomic.LoadInt32(&t.count)
	atomic.AddInt32(&t.count, 1)
	return testConn(n)
}
func (t *testService) Status() string {
	return fmt.Sprint(t.count)
}
func (t *testService) Close()     {}
func (t *testService) OnDisable() {}
func (t *testService) OnEnable()  {}

func checkAddrs(addrs []string, zkm *ZKMonitor, t *testing.T) {
	for _, addr := range addrs {
		idx, ok := zkm.servIdx[addr]
		if !ok {
			t.FailNow()
		}
		if !zkm.servCli[idx].isValid() {
			t.FailNow()
		}
	}
	n := 0
	for i := range zkm.servCli {
		if zkm.servCli[i].isValid() {
			n++
		}
	}
	if n != len(addrs) {
		t.FailNow()
	}
}

func TestUpdateCli(t *testing.T) {
	zkm := &ZKMonitor{
		servIdx:     make(map[string]int),
		servCli:     make([]monitorServCli, 0),
		servLock:    &sync.RWMutex{},
		addrMonitor: nil,
		serv:        &testService{},
		servArg:     nil,
	}
	addrs := []string{"addr1", "addr2", "addr3"}
	zkm.updateServCli(addrs)
	t.Log(zkm.Status())
	checkAddrs(addrs, zkm, t)

	deleted := addrs[len(addrs)-1]
	addrs = addrs[0 : len(addrs)-1]
	zkm.updateServCli(addrs)
	t.Log(zkm.Status())
	checkAddrs(addrs, zkm, t)

	idx, ok := zkm.servIdx[deleted]
	if !ok {
		t.FailNow()
	}
	if zkm.servCli[idx].isValid() {
		t.FailNow()
	}

	zkm.updateServCli(nil)
	t.Log(zkm.Status())
	zkm.updateServCli(addrs)
	t.Log(zkm.Status())
}

func TestAdjustCli(t *testing.T) {
	zkm := &ZKMonitor{
		servIdx:     make(map[string]int),
		servCli:     make([]monitorServCli, 0),
		servLock:    &sync.RWMutex{},
		addrMonitor: nil,
		serv:        &testService{},
		servArg:     nil,
	}
	addrs := []string{"addr1", "addr2", "addr3", "addr4", "addr5", "addr8"}
	zkm.updateServCli(addrs)
	newAddrs := make([]string, 0)
	newAddrs = append(newAddrs, addrs[0], addrs[2], addrs[5])
	zkm.updateServCli(newAddrs)
	t.Log(zkm.Status())

	zkm.adjustServCli()
	t.Log(zkm.Status())
	checkAddrs(newAddrs, zkm, t)
}

func TestZKMonitor(t *testing.T) {
	mPath := "/test"
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	cli.CreatePersistNode(mPath)
	defer cli.DeleteNode(mPath)

	fmt.Println("===============create watch dir===============")
	time.Sleep(2 * time.Second)

	zkm := NewZKMonitor(testServers, time.Second*5, &testService{}, nil, mPath)
	zkm.Run()
	defer zkm.Close()

	t.Log(zkm.Status())

	rtGetter := zkm.RoundTripGetter()
	conn := rtGetter.GetConn()
	if conn != nil {
		t.Fail()
	}
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6379")
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6380")

	fmt.Println("===============create addr node===============")
	time.Sleep(2 * time.Second)

	conn = rtGetter.GetConn()
	if conn == nil {
		t.Fail()
	}
	t.Log(zkm.Status())
}

func TestRoundTripGet(t *testing.T) {
	mPath := "/test"
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	cli.CreatePersistNode(mPath)
	defer cli.DeleteNode(mPath)

	fmt.Println("===============create watch dir===============")
	time.Sleep(2 * time.Second)

	zkm := NewZKMonitor(testServers, time.Second*5, &testService{}, nil, mPath)
	zkm.Run()
	defer zkm.Close()

	t.Log(zkm.Status())

	rtGetter := zkm.RoundTripGetter()
	conn := rtGetter.GetConn()
	if conn != nil {
		t.Fail()
	}
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6379")
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6380")

	fmt.Println("===============create addr node===============")
	time.Sleep(2 * time.Second)

	trace := make([]int32, 0)
	for i := 0; i < 10; i++ {
		conn = rtGetter.GetConn()
		if conn == nil {
			t.Fail()
		}
		tc := conn.(testConn)
		trace = append(trace, int32(tc))
	}

	t.Log(zkm.Status())
	t.Log(trace)
}

func TestHashGet(t *testing.T) {
	mPath := "/test"
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	err := cli.CreatePersistNode(mPath)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	defer cli.DeleteNode(mPath)

	fmt.Println("===============create watch dir===============")
	time.Sleep(2 * time.Second)

	zkm := NewZKMonitor(testServers, time.Second*5, &testService{}, nil, mPath)
	zkm.Run()
	defer zkm.Close()

	t.Log(zkm.Status())

	hGetter := zkm.HashGetter(nil)
	conn := hGetter.GetConn([]byte("key"))
	if conn != nil {
		t.Fail()
	}
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6379")
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6380")

	fmt.Println("===============create addr node===============")
	time.Sleep(2 * time.Second)

	key := []byte{0, 1, 2, 3}
	trace := make([]int32, 0)
	for i := 0; i < 10; i++ {
		key[0] = byte(rand.Int() % 256)
		conn = hGetter.GetConn(key)
		if conn == nil {
			t.Fail()
		}
		tc := conn.(testConn)
		trace = append(trace, int32(tc))
	}

	t.Log(zkm.Status())
	t.Log(trace)
}

func BenchmarkRoundTripGet(b *testing.B) {
	mPath := "/test"
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	cli.CreatePersistNode(mPath)
	defer cli.DeleteNode(mPath)

	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6379")
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6380")

	fmt.Println("===============create addr dir===============")
	time.Sleep(2 * time.Second)

	zkm := NewZKMonitor(testServers, time.Second*5, &testService{}, nil, mPath)
	zkm.Run()
	defer zkm.Close()
	time.Sleep(2 * time.Second)

	rtGetter := zkm.RoundTripGetter()

	nilCnt := int32(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			conn := rtGetter.GetConn()
			if conn == nil {
				atomic.AddInt32(&nilCnt, 1)
			}
		}
	})
	b.Log("nil connection count:", nilCnt)
}

func BenchmarkHashGet(b *testing.B) {
	mPath := "/test"
	cli := NewZKClient(testServers, time.Second*5, nil)
	defer cli.Close()
	cli.CreatePersistNode(mPath)
	defer cli.DeleteNode(mPath)

	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6379")
	cli.CreateEphemeralNode(mPath + "/127.0.0.1:6380")

	fmt.Println("===============create addr dir===============")
	time.Sleep(2 * time.Second)

	zkm := NewZKMonitor(testServers, time.Second*5, &testService{}, nil, mPath)
	zkm.Run()
	defer zkm.Close()
	time.Sleep(2 * time.Second)

	hGetter := zkm.HashGetter(nil)
	key := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	nilCnt := int32(0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key[0] = byte(rand.Int())
			conn := hGetter.GetConn(key)
			if conn == nil {
				atomic.AddInt32(&nilCnt, 1)
			}
		}
	})
	b.Log("nil connection count:", nilCnt)
}
