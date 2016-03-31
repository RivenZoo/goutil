package redis

// redis client with zookeeper
// connect redis use "github.com/garyburd/redigo/redis"

import (
	"fmt"
	"time"

	"github.com/RivenZoo/goutil/zk"

	"github.com/garyburd/redigo/redis"
)

const (
	// zk session timeout, 5 seconds
	ZkTimeout = 5 * time.Second
)

type redisService struct{}

type redisCli struct {
	cli *RedisCli
}

func (s *redisService) InitCli(addr string, arg interface{}) zk.ServiceCli {
	dbConf := arg.(*RedisDbConf)
	conf := &RedisConf{
		Addr:     addr,
		Password: dbConf.Password,
		DbNo:     dbConf.DbNo,
		Timeout:  dbConf.Timeout,
	}
	cli := NewRedisCli(conf)
	return &redisCli{cli}
}

func (c *redisCli) GetConn() zk.ServiceConn {
	conn := c.cli.GetConn()
	return conn
}

func (c *redisCli) Status() string {
	conf := c.cli.Config()
	connCount := c.cli.ActiveConn()
	return fmt.Sprintf(`{"addr":"%s","db":%d,"activeConn":%d}`, conf.Addr, conf.DbNo, connCount)
}

func (c *redisCli) Close() {
	c.cli.Close()
}

func (c *redisCli) OnDisable() {}
func (c *redisCli) OnEnable()  {}

type RedisDbConf struct {
	Password string        `json:"password"`
	DbNo     int           `json:"dbNo"`
	Timeout  time.Duration `json:"timeout"`
}

type ZkRedisCli struct {
	*zk.ZKMonitor
	hashGetter  *zk.HashGetter
	roundGetter *zk.RoundTripGetter
}

// monitor redisAddrPath and update redis valid service
func NewZKRedisCli(zkServers []string, redisAddrPath string, dbConf *RedisDbConf) *ZkRedisCli {
	return &ZkRedisCli{
		zk.NewZKMonitor(zkServers, ZkTimeout, &redisService{}, dbConf, redisAddrPath),
		nil,
		nil,
	}
}

func (cli *ZkRedisCli) UseRoundTripGet() {
	if cli.roundGetter != nil {
		return
	}
	cli.roundGetter = cli.RoundTripGetter()
}

func (cli *ZkRedisCli) UseHashGet(hashFn func([]byte) uint32) {
	if cli.hashGetter != nil {
		return
	}
	cli.hashGetter = cli.HashGetter(hashFn)
}

// call redis.Conn Close after used
func (cli *ZkRedisCli) RoundTripGet() redis.Conn {
	if cli.roundGetter == nil {
		return nil
	}
	v := cli.roundGetter.GetConn()
	conn, ok := v.(redis.Conn)
	if !ok {
		return nil
	}
	return conn
}

// call redis.Conn Close after used
func (cli *ZkRedisCli) HashGet(key []byte) redis.Conn {
	if cli.hashGetter == nil {
		return nil
	}
	v := cli.hashGetter.GetConn(key)
	conn, ok := v.(redis.Conn)
	if !ok {
		return nil
	}
	return conn
}
