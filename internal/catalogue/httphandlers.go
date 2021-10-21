package catalogue

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type httpHandler struct {
	service *Service
}

func Setup(router *chi.Mux, service *Service) {
	handler := &httpHandler{service: service}
	router.Group(func(group chi.Router) {
		group.Post("/api/v1/catalogue", handler.InsertFilm)
		group.Get("/api/v1/catalogue/{uuid}", handler.GetFilm)
		group.Put("/api/v1/catalogue/{uuid}", handler.UpdateFilm)
	})
}

func (h httpHandler) InsertFilm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filmRequest := &Film{}
	if err := json.NewDecoder(r.Body).Decode(filmRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	film, err := h.service.InsertFilm(ctx, *filmRequest)
	if err != nil {
		log.Println("ERROR: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(film)
}

func (h httpHandler) GetFilm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filmUUID := chi.URLParam(r, "uuid")
	if filmUUID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	film, err := h.service.GetFilm(ctx, filmUUID)
	if err == ErrNoFilmFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("ERROR: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(film)
}

func (h httpHandler) UpdateFilm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filmUUID := chi.URLParam(r, "uuid")
	if filmUUID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	filmRequest := &Film{}
	if err := json.NewDecoder(r.Body).Decode(filmRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	film, err := h.service.UpdateFilm(ctx, filmUUID, *filmRequest)
	if err == ErrNoFilmFound {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("ERROR: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(film)
}
