package litterboxgrpc

import (
	"context"
	"fmt"
	"io"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apivalues"

	pb "go.autokitteh.dev/idl/go/litterboxsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/litterbox"
)

type LitterBox struct {
	Client pb.LitterBoxClient
}

var _ litterbox.LitterBox = &LitterBox{}

func (lb *LitterBox) Setup(ctx context.Context, id litterbox.LitterBoxID, sources map[string][]byte, main string) (litterbox.LitterBoxID, error) {
	resp, err := lb.Client.Setup(ctx, &pb.SetupRequest{Id: string(id), Sources: sources, MainSourceName: main})
	if err != nil {
		return "", err
	}

	if err := resp.Validate(); err != nil {
		return "", fmt.Errorf("validate response: %w", err)
	}

	return litterbox.LitterBoxID(resp.Id), nil
}

func (lb *LitterBox) Scoop(ctx context.Context, id litterbox.LitterBoxID) error {
	_, err := lb.Client.Scoop(ctx, &pb.ScoopRequest{Id: string(id)})
	if err != nil {
		return err
	}

	return nil
}

func (lb *LitterBox) RunEvent(
	ctx context.Context,
	id litterbox.LitterBoxID,
	event *litterbox.LitterBoxEvent,
	ch chan<- *apievent.TrackIngestEventUpdate,
) (err error) {
	client, err := lb.Client.Run(ctx, &pb.RunRequest{
		Id: string(id),
		Event: &pb.LitterBoxEvent{
			Type:       event.Type,
			SrcBinding: event.SrcBinding,
			OriginalId: event.OriginalID,
			Data:       apivalues.StringValueMapToProto(event.Data),
		},
	})
	if err != nil {
		return err
	}

	for {
		pbupd, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return fmt.Errorf("recv: %w", err)
		}

		upd, err := apievent.TrackIngestEventUpdateFromProto(pbupd)
		if err != nil {
			return fmt.Errorf("invalid event: %w", err)
		}

		ch <- upd
	}
}

func (lb *LitterBox) Run(
	ctx context.Context,
	id litterbox.LitterBoxID,
	ch chan<- *apievent.TrackIngestEventUpdate,
) (err error) {
	client, err := lb.Client.Run(ctx, &pb.RunRequest{
		Id: string(id),
	})
	if err != nil {
		return err
	}

	for {
		pbupd, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return fmt.Errorf("recv: %w", err)
		}

		upd, err := apievent.TrackIngestEventUpdateFromProto(pbupd)
		if err != nil {
			return fmt.Errorf("invalid event: %w", err)
		}

		ch <- upd
	}
}
