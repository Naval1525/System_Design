package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"url-shortener/internal/service"

	"github.com/gorilla/mux"
)

type URLHandler struct {
	Service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{Service: service}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {

	var req ShortenRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	shortCode, err := h.Service.ShortenURL(req.URL)
	if err != nil {
		fmt.Println("ERROR:", err)
		http.Error(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{
		ShortURL: "http://localhost:8080/" + shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	code := params["code"]

	originalURL, err := h.Service.GetOriginalURL(code)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
