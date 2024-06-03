package webhook

import (
	"net/http"

	"github.com/in-toto/go-witness/dsse"
)

type Handler interface {
	HandleRequest(*http.Request) (dsse.Envelope, error)
}
