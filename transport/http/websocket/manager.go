package websocket

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	gopostgrespubsub "postgres_pub_sub"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/julienschmidt/httprouter"
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
				err := wsutil.WriteServerMessage(c, ws.OpText, []byte(event.Payload))
				if err != nil {
					fmt.Fprintf(os.Stderr, "erro na conn [%d] %s. %s retornando.\n", i, c.RemoteAddr(), err.Error())
					continue
				}

				fmt.Printf("Enviado com sucesso: conn [%d] %s.\n", i, c.RemoteAddr())
			}
		}
	}()

	return w
}

func (wm *WsManager) MakeListenToEventsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		topics, err := validateListen(r.URL.Query().Get("listen"))
		if err != nil {
			println(err.Error())
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("PubSubApp-Error", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Println(topics)

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		for _, topic := range topics {
			if conns, found := wm.EventTypeConns[topic]; found {
				wm.EventTypeConns[topic] = append(conns, conn)
			} else {
				wm.EventTypeConns[topic] = []net.Conn{conn}
			}
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

func validateListen(listenStr string) ([]gopostgrespubsub.EventType, error) {
	if listenStr == "" {
		return nil, errors.New("a requisição deve informar no minimo um listen")
	}

	listens := strings.Split(listenStr, ",")
	evetntTypeList := make([]gopostgrespubsub.EventType, 0)
	for _, listen := range listens {
		evetntTypeList = append(evetntTypeList, gopostgrespubsub.EventType(listen))
	}

	return evetntTypeList, nil
}
