package github

import (
	"fmt"
	"os"

	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/in-toto/go-witness/registry"
	"github.com/testifysec/witness-webhook/webhook"
)

func init() {
	webhook.Register(
		"github",
		func() webhook.Handler {
			return &Handler{}
		},
		registry.StringConfigOption(
			"secret-file-path",
			"Path to the file containing the Github Webhook secret",
			"",
			func(h webhook.Handler, val string) (webhook.Handler, error) {
				githubHandler, ok := h.(*Handler)
				if !ok {
					return h, fmt.Errorf("received webhook handler is not a github handler")
				}

				if err := WithSecretFile(val)(githubHandler); err != nil {
					return h, fmt.Errorf("could not set github secret file: %w", err)
				}

				return githubHandler, nil
			},
		),
	)
}

type Option func(*Handler) error

func WithSecretFile(secretFilePath string) Option {
	return func(h *Handler) error {
		secretBytes, err := os.ReadFile(secretFilePath)
		if err != nil {
			return fmt.Errorf("could not load github secret from file: %w", err)
		}

		h.secret = secretBytes
		return nil
	}
}

func WithSigner(signer cryptoutil.Signer) Option {
	return func(h *Handler) error {
		h.signer = signer
		return nil
	}
}
