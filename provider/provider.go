package provider

type ProviderIface interface {
	AddSubscriber(chan any) string
	RemoveSubscriber(id string)
	Start() error
	Stop()
}
