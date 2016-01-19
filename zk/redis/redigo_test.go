package redis

import (
	"testing"
	"time"
	"gopublic/zk"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	testServers = []string{"127.0.0.1:2181"}
)

func TestRedisGetConn(t *testing.T) {
	redisAddr := "10.20.231.32:6379"
	cli := zk.NewZKClient(testServers, time.Second*5, nil)
	mpath := "/test"
	rds := NewZKRedisCli(testServers, mpath, &RedisDbConf{Timeout:time.Second*5})
	rds.UseRoundTripGet()
	rds.Run()
	cli.CreateEphemeralNode(mpath+"/"+redisAddr)
	time.Sleep(time.Second*2)
	fmt.Println("=============add redis addr==============")
	fmt.Println(rds.Status())

	cnt:=int32(0)
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j< 1000; j++{
				conn := rds.RoundTripGet()
				if conn == nil {
					atomic.AddInt32(&cnt, 1)
					continue
				}
				conn.Close()
			}
		}()
	}
	wg.Wait()
	fmt.Println("nil conn:", cnt)
}
