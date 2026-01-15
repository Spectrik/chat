package room

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var log = slog.Default().With("component", "room")
var ErrRoomStopped = errors.New("room stopped")

type envelope struct {
	ctx context.Context
	cmd Command
	rsp chan error
}

type Room struct {
	state   RoomState
	cmds    chan envelope
	queries chan queryReq
	events  chan Event

    stop   chan struct{}
	done    chan struct{}

	stopOnce sync.Once
	policies PolicySet
	Permanent bool
}

func NewRoom(name string, pol PolicySet) *Room {
	r := &Room{
		state:     newState(name),
		cmds:      make(chan envelope, 64),
		queries:   make(chan queryReq, 64),
		events:    make(chan Event, 256),
		done:      make(chan struct{}),
		stop:      make(chan struct{}),
		policies:  pol,
		Permanent: true, // TODO: Revert na false
	}

	go r.loop()
	return r
}

func (r *Room) Name() string { return r.state.Name }

func (r *Room) Events() <-chan Event { return r.events }

func (r *Room) Stop() {
    r.stopOnce.Do(func() {
        close(r.stop)
    })
	<-r.done
}

func (r *Room) ClientCount(ctx context.Context) (int, error) {
	var n int
	err := r.Query(ctx, func(rc *RoomCtx) error {
		n = len(rc.State.Clients)
		return nil
	})

	return n, err
}

func (r *Room) Query(ctx context.Context, apply func(*RoomCtx) error) error {
	rsp := make(chan error, 1)

	select {
		case <-r.done:
        	return ErrRoomStopped
		case <-ctx.Done():
			return ctx.Err()
		case r.queries <- queryReq{ctx: ctx, apply: apply, rsp: rsp}:
	}

	select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-rsp:
			return err
	}
}

func (r *Room) Exec(ctx context.Context, cmd Command) error {
	rsp := make(chan error, 1)

	select {
		case <-r.done:
			return ErrRoomStopped
		case <-ctx.Done():
			return ctx.Err()
		case r.cmds <- envelope{ctx: ctx, cmd: cmd, rsp: rsp}:
	}

	select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-rsp:
			return err
	}
}

func (r *Room) loop() {
	defer close(r.done)
	defer close(r.events)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
			case <-r.stop:
				return
			case env := <-r.cmds:
				ctx := &RoomCtx{State: &r.state, Policies: r.policies, Now: time.Now}
				evs, err := env.cmd.Apply(ctx)

				for _, ev := range evs {
					// tady si vyber: buď blokovat, nebo dropovat když je pomalý consumer
					select {
					case r.events <- ev:
					default:
						// drop + (můžeš přidat metriky)
					}
				}

				env.rsp <- err
				close(env.rsp)

			case q := <-r.queries:
				ctx := &RoomCtx{State: &r.state, Policies: r.policies, Now: time.Now}
				err := q.apply(ctx)
				q.rsp <- err
				close(q.rsp)

			case <-ticker.C:
				for c, mute := range r.state.Mutes {
					if mute.isExpired() {
						delete(r.state.Mutes, c)
					}
				}
		}
	}
}
