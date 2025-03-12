package splitflap

import (
	"errors"
	gen "github.com/denverquane/go-splitflap/serdiev/generated"
	"github.com/denverquane/go-splitflap/serdiev/usb_serial"
	"log/slog"
)

type OutMessage struct {
	payload string
}

type Client struct {
	serial   *usb_serial.Splitflap
	lastSent string
}

func NewSplitflapClient() Client {
	return Client{
		serial:   nil,
		lastSent: "",
	}
}

func (c *Client) Connect(port string, notify func(state *gen.SplitflapState)) error {
	connection := usb_serial.NewSerialConnectionOnPort(port)
	if connection == nil {
		return errors.New("couldn't connect over USB")
	}
	sf := usb_serial.NewSplitflap(connection, notify)
	sf.Start()

	c.serial = sf
	return nil
}

// SetSerial allows setting the serial connection directly
// This is useful for mock connections
func (c *Client) SetSerial(sf *usb_serial.Splitflap) {
	c.serial = sf
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
				c.lastSent = msg.payload
				err := c.serial.SetTextWithMovement(msg.payload, usb_serial.ForceMovementNone)
				if err != nil {
					slog.Error(err.Error())
				}
			}
		}
	}
}
