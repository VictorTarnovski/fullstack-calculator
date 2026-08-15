package httpx

import (
	"encoding/json"
	"net/http"
)

// ProblemDetail is an RFC 9457 problem detail, the response body for every
// failed request. Type is a URL identifying the problem kind, Title is the
// human-readable status name, and Instance is the path of the request that
// failed. Handlers leave Instance zero-valued; Kit.Wrap fills it in.
type ProblemDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

type ProblemDetailOption func(*ProblemDetail)

// WithDetail attaches an explanation of this specific occurrence. Only set it
// from errors that are safe to show a caller: the detail is sent verbatim.
func WithDetail(detail string) ProblemDetailOption {
	return func(p *ProblemDetail) { p.Detail = detail }
}

// NewProblemDetail builds a problem whose Title is derived from status, so the
// two can never disagree.
func NewProblemDetail(problemType string, status int, instance string, opts ...ProblemDetailOption) ProblemDetail {
	p := ProblemDetail{
		Type:     problemType,
		Title:    http.StatusText(status),
		Status:   status,
		Instance: instance,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

// Envelope wraps every successful response body, so a payload is always found
// under the same key regardless of the endpoint.
type Envelope struct {
	Data any `json:"data"`
}

func NewEnvelope(data any) Envelope {
	return Envelope{Data: data}
}

// WriteJSON writes a successful response. It is the only way a handler should
// produce a body, so that the envelope is never bypassed.
func WriteJSON(w http.ResponseWriter, status int, resp Envelope) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(resp)
}

// writeProblem is unexported because handlers never choose to write a problem:
// they return an error and Kit.Wrap classifies it.
func writeProblem(w http.ResponseWriter, p ProblemDetail) error {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)

	return json.NewEncoder(w).Encode(p)
}
