package splitflap

import (
	"errors"
	"github.com/bep/debounce"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	mqtt  mqtt.Client
	topic string
	qos   byte

	previousMessage string
	currentMessage  string
	messageLock     sync.Mutex
	debounce        func(func())
}

type ClientConfig struct {
	Host     string
	Username string
	Password string
	Topic    string
	Qos      byte
}

func ClientConfigFromEnv() (*ClientConfig, error) {
	url := os.Getenv("MQTT_URL")
	if url == "" {
		return nil, errors.New("no MQTT_URL provided")
	}
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")
	topic := os.Getenv("MQTT_TOPIC")
	if topic == "" {
		return nil, errors.New("no MQTT_TOPIC provided")
	}
	qos := os.Getenv("MQTT_QOS")
	qosByte := byte(0)

	qos = strings.ReplaceAll(qos, " ", "")
	if qos == "" {
		slog.Info("No MQTT_QOS provided, defaulting to 0 (fire and forget)")
	} else if qos == "1" {
		qosByte = 1
	} else if qos == "2" {
		qosByte = 2
	} else {
		slog.Error("MQTT_QOS provided was not 0, 1, or 2. Defaulting to 0 (fire and forget)")
	}
	return &ClientConfig{
		Host:     url,
		Username: username,
		Password: password,
		Topic:    topic,
		Qos:      qosByte,
	}, nil
}

func NewSplitflapClient(cfg ClientConfig) Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Host)
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}
	return Client{
		mqtt:  mqtt.NewClient(opts),
		topic: cfg.Topic,
		qos:   cfg.Qos,

		previousMessage: "",
		currentMessage:  "",
		messageLock:     sync.Mutex{},
		debounce:        debounce.New(time.Millisecond * 250),
	}
}

func (c *Client) Run(kill <-chan struct{}, messages <-chan string) {
	err := c.connect()
	if err != nil {
		slog.Error(err.Error())
		return
	} else {
		slog.Info("Successful connection to MQTT broker")
	}
	defer c.disconnect()

	for {
		select {
		case <-kill:
			return
		case msg := <-messages:
			c.messageLock.Lock()
			c.currentMessage = msg
			c.messageLock.Unlock()
			// debounce the publish so we're not constantly spamming MQTT (and thus our splitflap) with messages,
			// when we could consolidate them into a single send
			c.debounce(c.publishCurrentMessage)
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (c *Client) connect() error {
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *Client) disconnect() {
	c.mqtt.Disconnect(250)
}

func (c *Client) publishCurrentMessage() {
	c.messageLock.Lock()
	c.publish(c.currentMessage)
	c.messageLock.Unlock()
}

func (c *Client) publish(message string) {
	if message == c.previousMessage {
		return
	}
	slog.Info("Publishing MQTT message", "topic", c.topic, "qos", c.qos, "message", message)
	token := c.mqtt.Publish(c.topic, c.qos, false, message)
	token.Wait()
	c.previousMessage = message
}
