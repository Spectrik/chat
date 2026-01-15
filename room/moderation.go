package room

import "time"

type Mute struct {
	When time.Time
	Duration time.Duration
}

func (m Mute) isExpired() bool {
	return time.Now().After(m.When.Add(m.Duration))
}
