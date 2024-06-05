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
	"os"

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

				WithSecretFile(val)(githubHandler)
				return githubHandler, nil
			},
		),
	)
}

type Option func(*Handler) error

func WithSecretFile(secretFilePath string) Option {
	return func(h *Handler) error {
		if len(secretFilePath) == 0 {
			return nil
		}

		secret, err := os.ReadFile(secretFilePath)
		if err != nil {
			return fmt.Errorf("could not read webhook secret from file: %w", err)
		}

		h.secret = secret
		return nil
	}
}
