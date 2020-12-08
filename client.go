package wsclientsdk

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/luaxlou/goutils/tools/debugutils"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var REQUEST_TIMEOUT = 20

type Message struct {
	MessageType int
	Data        []byte
}

type WsClient struct {
	Mux sync.RWMutex

	conn *websocket.Conn

	url url.URL

	isConnected bool

	handlers sync.Map

	readCh chan *Message

	Ready func()

	header http.Header
}

type RequestHandler struct {
	Callback func(statusCode int, data []byte)

	Ts time.Time
}

func New(addr string, header http.Header) *WsClient {

	wsc := &WsClient{
		header: header,
		url:    url.URL{Scheme: "ws", Host: addr, Path: "/ws"},
		readCh: make(chan *Message),
	}

	go wsc.timeoutCheck()

	go wsc.connect()

	return wsc
}

func (wsc *WsClient) reconnect() {

	wsc.Close()
	wsc.connect()
}

func (wsc *WsClient) Close() {

	if wsc.conn != nil {
		wsc.conn.Close()
	}

}
func (wsc *WsClient) connect() {
	log.Println("Connecting websocket")

	ws, _, err := websocket.DefaultDialer.Dial(wsc.url.String(), wsc.header)

	wsc.conn = ws

	if err == nil {
		log.Printf("Dial: connection was successfully established with %s\n", wsc.url.String())

		wsc.start()

		if wsc.Ready != nil {
			wsc.Ready()
		}
		return
	}

	log.Println("could not connect:", err.Error())

	after := time.After(time.Second * 2)

	<-after

	wsc.connect()

}

func (wsc *WsClient) ReadMessage() (messageType int, message []byte, err error) {

	messageType, message, err = wsc.conn.ReadMessage()

	return
}

func (wsc *WsClient) WriteMessage(messageType int, data []byte) error {
	if wsc.conn == nil {
		return errors.New("no connection")
	}

	wsc.Mux.Lock()
	err := wsc.conn.WriteMessage(messageType, data)
	wsc.Mux.Unlock()
	if err != nil {

		return err

	}

	return nil
}

func (wsc *WsClient) keepalive() {

	wsc.conn.SetPongHandler(func(string) error {


		wsc.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {

		<-ticker.C

		wsc.WriteMessage(websocket.PingMessage, []byte{})

	}
}

func (wsc *WsClient) timeoutCheck() {
	t := time.NewTicker(time.Second)

	go func() {

		for {

			nt := <-t.C

			wsc.handlers.Range(func(key, value interface{}) bool {

				h := value.(*RequestHandler)

				if nt.Sub(h.Ts).Seconds() > float64(REQUEST_TIMEOUT) {

					h.Callback(http.StatusRequestTimeout, nil)

					wsc.handlers.Delete(key)

				}

				return true
			})

		}

	}()
}

func (wsc *WsClient) start() {

	go wsc.loopReadMessage()
	go wsc.keepalive()

}

func (wsc *WsClient) loopReadMessage() {

	defer wsc.reconnect()

	for {

		mt, message, err := wsc.ReadMessage()

		if err != nil {

			log.Println("client read error :", err.Error())
			break
		}

		switch mt {

		case websocket.BinaryMessage:

			var res TempCommand

			if err := json.Unmarshal(message, &res); err != nil {

				debugutils.Println(err.Error())

				continue

			}

			if res.Type == 1 {

				if temp, ok := wsc.handlers.Load(res.RequestId); ok && temp != nil {

					h := temp.(*RequestHandler)

					h.Callback(http.StatusOK, message)

					wsc.handlers.Delete(res.RequestId)

				}
			}
			//TODO 接收服务端发过来的请求

		}

	}
}

func (wsc *WsClient) Send(name string, data interface{}) error {

	cmd := NewCommand(name, data)

	bs, _ := json.Marshal(&cmd)

	return wsc.WriteMessage(websocket.BinaryMessage, bs)

}

func (wsc *WsClient) SendRequest(name string, data interface{}, callback func(code int, data []byte)) error {

	uid := uuid.New().String()

	cmd := NewCommand(name, data)

	cmd.RequestId = uid

	if callback != nil {
		wsc.handlers.Store(uid, &RequestHandler{
			Callback: callback,
			Ts:       time.Now(),
		})
	}

	bs, _ := json.Marshal(&cmd)

	return wsc.WriteMessage(websocket.BinaryMessage, bs)

}
