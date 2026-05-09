package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"whale/models"
	"whale/repository"

	"gorm.io/gorm"
)

type PersonHandler struct {
	repo repository.PersonRepository
}

func NewPersonHandler(repository repository.PersonRepository) *PersonHandler {
	return &PersonHandler{repo: repository}
}

func (h *PersonHandler) GetPerson(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := req.URL.Path
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	id, _ = strings.CutPrefix(id, "/")

	person, err := h.repo.GetByExternalID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

type SavePersonRequest struct {
	ExternalID  string `json:"external_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	DateOfBirth string `json:"date_of_birth"` // parse from ISO8601 string
}

func (h *PersonHandler) PostPerson(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var person SavePersonRequest
	if err := json.NewDecoder(req.Body).Decode(&person); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dateOfBirth, err := time.Parse(time.RFC3339, person.DateOfBirth)
	if err != nil {
		http.Error(w, "invalid dateOfBirth format, use ISO8601", http.StatusBadRequest)
		return
	}

	personRec := models.Person{
		ExternalID:  person.ExternalID,
		Name:        person.Name,
		Email:       person.Email,
		DateOfBirth: dateOfBirth,
	}

	err = h.repo.Save(&personRec)
	if err != nil {
		http.Error(w, "failed to save to db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
