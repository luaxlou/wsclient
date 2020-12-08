package wsclient

import (
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

const HOST = "localhost:9999"

func TestNew(t *testing.T) {

	wsc := New("localhost:9999", http.Header{
		"MERCHANT_KEY": []string{
			"1001",
		},
	})

	wsc.Ready = func() {
		err := wsc.WriteMessage(websocket.TextMessage, []byte("hello"))

		if err != nil {
			log.Println(err.Error())
		}

	}

	select {}
}
