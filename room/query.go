package room

import "context"

type queryReq struct {
	ctx   context.Context
	apply func(*RoomCtx) error
	rsp   chan error
}
