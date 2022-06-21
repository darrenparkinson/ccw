package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/darrenparkinson/ccw"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type application struct {
	ccwc *ccw.Client
}

func main() {

	var username, password, clientID, clientSecret string
	mustMapEnv(&username, "CCW_USERNAME")
	mustMapEnv(&password, "CCW_PASSWORD")
	mustMapEnv(&clientID, "CCW_CLIENTID")
	mustMapEnv(&clientSecret, "CCW_CLIENTSECRET")
	c, err := ccw.NewClient(username, password, clientID, clientSecret, nil)
	if err != nil {
		log.Fatal(err)
	}
	app := application{ccwc: c}

	r := mux.NewRouter()
	// r.HandleFunc("/quotes", QuotesHandler)
	r.HandleFunc("/quotes/{dealid}", app.QuoteHandler).Methods(http.MethodGet)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	srv := &http.Server{Addr: ":8000", Handler: corsMiddleware.Handler(r)}
	log.Println("starting server on", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

func (app *application) QuoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	qr, err := app.ccwc.QuoteService.AcquireByDealID(context.Background(), vars["dealid"])
	if err != nil {
		if errors.Is(err, ccw.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, ccw.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, ccw.ErrUnauthorized) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if errors.Is(err, ccw.ErrForbidden) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	js, err := json.MarshalIndent(qr, "", "\t")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(js)

}

func mustMapEnv(target *string, envKey string) {
	v := os.Getenv(envKey)
	if v == "" {
		log.Fatalf("environment variable %q not set", envKey)
	}
	*target = v
}
