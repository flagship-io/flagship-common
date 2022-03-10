package decision

import (
	"time"
)

type Tracker struct {
	StartTime time.Time
	Enabled   bool
}

// TimeTrack logs execution since start
func (tracker *Tracker) TimeTrack(name string) {
	if tracker == nil || !tracker.Enabled {
		return
	}

	elapsed := time.Since(tracker.StartTime)
	logger.Logf(DebugLevel, "[PERFORMANCE] %s: %d ms since start", name, elapsed.Milliseconds())
}
