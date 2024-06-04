package webhook

import (
	"net/http"

	"github.com/in-toto/go-witness/attestation"
)

type Handler interface {
	HandleRequest(*http.Request) (attestation.Attestor, error)
}
