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

func NewSignerProvider(name string, opts ...func(Handler) (Handler, error)) (Handler, error) {
	return webhookRegistry.NewEntity(name, opts...)
}
