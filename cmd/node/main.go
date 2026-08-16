package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/anjuajay2468-glitch/cloudscale/internal/storage"
)

func main() {
	store, err := storage.NewStore("./data")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "CloudScale node is healthy")
	})

	http.HandleFunc("/objects/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[len("/objects/"):]

		if name == "" {
			http.Error(w, "object name is required", http.StatusBadRequest)
			return
		}

		switch r.Method {

		case http.MethodPut:
			data := make([]byte, r.ContentLength)

			_, err := r.Body.Read(data)
			if err != nil && len(data) == 0 {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}

			err = store.Put(name, data)
			if err != nil {
				http.Error(w, "failed to store object", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			fmt.Fprintln(w, "Object stored successfully")

		case http.MethodGet:
			data, err := store.Get(name)
			if err == storage.ErrNotFound {
				http.Error(w, "object not found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, "failed to read object", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)

		case http.MethodDelete:
			err := store.Delete(name)
			if err == storage.ErrNotFound {
				http.Error(w, "object not found", http.StatusNotFound)
				return
			}

			if err != nil {
				http.Error(w, "failed to delete object", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("CloudScale storage node running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}