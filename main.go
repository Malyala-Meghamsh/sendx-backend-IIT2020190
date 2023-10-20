package main

import (
	"net/http"
	routes "sendx/Routes"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/query/", routes.Serve).Methods("GET")
	routes.Main()
	http.Handle("/", router)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
