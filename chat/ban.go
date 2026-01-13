package chat

import "time"

type BanList map[string]BanEntry

type BanEntry struct {
	timestamp time.Time
	reason    string
}
