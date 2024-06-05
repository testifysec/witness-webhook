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

package webhook

import "github.com/in-toto/go-witness/registry"

var (
	webhookRegistry = registry.New[Handler]()
)

func Register(name string, factory registry.FactoryFunc[Handler], opts ...registry.Configurer) {
	webhookRegistry.Register(name, factory, opts...)
}

func RegistryEntries() []registry.Entry[Handler] {
	return webhookRegistry.AllEntries()
}

func NewWebhookHandler(name string, opts ...func(Handler) (Handler, error)) (Handler, error) {
	return webhookRegistry.NewEntity(name, opts...)
}

func NewWebhookHandlerFromConfigMap(name string, configMap map[string]any) (Handler, error) {
	return webhookRegistry.NewEntityFromConfigMap(name, configMap)
}
