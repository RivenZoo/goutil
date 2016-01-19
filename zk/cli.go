package zk

import (
	"errors"
	"net"
	"path"
	"time"

	log "github.com/alecthomas/log4go"

	"sync"

	"strings"

	"github.com/samuel/go-zookeeper/zk"
)

const zkSpliter = "/"

var (
	ErrNoZkConnection = errors.New("no zk connection")
)

type ZKClient struct {
	Servers        []string
	SessionTimeout time.Duration
	Dialer         zk.Dialer

	fnLock           *sync.RWMutex
	fnOnSessionBuild func(*ZKClient)
	connLock         *sync.RWMutex
	conn             *zk.Conn
}

func NewZKClient(addrs []string, sessionTimeout time.Duration,
	dialer func(network, address string, timeout time.Duration) (net.Conn, error)) *ZKClient {
	cli := &ZKClient{
		Servers:        addrs,
		SessionTimeout: sessionTimeout,
		Dialer:         zk.Dialer(dialer),
		fnLock:         &sync.RWMutex{},
		connLock:       &sync.RWMutex{},
	}
	return cli
}

func (cli *ZKClient) Close() {
	log.Debug("zk client close")
	cli.connLock.Lock()
	defer cli.connLock.Unlock()
	if cli.conn != nil {
		cli.conn.Close()
		cli.conn = nil
	}
}

func (cli *ZKClient) getConn() *zk.Conn {
	cli.connLock.RLock()
	conn := cli.conn
	cli.connLock.RUnlock()

	if conn != nil {
		return conn
	}

	err := cli.Connect()
	if err != nil {
		log.Error("connect zk server:%s error:%v", cli.Servers, err)
		return nil
	}
	return cli.conn
}

func (cli *ZKClient) Connect() error {
	conn, session, err := zk.ConnectWithDialer(cli.Servers, cli.SessionTimeout, cli.Dialer)
	if err != nil {
		return err
	}
	cli.connLock.Lock()
	cli.conn = conn
	cli.connLock.Unlock()

	sched := make(chan struct{})
	go func() {
		close(sched)
		for e := range session {
			if e.State == zk.StateDisconnected {
				log.Error("session disconnected, event:%s", e)
			} else if e.State == zk.StateHasSession {
				log.Debug("session build, event:%s", e)
				cli.fnLock.RLock()
				fn := cli.fnOnSessionBuild
				cli.fnLock.RUnlock()
				if fn != nil {
					fn(cli)
				}
			} else {
				log.Debug("session recv event:%s", e)
			}
		}
		log.Info("session channel closed")
	}()
	<-sched
	return nil
}

// set func which would be called on session built with zk server
// usually used to create ephemeral node after disconnect
func (cli *ZKClient) SetFuncOnSessionBuild(fn func(*ZKClient)) {
	cli.fnLock.Lock()
	cli.fnOnSessionBuild = fn
	cli.fnLock.Unlock()
}

func (cli *ZKClient) CreatePersistNode(npath string) error {
	npath = strings.TrimRight(npath, zkSpliter)
	return cli.createNode(npath, int32(0))
}

func (cli *ZKClient) CreateEphemeralNode(npath string) error {
	npath = strings.TrimRight(npath, zkSpliter)
	d := path.Dir(npath)
	err := cli.createNode(d, int32(0))
	if err != nil {
		return err
	}

	conn := cli.getConn()
	if conn == nil {
		return ErrNoZkConnection
	}
	_, err = conn.Create(npath, []byte(""), int32(zk.FlagEphemeral), zk.WorldACL(zk.PermAll))
	return err
}

func (cli *ZKClient) createNode(npath string, flags int32) error {
	conn := cli.getConn()
	if conn == nil {
		return ErrNoZkConnection
	}
	level := strings.Split(npath, zkSpliter)
	if len(level) == 0 {
		return nil
	}

	acl := zk.WorldACL(zk.PermAll)
	npath = ""
	for _, lvl := range level {
		if lvl == "" {
			continue
		}
		npath = npath + zkSpliter + lvl

		exist, _, err := conn.Exists(npath)
		if err != nil {
			return err
		}
		if !exist {
			_, err = conn.Create(npath, []byte(lvl), flags, acl)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cli *ZKClient) NodeExist(npath string) (bool, error) {
	conn := cli.getConn()
	if conn == nil {
		return false, ErrNoZkConnection
	}
	npath = strings.TrimRight(npath, zkSpliter)
	ok, _, err := conn.Exists(npath)
	return ok, err
}

func (cli *ZKClient) DeleteNode(npath string) error {
	conn := cli.getConn()
	if conn == nil {
		return ErrNoZkConnection
	}
	npath = strings.TrimRight(npath, zkSpliter)
	return conn.Delete(npath, int32(0))
}

func (cli *ZKClient) GetChildren(path string) ([]string, error) {
	conn := cli.getConn()
	if conn == nil {
		return nil, ErrNoZkConnection
	}
	children, _, err := conn.Children(path)
	if err != nil {
		return nil, err
	}
	return children, nil
}

type DirWatcher struct {
	closed   bool
	stop     chan struct{}
	exited   chan struct{}
	Children chan []string
	ErrChan  chan error
}

func (w *DirWatcher) Close() {
	if w.closed {
		return
	}
	w.closed = true
	close(w.stop)
	<-w.exited
	close(w.Children)
	close(w.ErrChan)
}

// watch dir and get dir mirror after event occur
func (cli *ZKClient) WatchDir(dir string) (w *DirWatcher) {
	conn := cli.getConn()
	if conn == nil {
		log.Error("watch dir:%s fail, no zk connection", dir)
		return nil
	}
	w = &DirWatcher{
		stop:     make(chan struct{}),
		exited:   make(chan struct{}),
		Children: make(chan []string),
		ErrChan:  make(chan error),
	}
	sched := make(chan struct{})
	go func() {
		close(sched)
		for {
			if w.closed {
				close(w.exited)
				log.Info("stop watch path:%s", dir)
				return
			}
			snapshot, _, ev, err := conn.ChildrenW(dir)
			if err != nil {
				w.ErrChan <- err
				log.Error("watch path:%s error:%s, wait 5seconds", dir, err.Error())
				time.Sleep(time.Second * 5)
				continue
			}
			w.Children <- snapshot

			select {
			case e := <-ev:
				log.Debug("watch path:%s recv event:%s", dir, e)
				if e.Err != nil {
					w.ErrChan <- e.Err
					log.Error("watch path:%s event error:%s", dir, e.Err.Error())
				}
			case <-w.stop:
				// ensure exit in next loop
				w.closed = true
			}
		}
	}()
	<-sched
	return w
}

func (cli *ZKClient) WatchDirOnce(path string) ([]string, <-chan zk.Event, error) {
	conn := cli.getConn()
	if conn == nil {
		return nil, nil, ErrNoZkConnection
	}
	children, _, ev, err := conn.ChildrenW(path)
	if err != nil {
		return nil, nil, err
	}
	return children, ev, nil
}
