package utils

import (
	"log"
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
	log.Printf("[PERFORMANCE] %s : %d ms since start", name, elapsed.Milliseconds())
}
