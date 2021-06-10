package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/menezes25/golang-postgres-pub-sub/internal"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/julienschmidt/httprouter"
)

type GenericContract struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WsManager struct {
	EventTypeConns map[internal.EventType][]net.Conn
}

func New(ctx context.Context, eventChan <-chan internal.Event) *WsManager {
	wm := &WsManager{
		EventTypeConns: make(map[internal.EventType][]net.Conn),
	}

	go func() {
		for {
			select {
			case event := <-eventChan:
				fmt.Printf("got message to send to client: %v\n", event)

				for i, c := range wm.EventTypeConns[event.Type] {
					contract := GenericContract{
						Type:    event.Type.Type(),
						Payload: event.Payload,
					}

					cBytes, err := json.Marshal(contract)
					if err != nil {
						fmt.Fprintf(os.Stderr, "erro no Marshal. err: %v\n", err.Error())
						continue
					}

					err = wsutil.WriteServerMessage(c, ws.OpText, cBytes)
					if err != nil {
						fmt.Fprintf(os.Stderr, "erro na conn [%d] %s. %s retornando.\n", i, c.RemoteAddr(), err.Error())
						continue
					}

					fmt.Printf("Enviado com sucesso: conn [%d] %s.\n", i, c.RemoteAddr())
				}

			case <-time.After(2 * time.Second):
				wm.pingAndRemoveConnections()

			case <-ctx.Done():
				fmt.Printf("shutting down wsManager. Error: %s\n", ctx.Err().Error())
				wm.closeAllClientConnections()
				return
			}
		}
	}()

	return wm
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

		fmt.Printf("WS REQUEST [%s] listen: %s\n", r.RemoteAddr, topics)

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
	}
}

func validateListen(listenStr string) ([]internal.EventType, error) {
	if listenStr == "" {
		return nil, errors.New("a requisição deve informar no minimo um listen")
	}

	listens := strings.Split(listenStr, ",")
	evetntTypeList := make([]internal.EventType, 0)
	for _, listen := range listens {
		evetntTypeList = append(evetntTypeList, internal.EventType(listen))
	}

	return evetntTypeList, nil
}

func (wm *WsManager) pingAndRemoveConnections() {
	connToBeClosed := make(map[net.Conn]interface{}) // Set (abstract data type)
	activeConnMap := make(map[internal.EventType][]net.Conn)

	for eventType := range wm.EventTypeConns {
		for i, conn := range wm.EventTypeConns[eventType] {
			err := wsutil.WriteServerMessage(conn, ws.OpPing, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "PING CONN TO BE CLOSED [conn %d] [%s] ERRO '%s'\n", i, conn.RemoteAddr(), err.Error())
				connToBeClosed[conn] = nil
				continue
			}

			fmt.Printf("PING [conn %d] [%s] '%s'\n", i, conn.RemoteAddr(), eventType)
			if conns, found := activeConnMap[eventType]; found {
				activeConnMap[eventType] = append(conns, conn)
			} else {
				activeConnMap[eventType] = []net.Conn{conn}
			}
		}
	}

	for conn := range connToBeClosed {
		err := conn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			fmt.Fprintf(os.Stderr, "CLOSE CONN ERRO [conn %s] %s\n", conn.RemoteAddr(), err.Error())
			continue

		} else if err != nil && errors.Is(err, net.ErrClosed) {
			fmt.Printf("CLOSE CONN OK [conn %s] já estava fechada\n", conn.RemoteAddr())
			continue
		}

		fmt.Printf("CLOSE CONN OK [conn %s] fechada\n", conn.RemoteAddr())
	}

	wm.EventTypeConns = activeConnMap
}

func (wm *WsManager) closeAllClientConnections() {
	fmt.Printf("INFO fechando todas conexões websocket\n")
	for eventType := range wm.EventTypeConns {
		for _, conn := range wm.EventTypeConns[eventType] {
			err := conn.Close()
			if err != nil && !errors.Is(err, net.ErrClosed) {
				fmt.Fprintf(os.Stderr, "ERRO [conn %s] %s\n", conn.RemoteAddr(), err.Error())
				continue

			} else if err != nil && errors.Is(err, net.ErrClosed) {
				fmt.Printf("CLOSE CONN OK [conn %s] já estava fechada\n", conn.RemoteAddr())
				continue
			}

			fmt.Printf("CLOSE CONN OK [conn %s] fechada\n", conn.RemoteAddr())
		}
	}
}
