package handler

import "github.com/gorilla/mux"

func SetupRouter(handler *URLHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/shorten", handler.ShortenURL).Methods("POST")
	router.HandleFunc("/{code}", handler.RedirectURL).Methods("GET")
	return router
}
