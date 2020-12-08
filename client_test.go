package wsclient

import (
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)


func TestNew(t *testing.T) {

	wsc := New("localhost:9999", http.Header{
		"some": []string{
			"one",
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
