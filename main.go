package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/maestre3d/sse-example/sse"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	b := sse.NewBroker(logger)
	defer b.Close()

	go func() {
		// Broker listen
		http.Handle("/event", b)
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
	}()

	r := mux.NewRouter()
	r.PathPrefix("/push").Methods(http.MethodPost).HandlerFunc(publishEventHandler(b))
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}

func publishEventHandler(b *sse.Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		body := struct {
			Message string `json:"message"`
			Consumer uint64 `json:"consumer_id"`
		}{
			r.PostFormValue("message"),
			0,
		}

		c, err := strconv.ParseUint(r.PostFormValue("consumer_id"), 10, 64)
		if err == nil {
			body.Consumer = c
		}

		res := struct {
			Message string `json:"message"`
		}{}
		if body.Message == "" {
			w.WriteHeader(http.StatusBadRequest)
			res.Message = "invalid value"
			_ = json.NewEncoder(w).Encode(res)
			return
		}

		b.Publish(sse.NewEvent([]byte(body.Message), body.Consumer))

		w.WriteHeader(http.StatusOK)
	}
}
