package myspider

import (
	"encoding/json"
	// "fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/hu17889/go_spider/core/common/request"
)

type RedisScheduler struct {
	locker *sync.Mutex
	conn   redis.Conn
}
type myRequest struct {
	Url string

	// Responce type: html json jsonp text
	RespType string

	// GET POST
	Method string

	// POST data
	Postdata string

	// name for marking url and distinguish different urls in PageProcesser and Pipeline
	Urltag string

	Meta interface{}
}

func NewRedisScheduler() *RedisScheduler {
	locker := new(sync.Mutex)
	conn, err := redis.DialTimeout("tcp", "127.0.0.1:6379", 0, 1*time.Second, 1*time.Second)
	if err != nil {
		panic(err)
	}
	return &RedisScheduler{locker, conn}
}

func (this *RedisScheduler) Push(requ *request.Request) {
	this.locker.Lock()
	defer this.locker.Unlock()
	req := myRequest{}
	req.Method = requ.GetMethod()
	req.Postdata = requ.GetPostdata()
	req.RespType = requ.GetResponceType()
	req.Url = requ.GetUrl()
	req.Urltag = requ.GetUrlTag()
	// req.Meta = requ.GetMeta()
	b, _ := json.Marshal(req)
	this.conn.Do("LPUSH", "Requests", b)
}

func (this *RedisScheduler) Poll() *request.Request {

	this.locker.Lock()
	defer this.locker.Unlock()
	var myReq myRequest
	r, err := this.conn.Do("LPOP", "Requests")
	if err != nil {
		return nil
	} else if r == nil {
		return nil
	}
	b, err := redis.Bytes(r, err)
	if err != nil {
		return nil
	}
	json.Unmarshal(b, &myReq)
	return request.NewRequest(myReq.Url, myReq.RespType, myReq.Urltag, myReq.Method, myReq.Postdata, nil, nil, nil, myReq.Meta)
}

func (this *RedisScheduler) Count() int {
	this.locker.Lock()
	defer this.locker.Unlock()
	count, _ := redis.Int(this.conn.Do("LLEN", "Requests"))
	return count
}
