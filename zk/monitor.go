package zk

import (
	"sync"
	"sync/atomic"

	"time"

	"hash/fnv"

	"bytes"
	"fmt"
	"strings"

	log "github.com/alecthomas/log4go"
)

// use zk to monitor service address
// for example, redis regist in path like /redis with ephemeral node named addr:port
// eg. /redis/10.0.0.1:6379 /redis/10.0.0.2:6379
// watch dir /redis and update redis instance addr with sub node

type ServiceConn interface{}

type ServiceCli interface {
	GetConn() ServiceConn
	Status() string
	Close()
	OnDisable()
	OnEnable()
}

// watch zk service address dir
// when got new address event call InitCli
// when got address deleted event call ServiceCli.OnDisable and set cli invalid, will skip the cli on GetConn
// when got address recovery event call ServiceCli.OnRecover and set cli valid
type Service interface {
	InitCli(addr string, arg interface{}) ServiceCli
}

type AddrMonitor interface {
	ValidAddr() (<-chan []string, error)
	Close()
}

type addrMonitor struct {
	closed  bool
	cli     *ZKClient
	mpath   string
	watcher []*DirWatcher
}

func newAddrMonitor(cli *ZKClient, monitorPath string) *addrMonitor {
	return &addrMonitor{
		cli:     cli,
		mpath:   monitorPath,
		watcher: make([]*DirWatcher, 0),
	}
}

func (m *addrMonitor) ValidAddr() (<-chan []string, error) {
	w := m.cli.WatchDir(m.mpath)
	go func() {
		for err := range w.ErrChan {
			log.Error("watch path:%s error:%v", m.mpath, err)
		}
	}()
	m.watcher = append(m.watcher, w)
	return w.Children, nil
}

func (m *addrMonitor) Close() {
	if m.closed {
		return
	}
	m.closed = true
	for _, w := range m.watcher {
		w.Close()
	}
	m.cli.Close()
}

const (
	invalidMark = int32(0)
	validMark   = ^invalidMark
)

type monitorServCli struct {
	valid        int32
	cli          ServiceCli
	addr         string
	enableCount  int
	disableCount int
}

func (c *monitorServCli) setValid(valid bool) {
	if valid {
		atomic.StoreInt32(&c.valid, validMark)
	} else {
		atomic.StoreInt32(&c.valid, invalidMark)
	}
}

func (c *monitorServCli) isValid() bool {
	v := atomic.LoadInt32(&c.valid)
	return v == validMark
}

func (c *monitorServCli) disable() {
	if !c.isValid() {
		return
	}
	c.setValid(false)
	c.cli.OnDisable()
	c.disableCount++
}

func (c *monitorServCli) enable() {
	if c.isValid() {
		return
	}
	c.setValid(true)
	c.cli.OnEnable()
	c.enableCount++
}

func (c *monitorServCli) close() {
	c.setValid(false)
	c.cli.Close()
}

func (c *monitorServCli) String() string {
	w := bytes.NewBuffer(make([]byte, 0))
	valid := c.isValid()

	fmt.Fprintf(w, "{\"valid\":%t,\"addr\":\"%s\",\"enableCnt\":%d,\"disableCnt\":%d",
		valid, c.addr, c.enableCount, c.disableCount)
	if valid {
		fmt.Fprintf(w, ",\"cli\":\"%s\"}", c.cli.Status())
	} else {
		fmt.Fprintf(w, "}")
	}
	return w.String()
}

type ZKMonitor struct {
	servIdx          map[string]int
	servCli          []monitorServCli // this array never shrink, set monitorServCli invalid to skip this element
	servLock         *sync.RWMutex
	availAddr        int32
	addrMonitor      AddrMonitor
	serv             Service
	servArg          interface{}
	addrChangedCount int
}

func NewZKMonitor(zkServers []string, timeout time.Duration, serv Service, servArg interface{},
	monitorPath string) *ZKMonitor {
	zkCli := NewZKClient(zkServers, timeout, nil)
	addrM := newAddrMonitor(zkCli, monitorPath)
	return &ZKMonitor{
		servIdx:     make(map[string]int),
		servCli:     make([]monitorServCli, 0),
		servLock:    &sync.RWMutex{},
		addrMonitor: addrM,
		serv:        serv,
		servArg:     servArg,
	}
}

func (zkm *ZKMonitor) Run() {
	validAddr, err := zkm.addrMonitor.ValidAddr()
	if err != nil {
		log.Error("run zk monitor failed, error:%v", err)
		return
	}
	log.Debug("zk monitor start")

	sched := make(chan struct{})
	go func() {
		close(sched)
		for addrs := range validAddr {
			zkm.onAddrChange(addrs)
			zkm.addrChangedCount++
		}
		log.Debug("zk monitor stopped")
	}()
	<-sched
}

func (zkm *ZKMonitor) Close() {
	zkm.addrMonitor.Close()
	zkm.servLock.Lock()
	for _, cli := range zkm.servCli {
		cli.close()
	}
	zkm.servLock.Unlock()
}

