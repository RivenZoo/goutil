package redis

import (
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
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
