package usb_serial

import (
	"errors"
	gen "github.com/denverquane/go-splitflap/serdiev/generated"
	"github.com/denverquane/go-splitflap/serdiev/utils"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

const (
	ForceMovementNone ForceMovement = iota
	ForceMovementOnlyNonBlank
	ForceMovementAll
	RetryTime     = time.Millisecond * 500
	HoldCharacter = uint32('a')
)

var GlobalAlphabet []rune

type ForceMovement int

type EnqueuedMessage struct {
	nonce uint32
	bytes []byte // bytes with CRC32 + null ending
}

type Splitflap struct {
	serial          SerialConnection
	outQueue        chan EnqueuedMessage
	ackQueue        chan uint32
	nextNonce       uint32
	run             bool
	lock            sync.Mutex
	currentConfig   *gen.SplitflapConfig
	numModules      int
	handleReadState func(state *gen.SplitflapState)
}

func NewSplitflap(serialInstance SerialConnection, handleState func(state *gen.SplitflapState), modules int) *Splitflap {
	s := &Splitflap{
		serial:          serialInstance,
		outQueue:        make(chan EnqueuedMessage, 100),
		ackQueue:        make(chan uint32, 100),
		nextNonce:       uint32(rand.Intn(256)),
		run:             true,
		currentConfig:   nil,
		handleReadState: handleState,
	}
	if modules > 0 {
		s.initializeModuleList(modules)
	}

	return s
}

func (sf *Splitflap) initializeModuleList(moduleCount int) {
	sf.numModules = moduleCount
	sf.currentConfig = &gen.SplitflapConfig{
		Modules: []*gen.SplitflapConfig_ModuleConfig{},
	}
	for i := 0; i < moduleCount; i++ {
		newModule := gen.SplitflapConfig_ModuleConfig{
			TargetFlapIndex: 0,
			MovementNonce:   0,
			ResetNonce:      0,
		}

		sf.currentConfig.Modules = append(sf.currentConfig.Modules, &newModule)
	}
}

func (sf *Splitflap) readLoop() {
	slog.Info("Read loop started")
	buffer := []byte{}
	for {
		if !sf.run {
			return
		}

		newBytes, err := sf.serial.Read()
		if err != nil {
			slog.Error("Error reading from serial", "error", err)
			return
		}

		if len(newBytes) == 0 {
			continue
		}

		buffer = append(buffer, newBytes...)
		lastByte := buffer[len(buffer)-1]
		if lastByte != 0 {
			continue
		}

		sf.processFrame(buffer[:len(buffer)-1])
		buffer = []byte{}
	}
}

func (sf *Splitflap) processFrame(decoded []byte) {
	payload, validCrc := utils.ParseCRC32EncodedPayload(decoded)
	if !validCrc {
		return
	}

	message := &gen.FromSplitflap{}

	if err := proto.Unmarshal(payload, message); err != nil {
		slog.Error("Failed to unmarshal", "error", err, "payload", payload)
		return
	}

	message.PrintSplitflapState()

	switch message.GetPayload().(type) {
	case *gen.FromSplitflap_Ack:
		nonce := message.GetAck().GetNonce()
		sf.ackQueue <- nonce
	case *gen.FromSplitflap_GeneralState:
		if GlobalAlphabet == nil {
			chars := message.GetGeneralState().GetFlapCharacterSet()
			GlobalAlphabet = make([]rune, len(chars))
			for i, v := range chars {
				GlobalAlphabet[i] = rune(v)
			}
			slog.Info("Set Alphabet using state received from Splitflap", "characters", string(GlobalAlphabet))
		}

	case *gen.FromSplitflap_SplitflapState:
		numModulesReported := len(message.GetSplitflapState().GetModules())

		if sf.numModules == 0 {
			sf.initializeModuleList(numModulesReported)
		} else if sf.numModules != numModulesReported {
			slog.Info("Number of reported modules changed\n", "old", sf.numModules, "new", numModulesReported)
		}

		sf.handleReadState(message.GetSplitflapState())
	}
}

func (sf *Splitflap) waitingForOutgoingMessage() bool {
	return len(sf.outQueue) == 0
}

func (sf *Splitflap) waitingForIncomingMessage() bool {
	return len(sf.ackQueue) == 0
}

func (sf *Splitflap) writeLoop() {

	slog.Info("Write loop started")

	for {
		if !sf.run {
			slog.Info("Stop running, exiting write loop")
			return
		}

		if sf.waitingForOutgoingMessage() {
			continue
		}

		enqueuedMessage := <-sf.outQueue

		nextRetry := time.Now()
		writeCount := 0
		for {
			if !sf.run {
				slog.Info("Stop running, exiting write loop")
				return
			}

			if time.Now().After(nextRetry) {
				if writeCount > 0 {
					slog.Info("Failed to write message, resetting queue")
					sf.outQueue = make(chan EnqueuedMessage, 100)
					break
				}

				writeCount++
				sf.serial.Write(enqueuedMessage.bytes)
				nextRetry = time.Now().Add(RetryTime)
			}

			if sf.waitingForIncomingMessage() {
				continue
			}

			latestAckNonce := <-sf.ackQueue
			if enqueuedMessage.nonce == latestAckNonce {
				break
			}
		}
	}
}

func (sf *Splitflap) SetText(text string) error {
	return sf.SetTextWithMovement(text, ForceMovementNone)
}

func (sf *Splitflap) SetTextWithMovement(text string, forceMovement ForceMovement) error {
	// Transform text to a list of flap indexes (and pad with blanks so that all modules get updated even if text is shorter)
	var positions []uint32
	for _, c := range text {
		idx := uint32(AlphabetIndex(c))
		if uint32(c) == HoldCharacter {
			idx = HoldCharacter
		}
		positions = append(positions, idx)
	}

	// Pad with blanks if text is shorter than the number of modules
	for i := len(text); i < sf.numModules; i++ {
		positions = append(positions, uint32(AlphabetIndex(' ')))
	}

	var forceMovementList []bool
	switch forceMovement {
	case ForceMovementNone:
		forceMovementList = nil
	case ForceMovementOnlyNonBlank:
		for _, c := range text {
			forceMovementList = append(forceMovementList, AlphabetIndex(c) != 0 && uint32(c) != HoldCharacter)
		}
		// Pad with false if text is shorter than the number of modules
		for i := len(text); i < sf.numModules; i++ {
			forceMovementList = append(forceMovementList, false)
		}
	case ForceMovementAll:
		forceMovementList = make([]bool, sf.numModules)
		for i := range forceMovementList {
			forceMovementList[i] = true
		}
	default:
		panic("Bad movement value")
	}

	return sf.setPositions(positions, forceMovementList)
}

func (sf *Splitflap) SpinCharacter(idx int) {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	sf.currentConfig.Modules[idx].MovementNonce = (sf.currentConfig.Modules[idx].MovementNonce + 1) % 256
	message := &gen.ToSplitflap{
		Payload: &gen.ToSplitflap_SplitflapConfig{
			SplitflapConfig: sf.currentConfig,
		},
	}

	sf.enqueueMessage(message)
}

func (sf *Splitflap) setPositions(positions []uint32, forceMovementList []bool) error {
	sf.lock.Lock()
	defer sf.lock.Unlock()

	if sf.numModules == 0 {
		return errors.New("cannot set positions before the number of modules is known")
	}

	if len(positions) > sf.numModules {
		return errors.New("more positions specified than modules")
	}

	if forceMovementList != nil && len(positions) != len(forceMovementList) {
		return errors.New("positions and forceMovementList length must match")
	}

	for i, v := range positions {
		if v != HoldCharacter {
			sf.currentConfig.Modules[i].TargetFlapIndex = v
			if forceMovementList != nil && forceMovementList[i] {
				sf.currentConfig.Modules[i].MovementNonce = (sf.currentConfig.Modules[i].MovementNonce + 1) % 256
			}
		}
	}

	message := &gen.ToSplitflap{
		Payload: &gen.ToSplitflap_SplitflapConfig{
			SplitflapConfig: sf.currentConfig,
		},
	}

	sf.enqueueMessage(message)
	return nil
}

func (sf *Splitflap) enqueueMessage(message *gen.ToSplitflap) {
	message.Nonce = sf.nextNonce
	sf.nextNonce++

	payload, err := proto.Marshal(message)
	if err != nil {
		slog.Error("Error marshaling message", "error", err)
		return
	}

	newMessage := EnqueuedMessage{
		nonce: message.Nonce,
		bytes: utils.CreatePayloadWithCRC32Checksum(payload),
	}

	sf.outQueue <- newMessage

	approxQLength := len(sf.outQueue)
	// TODO: handle error in some way
	// logger.Info().Msgf("Out q length: %d\n", approxQLength)
	if approxQLength > 10 {
		slog.Info("Output queue length is high! (%d) Is the splitflap still connected and functional?", "length", approxQLength)
	}
}

func (sf *Splitflap) RequestState() {
	message := gen.ToSplitflap{}
	message.Payload = &gen.ToSplitflap_RequestState{
		RequestState: &gen.RequestState{},
	}

	sf.enqueueMessage(&message)
}

func AlphabetIndex(c rune) int {
	for i, char := range GlobalAlphabet {
		if char == c {
			return i
		}
	}

	return 0 // Default to 0 if character not found in alphabet
}

func AlphabetDistance(a, b rune) int {
	aIdx := AlphabetIndex(a)
	bIdx := AlphabetIndex(b)
	if bIdx < aIdx {
		return (len(GlobalAlphabet) - aIdx) + bIdx
	}
	return bIdx - aIdx
}

func (sf *Splitflap) Start() {
	go sf.readLoop()
	go sf.writeLoop()
	sf.RequestState()
}
