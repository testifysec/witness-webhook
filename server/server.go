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

package server

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/in-toto/go-witness/cryptoutil"
	"github.com/in-toto/go-witness/dsse"
	"github.com/in-toto/go-witness/registry"
	"github.com/in-toto/go-witness/signer"
	"github.com/in-toto/go-witness/signer/kms"
	"github.com/testifysec/witness-webhook/config"
	"github.com/testifysec/witness-webhook/webhook"
)

type Server struct {
	r                *mux.Router
	attestationDir   string
	archivistaClient *archivista.Client
}

func New(ctx context.Context, config config.Config) (Server, error) {
	s := Server{
		r: mux.NewRouter(),
	}

	if len(config.ArchivistaUrl) > 0 {
		s.archivistaClient = archivista.New(config.ArchivistaUrl)
	}

	if len(config.AttestationDirectory) > 0 {
		if err := canWriteToDirectory(config.AttestationDirectory); err != nil {
			return s, fmt.Errorf("could not write to attestation directory: %w", err)
		}

		s.attestationDir = config.AttestationDirectory
	}

	for name, webhookConfig := range config.Webhooks {
		signerProvider, err := signer.NewSignerProviderFromConfigMap(webhookConfig.Signer, webhookConfig.SignerOptions)
		if err != nil {
			return s, fmt.Errorf("could not create signer provider for webhook %v: %w", name, err)
		}

		// kms is kind of a special case since each kms provider has it's own configs one layer lower than the
		// provider itself. so.. we forward on the options from our config to be loaded into the specified provider.
		if webhookConfig.Signer == "kms" {
			if err := applyKmsSettings(signerProvider, webhookConfig.SignerOptions); err != nil {
				return s, fmt.Errorf("could not apply kms signer settings for webhook %v: %w", name, err)
			}
		}

		signer, err := signerProvider.Signer(ctx)
		if err != nil {
			return s, fmt.Errorf("could not create signer for webhook %v: %w", name, err)
		}

		handler, err := webhook.NewWebhookHandlerFromConfigMap(webhookConfig.Type, webhookConfig.Options)
		if err != nil {
			return s, fmt.Errorf("could not create handler for webhook %v: %w", name, err)
		}

		handlerFunc, err := s.createHttpHandler(name, handler, signer)
		if err != nil {
			return s, fmt.Errorf("could not create handler func for webhook %v: %w", name, err)
		}

		s.r.HandleFunc(fmt.Sprintf("/%v", name), handlerFunc).Methods("POST")
	}

	return s, nil
}

func (s Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.r.ServeHTTP(w, req)
}

func (s Server) createHttpHandler(name string, h webhook.Handler, signer cryptoutil.Signer) (http.HandlerFunc, error) {
	if h == nil {
		return nil, fmt.Errorf("webhook handler is required")
	}

	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("request for webhook %v received\n", name)
		attestor, err := h.HandleRequest(req)
		if err != nil {
			log.Printf("could not handle request for webhook %v: %v\n", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if attestor == nil {
			log.Printf("no attestor returned for webhook %v, skipping event", name)
			return
		}

		results, err := witness.RunWithExports(
			"webhook",
			witness.RunWithAttestors([]attestation.Attestor{attestor}),
			witness.RunWithSigners(signer),
		)

		if err != nil {
			log.Printf("could not create attestation for webhook %v: %v\n", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(results) == 0 {
			log.Printf("no attestation in results for webhook %v: %v\n", name, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if s.archivistaClient != nil {
			if gitoid, err := s.archivistaClient.Store(req.Context(), results[0].SignedEnvelope); err != nil {
				log.Printf("could not store attestation in archivista for webhook %v: %v\n", name, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			} else {
				log.Printf("attestation stored in archivista for webhook %v with gitoid %v\n", name, gitoid)
			}
		}

		if len(s.attestationDir) > 0 {
			filePath, err := writeAttestationToDisk(results[0].SignedEnvelope, s.attestationDir)
			if err != nil {
				log.Printf("could not write attestation to disk: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			log.Printf("attestation written to %v\n", filePath)
		}

	}, nil
}

func applyKmsSettings(sp signer.SignerProvider, signerConfig map[string]any) error {
	ksp, ok := sp.(*kms.KMSSignerProvider)
	if !ok {
		return fmt.Errorf("provided signer provider is not a kms signer provider")
	}

	providerName := strings.Split(ksp.Reference, ":")[0]
	providerOpts, ok := ksp.Options[fmt.Sprintf("kms-%v", providerName)]
	if !ok {
		return fmt.Errorf("no options found for kms %v", providerName)
	}

	// find just the config values that start with providerName-
	specificConfig := make(map[string]any)
	for name, val := range signerConfig {
		prefix := fmt.Sprintf("%v-", providerName)
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		specificConfig[strings.TrimPrefix(name, prefix)] = val
	}

	opts := providerOpts.Init()
	if _, err := registry.SetDefaultVals[signer.SignerProvider](ksp, opts); err != nil {
		return fmt.Errorf("could not set default options for kms provider %v: %w", providerName, err)
	}

	if _, err := registry.SetOptionsFromConfigMap[signer.SignerProvider](ksp, providerOpts.Init(), specificConfig); err != nil {
		return fmt.Errorf("could not set options for kms provider %v: %w", providerName, err)
	}

	return nil
}

func canWriteToDirectory(dir string) error {
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		return fmt.Errorf("error creating test file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("error writing test file: %w", err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return fmt.Errorf("error removing test file: %w", err)
	}

	return nil
}

func writeAttestationToDisk(env dsse.Envelope, dir string) (string, error) {
	envBytes, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("could not marshal attestation to json: %w", err)
	}

	ds, err := cryptoutil.CalculateDigestSetFromBytes(envBytes, []cryptoutil.DigestValue{{Hash: crypto.SHA256, GitOID: true}})
	if err != nil {
		return "", fmt.Errorf("could not calculate gitoid of envelope: %w", err)
	}

	filePath := filepath.Join(dir, strings.TrimPrefix(fmt.Sprintf("%v.json", ds[cryptoutil.DigestValue{Hash: crypto.SHA256, GitOID: true}]), "gitoid:blob:sha256:"))
	if err := os.WriteFile(filePath, envBytes, 0644); err != nil {
		return "", fmt.Errorf("could not write attestation to file: %w", err)
	}

	return filePath, nil
}
