package kuu

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		CheckOrigin:      func(r *http.Request) bool { return true },
		HandshakeTimeout: time.Second * 5,
	}
	wsConns sync.Map
)

// ModelWSRoute
var ModelWSRoute = RouteInfo{
	Name:   "模型变更通知WebSocket接口",
	Method: "GET",
	Path:   "/model/ws",
	HandlerFunc: func(c *Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			ERROR("websocket.upgrade: %v", err)
			return
		}
		defer func() {
			if _, ok := wsConns.Load(conn); ok {
				wsConns.Delete(conn)
			}
			conn.Close()
			INFO("websocket.close: %p", conn)
		}()
		wsConns.Store(conn, conn)
		INFO("websocket.connect: %p", conn)
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				ERROR(err)
				break
			}
			INFO("websocket.recv: %s", message)
			err = conn.WriteMessage(mt, message)
			if err != nil {
				ERROR(err)
				break
			}
		}
	},
}

// NotifyModelChange
func NotifyModelChange(modelName string) {
	if modelName == "" {
		return
	}
	message := []byte(fmt.Sprintf("change::%s", modelName))
	wsConns.Range(func(_, value interface{}) bool {
		if v, ok := value.(*websocket.Conn); ok {
			if err := v.WriteMessage(websocket.TextMessage, message); err != nil {
				ERROR(err)
			}
		}
		return true
	})
}
