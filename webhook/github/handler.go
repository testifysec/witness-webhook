package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/attestation/githubwebhook"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/in-toto/go-witness/dsse"
	"github.com/testifysec/witness-webhook/webhook"
)

type Handler struct {
	secret []byte
	signer cryptoutil.Signer
}

func New(opts ...Option) (webhook.Handler, error) {
	h := &Handler{}
	errs := []error{}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("could not create github webhook hanlder: %w", errors.Join(errs...))
	}

	if len(h.secret) == 0 {
		return nil, fmt.Errorf("secret is required")
	}

	if h.signer == nil {
		return nil, fmt.Errorf("signer is required")
	}

	return h, nil
}

func (h *Handler) HandleRequest(req *http.Request) (dsse.Envelope, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return dsse.Envelope{}, fmt.Errorf("could not read request body: %w", err)
	}

	results, err := witness.RunWithExports(
		"webhook",
		witness.RunWithAttestors([]attestation.Attestor{
			githubwebhook.New(
				githubwebhook.WithBody(body),
				githubwebhook.WithRecievedSignature(req.Header.Get("X-Hub-Signature-256")),
				githubwebhook.WithSecret(h.secret),
			),
		}),
		witness.RunWithSigners(h.signer),
	)

	if err != nil {
		return dsse.Envelope{}, fmt.Errorf("could not create attestation: %w", err)
	}

	if len(results) == 0 {
		return dsse.Envelope{}, fmt.Errorf("no attestation created")
	}

	return results[0].SignedEnvelope, nil
}
