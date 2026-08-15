package expression

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"calculator/internal/httpx"
)

// newTestServer wires the handler exactly as cmd/api does, so the tests
// exercise the real Kit.Wrap error path rather than a stand-in for it.
func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	mux := http.NewServeMux()
	kit := &httpx.Kit{
		Log:      slog.New(slog.DiscardHandler),
		Classify: httpx.NewErrorMapper(ClassifyError),
	}

	(&Handler{}).RegisterRoutes(mux, kit)

	return mux
}

func post(t *testing.T, srv http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/evaluations", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	return rec
}

func TestEvaluateSuccess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected float32
	}{
		{
			name:     "addition without spaces",
			body:     "7+7",
			expected: 14,
		},
		{
			name:     "precedence without spaces",
			body:     "120+34×5",
			expected: 290,
		},
		{
			name:     "decimals",
			body:     "1.5+2.25",
			expected: 3.75,
		},
		{
			name:     "single atom",
			body:     "12",
			expected: 12,
		},
	}

	srv := newTestServer(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := post(t, srv, tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("POST %q status = %d, want %d (body %s)", tt.body, rec.Code, http.StatusOK, rec.Body)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}

			var env struct {
				Data struct {
					Result float32 `json:"result"`
				} `json:"data"`
			}

			if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
				t.Fatalf("decoding envelope: %v", err)
			}

			if env.Data.Result != tt.expected {
				t.Errorf("POST %q result = %v, want %v", tt.body, env.Data.Result, tt.expected)
			}
		})
	}
}

func TestEvaluateProblemDetail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
		expectedDetail string
	}{
		{
			name:           "operator without right operand",
			body:           "7+",
			expectedStatus: http.StatusBadRequest,
			expectedDetail: ErrUnexpectedToken.Error(),
		},
		{
			name:           "empty expression",
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedDetail: ErrUnexpectedToken.Error(),
		},
		{
			name:           "division by zero",
			body:           "6÷0",
			expectedStatus: http.StatusBadRequest,
			expectedDetail: ErrDivisionByZero.Error(),
		},
		{
			name:           "multiple decimal points",
			body:           "7..5",
			expectedStatus: http.StatusBadRequest,
			expectedDetail: ErrInvalidNumber.Error(),
		},
	}

	srv := newTestServer(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := post(t, srv, tt.body)

			if rec.Code != tt.expectedStatus {
				t.Fatalf("POST %q status = %d, want %d (body %s)", tt.body, rec.Code, tt.expectedStatus, rec.Body)
			}

			if got := rec.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/problem+json")
			}

			var problem httpx.ProblemDetail
			if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
				t.Fatalf("decoding problem detail: %v", err)
			}

			if problem.Status != tt.expectedStatus {
				t.Errorf("problem.Status = %d, want %d", problem.Status, tt.expectedStatus)
			}

			if problem.Title != http.StatusText(tt.expectedStatus) {
				t.Errorf("problem.Title = %q, want %q", problem.Title, http.StatusText(tt.expectedStatus))
			}

			if problem.Type == "" {
				t.Error("problem.Type is empty, want an RFC reference")
			}

			if problem.Instance != "/evaluations" {
				t.Errorf("problem.Instance = %q, want %q", problem.Instance, "/evaluations")
			}

			if !strings.Contains(problem.Detail, tt.expectedDetail) {
				t.Errorf("problem.Detail = %q, want it to contain %q", problem.Detail, tt.expectedDetail)
			}
		})
	}
}

// An unrecognized token is unreachable from the keypad, so it is deliberately
// left unclassified and surfaces as a logged 500 rather than a 400.
func TestEvaluateUnclassifiedErrorIsInternal(t *testing.T) {
	t.Parallel()

	rec := post(t, newTestServer(t), "1*2")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusInternalServerError, rec.Body)
	}

	var problem httpx.ProblemDetail
	if err := json.NewDecoder(rec.Body).Decode(&problem); err != nil {
		t.Fatalf("decoding problem detail: %v", err)
	}

	// The unclassified path must not leak the internal error text.
	if problem.Detail != "" {
		t.Errorf("problem.Detail = %q, want it empty", problem.Detail)
	}
}
