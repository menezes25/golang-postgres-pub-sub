package websocket

import (
	"fmt"
	"net"
	"net/http"
	gopostgrespubsub "postgres_pub_sub"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Event struct {
	Type    gopostgrespubsub.EventType
	Payload interface{}
}

type WsManager struct {
	EventTypeConns map[gopostgrespubsub.EventType][]net.Conn
}

func New(eventChan <-chan gopostgrespubsub.Event) *WsManager {
	w := &WsManager{
		EventTypeConns: make(map[gopostgrespubsub.EventType][]net.Conn),
	}

	go func() {
		for event := range eventChan {
			fmt.Printf("got message to send to client: %v\n", event)

			for i, c := range w.EventTypeConns[event.Type] {
				println("sending to conn: ", i)
				err := wsutil.WriteServerMessage(c, ws.OpText, []byte(event.Payload))
				if err != nil {
					return
				}
			}
		}
	}()

	return w
}

func (wm *WsManager) MakeListenToEventsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		paths := strings.Split(r.RequestURI, "/")
		event := paths[len(paths)-1]

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if _, ok := wm.EventTypeConns[gopostgrespubsub.EventType(event)]; ok {
			wm.EventTypeConns[gopostgrespubsub.EventType(event)] = append(wm.EventTypeConns[gopostgrespubsub.EventType(event)], conn)
		} else {
			wm.EventTypeConns[gopostgrespubsub.EventType(event)] = []net.Conn{conn}
		}

		go func() {
			defer conn.Close()

			for {
				msg, _, err := wsutil.ReadClientData(conn)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				println("read client data msg: ", string(msg))
			}
		}()
	}
}
