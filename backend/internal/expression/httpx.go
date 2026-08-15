package expression

import (
	"errors"
	"net/http"

	"calculator/internal/httpx"
)

type Handler struct{}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, kit *httpx.Kit) {
	mux.HandleFunc("POST /evaluations", kit.Wrap(h.Evaluate))
}

type evaluateResponse struct {
	Result float32 `json:"result"`
}

// Evaluate parses and evaluates the request body. The body is the expression
// itself as text/plain, not a field inside a JSON object: New already reads
// from an io.Reader, so the body is handed to it directly and never buffered.
//
// The operators are Unicode math characters, so a caller must send UTF-8; any
// other encoding reaches the lexer as unrecognized runes.
//
// Consuming the stream means the input is gone by the time the response is
// written, so the response carries only the result. The caller knows what it
// sent.
func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) error {
	expr, err := New(r.Body)
	if err != nil {
		return err
	}

	val, err := expr.Eval()
	if err != nil {
		return err
	}

	return httpx.WriteJSON(w, http.StatusOK, httpx.NewEnvelope(evaluateResponse{Result: val}))
}

// ClassifyError maps the errors a caller can cause to 400s. The errors left
// unclassified are deliberate: ErrUnknownOperator and ErrOperandCount mean the
// lexer and the evaluator disagree, and ErrUnrecognizedToken is unreachable
// from a keypad that only emits digits and the four operators. All three are
// bugs rather than bad input, so they surface as a logged 500.
func ClassifyError(err error) (httpx.ProblemDetail, bool) {
	switch {
	case errors.Is(err, ErrUnexpectedToken),
		errors.Is(err, ErrInvalidNumber),
		errors.Is(err, ErrDivisionByZero):
		return httpx.NewProblemDetail(
			httpx.TypeBadRequest,
			http.StatusBadRequest,
			"",
			httpx.WithDetail(err.Error()),
		), true
	}

	return httpx.ProblemDetail{}, false
}
