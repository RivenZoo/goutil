package zk

import (
	"errors"
	"sync"
	"time"

	log "github.com/alecthomas/log4go"
)

var (
	errRegisterClosed = errors.New("register closed")
)

// register service address to zookeeper path
type ZKRegister struct {
	closed       bool
	cli          *ZKClient
	servicePath  string
	serviceAddrs []string
	mutex        *sync.Mutex
	addrChan     chan string
	errChan      chan error
}

func NewZKRegister(servers []string, sessionTimeout time.Duration,
	servicePath string) *ZKRegister {
	r := &ZKRegister{
		cli:          NewZKClient(servers, sessionTimeout, nil),
		servicePath:  servicePath,
		serviceAddrs: make([]string, 0),
		mutex:        &sync.Mutex{},
		addrChan:     make(chan string),
		errChan:      make(chan error),
	}
	onZKSessionFunc := func(zkCli *ZKClient) {
		r.mutex.Lock()
		addrs := make([]string, len(r.serviceAddrs))
		copy(addrs, r.serviceAddrs)
		r.mutex.Unlock()
		for _, addr := range addrs {
			npath := r.servicePath + "/" + addr
			err := zkCli.CreateEphemeralNode(npath)
			if err != nil {
				log.Error("create node:%s failed on session build, error:%v", npath, err)
			}
		}
	}
	r.cli.SetFuncOnSessionBuild(onZKSessionFunc)
	go func() {
		for addr := range r.addrChan {
			npath := r.servicePath + "/" + addr
			err := r.cli.CreateEphemeralNode(npath)
			r.errChan <- err
		}
	}()
	return r
}

func (r *ZKRegister) Register(addr string) error {
	if r.closed {
		return errRegisterClosed
	}
	r.addrChan <- addr
	err := <-r.errChan
	if err != nil {
		return err
	}
	r.mutex.Lock()
	r.serviceAddrs = append(r.serviceAddrs, addr)
	r.mutex.Unlock()
	return nil
}

func (r *ZKRegister) Close() {
	r.closed = true
	close(r.addrChan)
	close(r.errChan)
	r.cli.Close()
}
