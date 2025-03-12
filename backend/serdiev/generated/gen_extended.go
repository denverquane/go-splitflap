package gen

import (
	"fmt"
	"log/slog"
)

func (m *FromSplitflap) PrintSplitflapState() {
	switch m.GetPayload().(type) {
	case *FromSplitflap_SplitflapState:
		modules := m.GetSplitflapState().GetModules()
		for i, module := range modules {
			if module.GetState().String() != "NORMAL" {
				slog.Error(fmt.Sprintf("Index %d: %s \n", i+1, module.GetState()))
			}
		}
	case *FromSplitflap_Log:
		//slog.Info(m.GetLog().Msg)
	case *FromSplitflap_Ack:
		//slog.Info("ack")
	case *FromSplitflap_SupervisorState:
		//slog.Info(m.GetSupervisorState().String())
	case *FromSplitflap_GeneralState:
		//slog.Info(m.GetGeneralState().String())
	default:
		slog.Info("Unknown message type received")
	}
}
