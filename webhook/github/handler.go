// Copyright 2024 Witness Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	event := req.Header.Get("X-Github-Event")
	if event == "ping" {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read request body: %w", err)
	}

	return githubwebhook.New(
		githubwebhook.WithBody(body),
		githubwebhook.WithRecievedSignature(req.Header.Get("X-Hub-Signature-256")),
		githubwebhook.WithSecret(h.secret),
		githubwebhook.WithEvent(event),
	), nil

}
