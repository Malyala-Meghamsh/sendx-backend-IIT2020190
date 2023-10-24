package main

import (
	"fmt"
	"net/http"
	"sendx/controllers"
)

func main() {
	controllers.Main()
	//Starting server and routes
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/query", controllers.CrawlHandler)
	http.HandleFunc("/api/setCrawlers", controllers.SetCrawlers)
	http.HandleFunc("/api/setSpeed", controllers.SetSpeed)

	server_err := http.ListenAndServe(":3000", nil)
	fmt.Println("Server started on the port 3000....")
	fmt.Println("Follow this link : http://localhost:3000/")
	if server_err != nil {
		fmt.Println("Sorry Unable to Start The server")
		return
	}
}
