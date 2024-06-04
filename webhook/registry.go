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
