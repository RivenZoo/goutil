package redis

// redis client with zookeeper
// connect redis use "github.com/garyburd/redigo/redis"

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/freecoder-zrw/goutil/zk"

	"github.com/garyburd/redigo/redis"
)

const (
	// zk session timeout, 5 seconds
	ZkTimeout = 5 * time.Second

	MaxRedisIdleConn = 32
)

// Redigo Redis client.
type RedisCli struct {
	p        *redis.Pool // redis connection pool
	conf     RedisConf
	dialFunc func() (redis.Conn, error)
}

type RedisConf struct {
	Addr     string        `json:"addr"`
	Password string        `json:"password"`
	DbNo     int           `json:"dbNo"`
	Timeout  time.Duration `json:"timeout"`
}

func (conf *RedisConf) String() string {
	s, err := json.Marshal(conf)
	if err != nil {
		return err.Error()
	}
	return string(s)
}

// create new redis client
// if no need to auth, let pwd=""
func NewRedisCli(conf *RedisConf) *RedisCli {
	cli := &RedisCli{}
	cli.conf.Addr = conf.Addr
	cli.conf.Password = conf.Password
	cli.conf.DbNo = conf.DbNo
	cli.conf.Timeout = conf.Timeout
	cli.initConnPool(conf)
	return cli
}

func ConnectRedis(conf *RedisConf) (redis.Conn, error) {
	c, err := redis.DialTimeout("tcp", conf.Addr, conf.Timeout, conf.Timeout, conf.Timeout)
	if err != nil {
		return nil, err
	}
	if conf.Password != "" {
		if _, err := c.Do("AUTH", conf.Password); err != nil {
			c.Close()
			return nil, err
		}
	}
	_, selectErr := c.Do("SELECT", conf.DbNo)
	if selectErr != nil {
		c.Close()
		return nil, selectErr
	}
	return c, nil
}

func (rc *RedisCli) Config() RedisConf {
	return rc.conf
}

func (rc *RedisCli) ActiveConn() int {
	return rc.p.ActiveCount()
}

func (rc *RedisCli) Close() {
	rc.p.Close()
}

// actually do the redis cmds
func (rc *RedisCli) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := rc.p.Get()
	defer c.Close()

	reply, err = c.Do(commandName, args...)
	return
}

// need to call PutConn after use
func (rc *RedisCli) GetConn() redis.Conn {
	return rc.p.Get()
}

func (rc *RedisCli) PutConn(c redis.Conn) {
	c.Close()
}

// create new connection pool to redis.
func (rc *RedisCli) initConnPool(conf *RedisConf) {
	if rc.dialFunc == nil {
		rc.dialFunc = func() (c redis.Conn, err error) {
			c, err = redis.DialTimeout("tcp", conf.Addr, conf.Timeout, conf.Timeout, conf.Timeout)
			if err != nil {
				return nil, err
			}
			if conf.Password != "" {
				if _, err := c.Do("AUTH", conf.Password); err != nil {
					c.Close()
					return nil, err
				}
			}
			_, selecterr := c.Do("SELECT", conf.DbNo)
			if selecterr != nil {
				c.Close()
				return nil, selecterr
			}
			return
		}
	}

	// initialize a new pool
	rc.p = &redis.Pool{
		MaxIdle:     MaxRedisIdleConn,
		IdleTimeout: 180 * time.Second,
		Dial:        rc.dialFunc,
	}
}

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
