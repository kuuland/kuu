package kuu

import (
	"fmt"
	"os"

	"github.com/garyburd/redigo/redis"
)

// EventHandler 事件处理器
type EventHandler func(...interface{})

// EventHandlers 事件处理器集合
type EventHandlers []EventHandler

var (
	events     = map[string]EventHandlers{}
	url        = os.Getenv("REDIS_URL")
	conn       redis.Conn
	psc        redis.PubSubConn
	subscribes = map[string]string{}
)

func init() {
	dial(url)
}

func dial(u string) {
	if u == "" {
		return
	}
	if c, err := redis.DialURL(u); err == nil {
		conn = c
		psc = redis.PubSubConn{Conn: c}
	}
}

// On 事件订阅
func On(event string, handlers ...EventHandler) {
	if event == "" || len(handlers) == 0 {
		return
	}
	for _, handler := range handlers {
		if events[event] == nil {
			events[event] = make(EventHandlers, 0)
		}
		events[event] = append(events[event], handler)
	}
	subscribe(event)
}

// Emit 事件触发
func Emit(event string, args ...interface{}) {
	if event == "" || events[event] == nil {
		return
	}
	handlers := events[event]
	for _, handler := range handlers {
		handler(args...)
	}
	publish(event, args...)
}

func subscribe(event string) error {
	if conn == nil || subscribes[event] != "" {
		return nil
	}
	subscribes[event] = event
	conn.Send("SUBSCRIBE", event)
	conn.Flush()
	for {
		reply, err := conn.Receive()
		fmt.Println("redis-reply:", reply)
		if err != nil {
			Error(err)
			continue
		}
		Emit(event, reply)
	}
}

func publish(event string, args ...interface{}) {
	if conn == nil {
		return
	}
	var data interface{}
	if len(args) > 0 {
		data = args[0]
	}
	conn.Send("PUBLISH", event, data)
	conn.Flush()
}
