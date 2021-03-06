package models

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)
const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space = []byte{' '}
)

type Subscription struct {
	Client *Client

	Room string
}

func (s *Subscription) ReadPump() {
	c := s.Client
	defer func() {
		c.Hub.Unregister <- s
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(fmt.Errorf("read message: %v", err))
			}
			break
		}
		msg := Message{
			Room: s.Room,
			Data:bytes.TrimSpace(bytes.Replace(message, newline, space, -1)),
		}
		c.Hub.Broadcast <- msg
	}
}

func (s Subscription) WritePump() {
	c := s.Client
  ticker := time.NewTicker(pingPeriod)
  defer func() {
    ticker.Stop()
    c.Conn.Close()
  }()
  for {
    select {
    case message, ok := <-c.Send:
      c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
      if !ok {
        // the hub closed the channel
        c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
      }

      w, err := c.Conn.NextWriter(websocket.TextMessage)
      if err != nil {
        return
      }
      w.Write(message)

      // Add queued chat messages to the current websocket message
      n := len(c.Send)
      for i := 0; i < n; i++ {
        w.Write(newline)
        w.Write(<-c.Send)
      }
      if err := w.Close(); err != nil {
        return
      }
    case <-ticker.C:
      c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
      if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
        return
      }
    }
	}
}