package provider

type ProviderIface interface {
	AddSubscriber(chan any)
	Start() error
	Stop()
}
