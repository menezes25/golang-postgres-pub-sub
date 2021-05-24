package websocket

import (
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/julienschmidt/httprouter"
)

//TODO: EventTypes documents, users

type EventType string

type WsManager struct {
	Conns             []net.Conn
	MapConnEventTypes map[EventType][]net.Conn
}

func New(writeToConnectionsChan chan string) *WsManager {
	w := &WsManager{
		Conns: make([]net.Conn, 0),
	}

	go func() {
		for write := range writeToConnectionsChan {
			println("got message to send to client: ", write)
			//TODO: EventType from write
			//TODO: w.MapConnEventTypes[eventType] on range
			// var eventType EventType = "EventType"

			for i, c := range w.Conns {
				println("sending to conn: ", i)
				err := wsutil.WriteServerMessage(c, ws.OpText, []byte(write))
				if err != nil {
					return
				}
			}
		}
	}()

	return w
}

func (wm *WsManager) MakeListenToEventsHandler() httprouter.Handle {
	//TODO: requisitar tipo de evento no request
	type Request struct {
		EventsToListen []string
	}
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		wm.Conns = append(wm.Conns, conn)

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
