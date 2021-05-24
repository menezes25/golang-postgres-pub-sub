package rest

import (
	"encoding/json"
	"net/http"
	"postgres_pub_sub/postgres"

	"github.com/julienschmidt/httprouter"
)

func MakePostEventHandler(postgresCli *postgres.PostgresClient) httprouter.Handle {
	type EventRequest struct {
		Event string `json:"event,omitempty"`
	}

	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			// Set CORS headers
			header := w.Header()
			header.Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, DELETE, PUT")
			header.Set("Access-Control-Allow-Origin", "*")
			header.Set("Access-Control-Allow-Credentials", "true")
			header.Set("Access-Control-Allow-Headers", "*")

			w.WriteHeader(http.StatusNoContent)
			return
		}
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
