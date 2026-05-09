package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"whalebone/repository"
)

func TestSaveHandler(t *testing.T) {
	// arrange
	repo := repository.NewPersonRepositoryMock()
	handler := NewPersonHandler(repo)

	body, _ := json.Marshal(SavePersonRequest{
		ExternalID:  "123e4567-e89b-12d3-a456-426614174000",
		Name:        "John Doe",
		Email:       "john@example.com",
		DateOfBirth: "1990-01-01T00:00:00+00:00",
	})

	req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// act
	handler.PostPerson(rec, req)

	// assert
	if rec.Code != http.StatusCreated {
		t.Errorf("expected http status code 201, got %d", rec.Code)
	}
	if repo.SaveCalled != 1 {
		t.Errorf("expected 1 call to Save, got %d", repo.SaveCalled)
	}
}
