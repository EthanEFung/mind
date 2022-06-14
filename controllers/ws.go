package controllers

import (
	"fmt"
	"net/http"

	"github.com/ethanefung/mind/models"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func WSHandler(h *models.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("upgrading")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctxErr := fmt.Errorf("upgrade conn: %v", err)
			http.Error(w, ctxErr.Error(), http.StatusInternalServerError)
			return
		}
		client := &models.Client{
			Hub: h,
			Conn: conn,
			Send: make(chan []byte, 256),
		}
		client.Hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
	}
}