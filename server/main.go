package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// App contains the router used to route HTTP requests
type App struct {
	Router *mux.Router
	Port   string
}

func main() {
	a := App{}
	a.Init()
	http.ListenAndServe(":8081", a.Router)
}

// Init the application with an http request router
func (a *App) Init() {
	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte("Hello World"))
	})
}

// Serve the application
func (a *App) Serve() {
	if port := os.Getenv("PORT"); len(port) > 0 {
		a.Port = string(port)
	} else {
		a.Port = "8081"
	}

	http.ListenAndServe(a.Port, a.Router)
}
