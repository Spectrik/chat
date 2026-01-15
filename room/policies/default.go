package policies

import "github.com/ondrej/chat/room"

func DefaultPolicies() room.PolicySet {
	return room.NewPolicies().
		WithJoin(CapacityPolicy{Max: 50}).
		Build()
}
