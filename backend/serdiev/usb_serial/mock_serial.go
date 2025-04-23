package usb_serial

import (
	gen "github.com/denverquane/go-splitflap/serdiev/generated"
	"github.com/denverquane/go-splitflap/serdiev/utils"
	"log"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

type MockConnection struct {
	outBuffer []byte
	inBuffer  []byte
	lock      sync.Mutex
}

func NewMockConnection(modules int) *MockConnection {
	GlobalAlphabet = []rune{' ', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L',
		'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y',
		'Z', 'g', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'r',
		'.', '?', '-', '$', '\'', '#', ':', 'd', ',', '!', '@', '&', 'w'}
	return &MockConnection{
		outBuffer: make([]byte, modules),
		inBuffer:  make([]byte, modules),
		lock:      sync.Mutex{},
	}
}

func (m *MockConnection) GetAvailablePorts() ([]string, error) {
	return []string{"COM1"}, nil
}

func (m *MockConnection) Open(portName string) error {
	return nil
}

// [4,8,100,34,5,47,210,156,12,0]
// Mock Write will automatically ack the message after 5 milliseconds
func (m *MockConnection) Write(data []byte) error {
	m.lock.Lock()
	m.outBuffer = append(m.outBuffer, data...)
	m.lock.Unlock()

	time.AfterFunc(time.Millisecond*5, func() {
		m.fakeOKAckMessage(data)
	})

	return nil
}

func (m *MockConnection) Read() ([]byte, error) {
	m.lock.Lock()
	buffer := m.inBuffer
	if len(buffer) > 0 {
		slog.Info("has data")
	}
	m.inBuffer = []byte{}
	m.lock.Unlock()
	return buffer, nil
}

func (m *MockConnection) Close() error {
	return nil
}

func (m *MockConnection) fakeOKAckMessage(bytes []byte) {
	payload, validCRC32 := utils.ParseCRC32EncodedPayload(bytes[:len(bytes)-1])
	if !validCRC32 {
		return
	}

	message := &gen.ToSplitflap{}
	if err := proto.Unmarshal(payload, message); err != nil {
		slog.Error("Failed to unmarshal message", "error", err)
		return
	}

	// Only handle splitflap config messages
	if config := message.GetSplitflapConfig(); config != nil {
		// Create a mock state response based on the received configuration
		stateMsg := createStateFromConfig(config)
		stateBytes, err := proto.Marshal(stateMsg)
		if err != nil {
			panic("err marshal state")
		}

		// Delay the state message slightly to simulate real hardware
		time.AfterFunc(time.Millisecond*10, func() {

			m.lock.Lock()
			m.inBuffer = append(m.inBuffer, utils.CreatePayloadWithCRC32Checksum(stateBytes)...)
			m.lock.Unlock()

			log.Println(config.String())
			slog.Info("Added state message")
		})
	}
}

func createStateFromConfig(config *gen.SplitflapConfig) *gen.FromSplitflap {
	// Create a state response that mirrors the requested configuration
	moduleStates := make([]*gen.SplitflapState_ModuleState, 0, len(config.Modules))

	for _, moduleConfig := range config.Modules {
		// Create a module state that matches what was requested in the config
		moduleState := &gen.SplitflapState_ModuleState{
			FlapIndex: moduleConfig.TargetFlapIndex, // Act as if module reached target immediately
			State:     gen.SplitflapState_ModuleState_NORMAL,
		}
		moduleStates = append(moduleStates, moduleState)
	}

	stateMsg := &gen.FromSplitflap{
		Payload: &gen.FromSplitflap_SplitflapState{
			SplitflapState: &gen.SplitflapState{
				Modules: moduleStates,
			},
		},
	}

	return stateMsg
}
