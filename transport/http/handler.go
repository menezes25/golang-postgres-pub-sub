package http

import (
	"context"
	"net/http"

	"github.com/menezes25/golang-postgres-pub-sub/internal"
	"github.com/menezes25/golang-postgres-pub-sub/postgres"
	"github.com/menezes25/golang-postgres-pub-sub/transport/http/rest"
	"github.com/menezes25/golang-postgres-pub-sub/transport/http/websocket"

	"github.com/julienschmidt/httprouter"
)

type TransportHttp struct {
	router      *httprouter.Router
	postgresCli *postgres.PostgresClient
	wsManager   *websocket.WsManager
}

// check if the type satisfies the interface `http.Handler`
// zero-memory way to represent a pointer to struct.
var _ http.Handler = (*TransportHttp)(nil)

func NewTransportHttp(ctx context.Context, postgresCli *postgres.PostgresClient, fanInEvent <-chan internal.Event) (*TransportHttp, error) {
	wsManager := websocket.New(ctx, fanInEvent)
	tr := TransportHttp{
		router:      httprouter.New(),
		postgresCli: postgresCli,
		wsManager:   wsManager,
	}
	tr.setHandlers()
	return &tr, nil
}

// ServeHTTP faz o TransportHttp implementar a interface http.Handler
func (tr *TransportHttp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tr.router.ServeHTTP(w, r)
}

func (tr *TransportHttp) setHandlers() {
	tr.router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			// julienschmidt/httprouter calcula somente os metodos permitidos e armazena no header 'Allow'
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Headers", "Content-Type")
			header.Set("Access-Control-Allow-Origin", "*")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	tr.router.POST("/api/event", rest.MakePostEventHandler(tr.postgresCli))
	tr.router.GET("/ws/topic", tr.wsManager.MakeListenToEventsHandler())
}
