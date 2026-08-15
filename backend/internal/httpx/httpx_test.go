package httpx

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKitWrapSuccessWritesNothingExtra(t *testing.T) {
	t.Parallel()

	kit := &Kit{Log: slog.New(slog.DiscardHandler)}

	handler := kit.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return WriteJSON(w, http.StatusOK, NewEnvelope("ok"))
	})

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestKitWrapClassifiedError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("bad input")

	classify := func(err error) (ProblemDetail, bool) {
		if errors.Is(err, sentinel) {
			return NewProblemDetail(TypeBadRequest, http.StatusBadRequest, "", WithDetail(err.Error())), true
		}

		return ProblemDetail{}, false
	}

	kit := &Kit{Log: slog.New(slog.DiscardHandler), Classify: classify}

	handler := kit.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return sentinel
	})

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, "/evaluations", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/problem+json")
	}

	var problem ProblemDetail
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decoding problem detail: %v", err)
	}

	if problem.Instance != "/evaluations" {
		t.Errorf("problem.Instance = %q, want %q", problem.Instance, "/evaluations")
	}

	if problem.Detail != sentinel.Error() {
		t.Errorf("problem.Detail = %q, want %q", problem.Detail, sentinel.Error())
	}
}

func TestKitWrapUnclassifiedErrorIsInternal(t *testing.T) {
	t.Parallel()

	classify := func(err error) (ProblemDetail, bool) { return ProblemDetail{}, false }

	kit := &Kit{Log: slog.New(slog.DiscardHandler), Classify: classify}

	handler := kit.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("boom")
	})

	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var problem ProblemDetail
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decoding problem detail: %v", err)
	}

	// The internal error text must never leak to the caller.
	if problem.Detail != "" {
		t.Errorf("problem.Detail = %q, want empty", problem.Detail)
	}
}

func TestNewErrorMapper(t *testing.T) {
	t.Parallel()

	errA := errors.New("a")
	errB := errors.New("b")

	classifyA := func(err error) (ProblemDetail, bool) {
		if errors.Is(err, errA) {
			return NewProblemDetail(TypeBadRequest, http.StatusBadRequest, ""), true
		}

		return ProblemDetail{}, false
	}

	classifyB := func(err error) (ProblemDetail, bool) {
		if errors.Is(err, errB) {
			return NewProblemDetail(TypeBadRequest, http.StatusUnprocessableEntity, ""), true
		}

		return ProblemDetail{}, false
	}

	mapper := NewErrorMapper(classifyA, classifyB)

	if problem, ok := mapper(errA); !ok || problem.Status != http.StatusBadRequest {
		t.Errorf("mapper(errA) = %+v, %v", problem, ok)
	}

	if problem, ok := mapper(errB); !ok || problem.Status != http.StatusUnprocessableEntity {
		t.Errorf("mapper(errB) = %+v, %v", problem, ok)
	}

	if _, ok := mapper(errors.New("unmapped")); ok {
		t.Error("mapper matched an error no classifier recognizes")
	}
}
