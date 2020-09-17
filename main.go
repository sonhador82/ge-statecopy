package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
	"github.com/sonhador82/ge-statecopy/data"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type ChangeState struct {
	From string `json:"src_user_id"`
	To   string `json:"dst_user_id"`
}

func main() {
	staticToken := os.Getenv("X_TOKEN")
	tableName := os.Getenv("DYNAMODB_TABLE")
	if staticToken == "" || tableName == "" {
		panic("Specify env vars X_TOKEN und DYNAMODB_TABLE")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	mux := http.NewServeMux()
	mux.HandleFunc("/transfer_user_id", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-TOKEN")

		if token != staticToken {
			http.Error(w, "Access Denied", 403)
			return
		}

		var cs ChangeState
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&cs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("Copy from, to", cs.From, cs.To)
		err = data.CopyState(svc, tableName, cs.From, cs.To)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %s", err), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\": \"ok\"}"))
	})

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://dirty.mib.neurohive.net"},
		AllowCredentials: true,
		Debug:            true,
		AllowedHeaders:   []string{"X-TOKEN", "Content-Type"},
	}).Handler(mux)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
