package client

import (
	"RunnerGo-engine/log"
	"RunnerGo-engine/tools"
	"github.com/gorilla/websocket"
	"time"
)

type WebsocketClient struct {
	Conn    *websocket.Conn
	Addr    *string
	IsAlive bool
	Timeout int
}

func WebSocketRequest(url string, body string, headers map[string][]string, timeout int) (resp []byte, requestTime uint64, sendBytes uint, err error) {
	websocketClient := NewWsClientManager(url, timeout)
	log.Logger.Info("connecting to ", url)
	if websocketClient.IsAlive == false {
		for i := 0; i < 3; i++ {
			startTime := time.Now().UnixMilli()
			websocketClient.Conn, _, err = websocket.DefaultDialer.Dial(url, headers)
			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				log.Logger.Error("第", i, "次connecting to:", url, "失败")
				continue
			}
			websocketClient.IsAlive = true
			bodyBytes := []byte(body)
			err = websocketClient.Conn.WriteMessage(websocket.TextMessage, bodyBytes)
			sendBytes = uint(len(body))
			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				log.Logger.Error("第", i, "次向", url, "写消息失败失败")
				continue
			}

			_, resp, err = websocketClient.Conn.ReadMessage()

			if err != nil {
				requestTime = tools.TimeDifference(startTime)
				log.Logger.Error("读取websocket消息错误, 尝试重连", err)
				websocketClient.IsAlive = false
				// 出现错误，退出读取，尝试重连
				continue
			}
			//requestTime = tools.TimeDifference(startTime)
			requestTime = tools.TimeDifference(startTime)
			break
		}
	}
	return

}

// NewWsClientManager 构造函数
func NewWsClientManager(url string, timeout int) *WebsocketClient {
	var conn *websocket.Conn
	return &WebsocketClient{
		Addr:    &url,
		Conn:    conn,
		IsAlive: false,
		Timeout: timeout,
	}
}
