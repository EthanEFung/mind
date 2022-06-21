package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ethanefung/mind/models"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

type RoomName struct {}
var room = RoomName{}

func LobbyContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), room, "lobby")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoomCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		param := chi.URLParam(r, "roomID")
		fmt.Println("roomId", param)
		ctx := context.WithValue(r.Context(), room, param)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WSHandler(h *models.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("upgrading")
		ctx := r.Context()
		fmt.Println(ctx)
		fmt.Println("value:", ctx.Value(room))
		if (ctx.Value(room) == nil) {
			http.Error(w, "no room", http.StatusForbidden)
		}
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
		/**
		TODO: Here we want to grab the name of the room from the query params
		and create a subscription to the proper room. Even the lobby should
		have a parameter
		*/

		subscription := &models.Subscription{
			Room: ctx.Value(room).(string),
			Client: client,
		}
		client.Hub.Register <- subscription
		go subscription.WritePump()
		go subscription.ReadPump()
	}
}