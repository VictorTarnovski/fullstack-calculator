// Package httpx is the transport kit shared by every HTTP handler: the
// response envelope, the problem detail, and the wrapper that turns an
// error-returning handler into an http.HandlerFunc.
//
// It knows nothing about any domain. A feature package maps its own errors to
// problem details with a Classifier and composes them with NewErrorMapper,
// which is what keeps the kit reusable.
package httpx

import (
	"log/slog"
	"net/http"
)

// RFC 9110 section anchors, used as the Type of a problem detail. A caller can
// follow the URL to read what the status means.
const (
	TypeBadRequest          = "https://datatracker.ietf.org/doc/html/rfc9110#section-15.5.1"
	TypeInternalServerError = "https://datatracker.ietf.org/doc/html/rfc9110#section-15.6.1"
)

// HandlerFunc is an http.HandlerFunc that may fail. Returning an error is the
// only way a handler reports failure; it never writes a status itself.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Classifier maps an error to a ProblemDetail, reporting false for an error it
// does not recognize. Its Instance field is left zero-valued by classifiers;
// Kit.Wrap fills it in from the request path.
type Classifier func(err error) (ProblemDetail, bool)

// NewErrorMapper composes classifiers into one, trying each in order and
// returning the first match. An error no classifier recognizes stays
// unclassified, which Kit.Wrap reports as a 500.
func NewErrorMapper(classifiers ...Classifier) Classifier {
	all := make([]Classifier, len(classifiers))
	copy(all, classifiers)

	return func(err error) (ProblemDetail, bool) {
		for _, classify := range all {
			if problem, ok := classify(err); ok {
				return problem, ok
			}
		}

		return ProblemDetail{}, false
	}
}

type Kit struct {
	Log      *slog.Logger
	Classify Classifier
}

// Wrap adapts an error-returning handler to net/http. An unclassified error is
// logged and reported as a bare 500: it is a bug rather than a caller mistake,
// so its text is kept out of the response.
func (k *Kit) Wrap(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			return
		}

		problem, ok := k.Classify(err)
		if !ok {
			logRequestError(k.Log, r, err)
			problem = NewProblemDetail(TypeInternalServerError, http.StatusInternalServerError, "")
		}

		problem.Instance = r.URL.Path

		if err := writeProblem(w, problem); err != nil {
			logRequestError(k.Log, r, err)
		}
	}
}

func logRequestError(logger *slog.Logger, r *http.Request, err error) {
	logger.Error(r.Method+" "+r.URL.Path, "err", err)
}
