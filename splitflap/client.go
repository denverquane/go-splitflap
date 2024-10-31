package splitflap

import (
	"errors"
	"github.com/bep/debounce"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"log/slog"
	"sync"
	"time"
)

type Client struct {
	serial         *usb_serial.Splitflap
	currentMessage string
	messageLock    sync.Mutex
	debounce       func(func())
}

func NewSplitflapClient() Client {
	return Client{
		serial:         nil,
		currentMessage: "",
		messageLock:    sync.Mutex{},
		debounce:       debounce.New(time.Millisecond * 250),
	}
}

func (c *Client) Connect(port string) error {
	connection := usb_serial.NewSerialConnectionOnPort(port)
	if connection == nil {
		return errors.New("couldn't connect over USB")
	}
	sf := usb_serial.NewSplitflap(connection)
	sf.Start()

	c.serial = sf
	return nil
}

func (c *Client) Run(messages <-chan string) {
	if c.serial == nil {
		slog.Error("Tried to start Client with a nil serial connection, exiting")
		return
	}
	for {
		select {
		case msg := <-messages:
			c.messageLock.Lock()
			c.currentMessage = msg
			c.messageLock.Unlock()
			// debounce the publish so we're not constantly spamming our splitflap with messages,
			// when we could consolidate them into a single send
			c.debounce(c.publishCurrentMessage)
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (c *Client) publishCurrentMessage() {
	c.messageLock.Lock()
	err := c.serial.SetText(c.currentMessage)
	c.messageLock.Unlock()
	if err != nil {
		slog.Error(err.Error())
	}
}