// enable/disable service cli, add new service cli
func (zkm *ZKMonitor) updateServCli(addrs []string) {
	addrMark := make(map[string]bool)
	var added []string

	for _, addr := range addrs {
		if idx, ok := zkm.servIdx[addr]; !ok {
			added = append(added, addr)
		} else {
			zkm.servCli[idx].enable()
		}
		addrMark[addr] = true
	}
	for addr, i := range zkm.servIdx {
		if _, ok := addrMark[addr]; !ok {
			zkm.servCli[i].disable()
		}
	}
	if len(added) > 0 {
		n := len(zkm.servCli)
		for _, addr := range added {
			cli := zkm.serv.InitCli(addr, zkm.servArg)

			zkm.servLock.Lock()
			zkm.servCli = append(zkm.servCli, monitorServCli{
				cli:   cli,
				valid: validMark,
				addr:  addr,
			})
			zkm.servLock.Unlock()

			zkm.servIdx[addr] = n
			n++
		}
	}
}

// adjust valid cli to avoid valid cli node not even
func (zkm *ZKMonitor) adjustServCli() {
	sz := len(zkm.servCli)
	if sz < 2 {
		return
	}
	head, tail := 0, sz-1
	for head < tail {
		if zkm.servCli[head].isValid() {
			head++
			continue
		}
		if !zkm.servCli[tail].isValid() {
			tail--
			continue
		}
		// head point to invalid cli and tail point to valid cli
		// swap head and tail
		headAddr, tailAddr := zkm.servCli[head].addr, zkm.servCli[tail].addr
		zkm.servLock.Lock()
		zkm.servCli[head], zkm.servCli[tail] = zkm.servCli[tail], zkm.servCli[head]
		zkm.servLock.Unlock()
		zkm.servIdx[headAddr], zkm.servIdx[tailAddr] = tail, head
		head++
		tail--
	}
}

func (zkm *ZKMonitor) onAddrChange(addrs []string) {
	availAddr := int32(len(addrs))
	zkm.updateServCli(addrs)
	zkm.adjustServCli()
	// update after adjust
	atomic.StoreInt32(&zkm.availAddr, availAddr)
}

func findFirstValid(servCli []monitorServCli, start int) (bool, int) {
	sz := len(servCli)
	if sz == 0 {
		return false, -1
	}
	if start < 0 || start >= sz {
		start = 0
	}
	if servCli[start].isValid() {
		return true, start
	}

	i := (start + 1) % sz
	for ; i != start; i = (i + 1) % sz {
		if servCli[i].isValid() {
			return true, i
		}
	}
	return false, -1
}

func (zkm *ZKMonitor) RoundTripGetter() *RoundTripGetter {
	return &RoundTripGetter{
		zkm: zkm,
	}
}

type RoundTripGetter struct {
	zkm *ZKMonitor
	nextUse uint32
}

// find first valid cli and GetConn by roundtrip
func (rtg *RoundTripGetter) GetConn() ServiceConn {
	zkm := rtg.zkm
	availAddr := atomic.LoadInt32(&zkm.availAddr)
	if availAddr == 0 {
		return nil
	}

	n := atomic.AddUint32(&rtg.nextUse, uint32(1))
	cur := n % uint32(availAddr)

	zkm.servLock.RLock()
	defer zkm.servLock.RUnlock()
	ok, idx := findFirstValid(zkm.servCli, int(cur))
	if ok {
		conn := zkm.servCli[idx].cli.GetConn()
		return conn
	}
	return nil
}

// default use fnv-1a hash func
func (zkm *ZKMonitor) HashGetter(hashFn func([]byte) uint32) *HashGetter {
	if hashFn == nil {
		hashFn = func(data []byte) uint32 {
			h := fnv.New32a()
			h.Write(data)
			return h.Sum32()
		}
	}
	return &HashGetter{
		zkm:    zkm,
		hashFn: hashFn,
	}
}

type HashGetter struct {
	zkm    *ZKMonitor
	hashFn func([]byte) uint32
}

// find first valid cli and GetConn by hash
func (hg *HashGetter) GetConn(key []byte) ServiceConn {
	hashVal := hg.hashFn(key)

	zkm := hg.zkm
	zkm.servLock.RLock()
	defer zkm.servLock.RUnlock()

	availAddr := atomic.LoadInt32(&zkm.availAddr)
	cur := int32(0)
	if availAddr > 1 {
		cur = int32(hashVal) % availAddr
		if cur < 0 {
			cur += availAddr
		}
	}

	ok, idx := findFirstValid(zkm.servCli, int(cur))
	if ok {
		return zkm.servCli[idx].cli.GetConn()
	}
	return nil
}

func (zkm *ZKMonitor) Status() string {
	zkm.servLock.RLock()
	defer zkm.servLock.RUnlock()

	s := make([]string, 0, len(zkm.servCli))
	for _, cli := range zkm.servCli {
		s = append(s, fmt.Sprintf("    %s", &cli))
	}
	return fmt.Sprintf("{\n  \"availAddr\":%d,\n  \"addrChangeCnt\":%d,\n  \"service\":[\n%s\n  ]\n}\n",
		zkm.availAddr, zkm.addrChangedCount, strings.Join(s, ",\n"))
}
