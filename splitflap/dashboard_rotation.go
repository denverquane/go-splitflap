package splitflap

import "time"

type DashboardRotation struct {
	Rotation []DashboardRotationEntry `json:"rotation"`
	idx      int
	kill     chan struct{}
}

type DashboardRotationEntry struct {
	Name         string `json:"name"`
	DurationSecs int    `json:"duration_secs"`
}

func (dr *DashboardRotation) Start(notifier chan<- string) {
	dr.kill = make(chan struct{})
	dr.idx = 0
	notifier <- dr.Rotation[0].Name
	timer := time.NewTimer(time.Duration(dr.Rotation[0].DurationSecs) * time.Second)
	go func() {
		for {
			select {
			case <-dr.kill:
				timer.Stop()
				return
			case <-timer.C:
				dr.idx = (dr.idx + 1) % len(dr.Rotation)
				notifier <- dr.Rotation[dr.idx].Name
				timer.Reset(time.Duration(dr.Rotation[dr.idx].DurationSecs) * time.Second)
			}
		}
	}()
}

func (dr *DashboardRotation) Stop() {
	dr.kill <- struct{}{}
}
