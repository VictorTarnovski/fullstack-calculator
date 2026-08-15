package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()

	if err := WriteJSON(rec, http.StatusCreated, NewEnvelope(map[string]int{"result": 42})); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var env struct {
		Data struct {
			Result int `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decoding body: %v", err)
	}

	if env.Data.Result != 42 {
		t.Errorf("Data.Result = %d, want %d", env.Data.Result, 42)
	}
}

func TestNewProblemDetail(t *testing.T) {
	t.Parallel()

	problem := NewProblemDetail(TypeBadRequest, http.StatusBadRequest, "/evaluations")

	if problem.Title != http.StatusText(http.StatusBadRequest) {
		t.Errorf("Title = %q, want %q", problem.Title, http.StatusText(http.StatusBadRequest))
	}

	if problem.Detail != "" {
		t.Errorf("Detail = %q, want empty", problem.Detail)
	}

	if problem.Instance != "/evaluations" {
		t.Errorf("Instance = %q, want %q", problem.Instance, "/evaluations")
	}
}

func TestNewProblemDetailWithDetail(t *testing.T) {
	t.Parallel()

	problem := NewProblemDetail(TypeBadRequest, http.StatusBadRequest, "", WithDetail("division by zero"))

	if problem.Detail != "division by zero" {
		t.Errorf("Detail = %q, want %q", problem.Detail, "division by zero")
	}
}

func TestNewEnvelope(t *testing.T) {
	t.Parallel()

	env := NewEnvelope("payload")

	if env.Data != "payload" {
		t.Errorf("Data = %v, want %v", env.Data, "payload")
	}
}
