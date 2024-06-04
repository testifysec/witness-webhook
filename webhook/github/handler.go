package github

import (
	"fmt"
	"io"
	"net/http"

	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/attestation/githubwebhook"
	"github.com/testifysec/witness-webhook/webhook"
)

type Handler struct {
	secret []byte
}

func New(opts ...Option) (webhook.Handler, error) {
	h := &Handler{}
	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

func (h *Handler) HandleRequest(req *http.Request) (attestation.Attestor, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read request body: %w", err)
	}

	return githubwebhook.New(
		githubwebhook.WithBody(body),
		githubwebhook.WithRecievedSignature(req.Header.Get("X-Hub-Signature-256")),
		githubwebhook.WithSecret(h.secret),
	), nil

}
