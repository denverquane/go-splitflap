package splitflap

import (
	"errors"
	"github.com/bep/debounce"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"log/slog"
	"time"
)

type OutMessage struct {
	payload string
}

type Client struct {
	serial   *usb_serial.Splitflap
	lastSent string
	debounce func(func())
}

func NewSplitflapClient() Client {
	return Client{
		serial:   nil,
		lastSent: "",
		debounce: debounce.New(time.Millisecond * 200),
	}
}

func (c *Client) Connect(port string, notify chan string) error {
	connection := usb_serial.NewSerialConnectionOnPort(port)
	if connection == nil {
		return errors.New("couldn't connect over USB")
	}
	sf := usb_serial.NewSplitflap(connection, notify)
	sf.Start()

	c.serial = sf
	return nil
}

func (c *Client) Run(outmessages <-chan OutMessage) {
	if c.serial == nil {
		slog.Error("Tried to start Client with a nil serial connection, exiting")
		return
	}
	for {
		select {
		case msg := <-outmessages:
			if msg.payload != c.lastSent {
				// debounce the publish so we're not constantly spamming our splitflap with messages,
				// when we could just consolidate them into a single send
				//c.debounce(func() {
				c.lastSent = msg.payload
				err := c.serial.SetTextWithMovement(msg.payload, usb_serial.ForceMovementNone)
				if err != nil {
					slog.Error(err.Error())
				}
				//})
			}
		}
	}
}
