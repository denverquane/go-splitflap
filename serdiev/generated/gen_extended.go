package gen

import (
	"fmt"
	"log/slog"
)

const (
	SplitflapStateType SplitFlapType = iota
	LogType
	AckType
	SupervisorStateType
	GeneralStateType
	Unknown
)

type SplitFlapType int

func (m *FromSplitflap) GetPayloadType() SplitFlapType {
	switch m.GetPayload().(type) {
	case *FromSplitflap_SplitflapState:
		return SplitflapStateType
	case *FromSplitflap_Log:
		return LogType
	case *FromSplitflap_Ack:
		return AckType
	case *FromSplitflap_SupervisorState:
		return SupervisorStateType
	case *FromSplitflap_GeneralState:
		return GeneralStateType
	}

	return Unknown
}

func (m *FromSplitflap) PrintSplitflapState() {
	payloadType := m.GetPayloadType()
	switch payloadType {
	case SplitflapStateType:
		state := m.GetSplitflapState()
		msg := ""
		for i, module := range state.GetModules() {
			msg += module.GetState().String()
			if module.GetState().String() != "NORMAL" {
				msg += fmt.Sprintf("Index %d: %s \n", i+1, module.GetState())
			}
		}
		if msg != "" {
			slog.Info(msg)
		}
	case LogType:
		//slog.Info(m.GetLog().Msg)
	case AckType:
		slog.Info("ack")
	case SupervisorStateType:
		//slog.Info(m.GetSupervisorState().String())
	case GeneralStateType:
		//slog.Info(m.GetGeneralState().String())
	case Unknown:
		slog.Info("Unknown message type")
	}

}
