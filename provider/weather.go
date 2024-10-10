package provider

type WeatherProvider struct {
	subscribers []chan any
}

func (wp *WeatherProvider) AddSubscriber(s chan any) {

}

func (wp *WeatherProvider) Start() error {
	go func() {

	}()
	return nil
}

func (wp *WeatherProvider) Stop() {

}
