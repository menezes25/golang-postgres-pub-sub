package rest

import (
	"encoding/json"
	"net/http"

	"github.com/menezes25/golang-postgres-pub-sub/postgres"

	"github.com/julienschmidt/httprouter"
)

func MakePostEventHandler(postgresCli *postgres.PostgresClient) httprouter.Handle {
	type EventRequest struct {
		Event string `json:"event,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var event EventRequest
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = postgresCli.AddEvent(r.Context(), event.Event)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("{}"))
	}
}
