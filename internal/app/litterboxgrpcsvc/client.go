package litterboxgrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pbevent "go.autokitteh.dev/idl/go/event"
	pbsvc "go.autokitteh.dev/idl/go/litterboxsvc"
)

type LocalClient struct {
	Server pbsvc.LitterBoxServer
}

var _ pbsvc.LitterBoxClient = &LocalClient{}

type runClient struct {
	grpc.ClientStream
	ctx context.Context
	ch  <-chan *pbevent.TrackIngestEventUpdate
}

type runServer struct {
	grpc.ServerStream
	ctx context.Context
	ch  chan<- *pbevent.TrackIngestEventUpdate
}

func (r *runClient) Recv() (*pbevent.TrackIngestEventUpdate, error) { return <-r.ch, nil }
func (r *runServer) Send(upd *pbevent.TrackIngestEventUpdate) error { r.ch <- upd; return nil }
func (r *runClient) Context() context.Context                       { return r.ctx }
func (r *runServer) Context() context.Context                       { return r.ctx }

func (c *LocalClient) Setup(ctx context.Context, in *pbsvc.SetupRequest, _ ...grpc.CallOption) (*pbsvc.SetupResponse, error) {
	return c.Server.Setup(ctx, in)
}

func (c *LocalClient) Run(ctx context.Context, in *pbsvc.RunRequest, _ ...grpc.CallOption) (pbsvc.LitterBox_RunClient, error) {
	ch := make(chan *pbevent.TrackIngestEventUpdate, 16)
	tx, rx := &runServer{ctx: ctx, ch: ch}, &runClient{ctx: ctx, ch: ch}
	go func() {
		_ = c.Server.Run(in, tx)
		close(ch)
	}()
	return rx, nil
}

func (c *LocalClient) Scoop(ctx context.Context, in *pbsvc.ScoopRequest, _ ...grpc.CallOption) (*pbsvc.ScoopResponse, error) {
	return c.Server.Scoop(ctx, in)
}
