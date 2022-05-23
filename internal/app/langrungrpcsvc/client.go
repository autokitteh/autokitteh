package langrungrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pblangsvc "go.autokitteh.dev/idl/go/langsvc"
)

type LocalClient struct {
	Server pblangsvc.LangRunServer
}

var _ pblangsvc.LangRunClient = &LocalClient{}

type runClient struct {
	grpc.ClientStream
	ctx context.Context
	ch  <-chan *pblangsvc.RunUpdate
}

type runServer struct {
	grpc.ServerStream
	ctx context.Context
	ch  chan<- *pblangsvc.RunUpdate
}

func (r *runClient) Recv() (*pblangsvc.RunUpdate, error) { return <-r.ch, nil }
func (r *runServer) Send(upd *pblangsvc.RunUpdate) error { r.ch <- upd; return nil }
func (r *runClient) Context() context.Context            { return r.ctx }
func (r *runServer) Context() context.Context            { return r.ctx }

func (c *LocalClient) Run(ctx context.Context, in *pblangsvc.RunRequest, _ ...grpc.CallOption) (pblangsvc.LangRun_RunClient, error) {
	ch := make(chan *pblangsvc.RunUpdate, 16)
	tx, rx := &runServer{ctx: ctx, ch: ch}, &runClient{ctx: ctx, ch: ch}
	go func() {
		_ = c.Server.Run(in, tx)
		close(ch)
	}()
	return rx, nil
}

func (c *LocalClient) CallFunction(ctx context.Context, in *pblangsvc.CallFunctionRequest, _ ...grpc.CallOption) (pblangsvc.LangRun_CallFunctionClient, error) {
	ch := make(chan *pblangsvc.RunUpdate, 16)
	tx, rx := &runServer{ctx: ctx, ch: ch}, &runClient{ctx: ctx, ch: ch}
	go func() {
		_ = c.Server.CallFunction(in, tx)
		close(ch)
	}()
	return rx, nil
}

func (c *LocalClient) RunGet(ctx context.Context, in *pblangsvc.RunGetRequest, _ ...grpc.CallOption) (*pblangsvc.RunGetResponse, error) {
	return c.Server.RunGet(ctx, in)
}

func (c *LocalClient) RunCallReturn(ctx context.Context, in *pblangsvc.RunCallReturnRequest, _ ...grpc.CallOption) (*pblangsvc.RunCallReturnResponse, error) {
	return c.Server.RunCallReturn(ctx, in)
}

func (c *LocalClient) RunLoadReturn(ctx context.Context, in *pblangsvc.RunLoadReturnRequest, _ ...grpc.CallOption) (*pblangsvc.RunLoadReturnResponse, error) {
	return c.Server.RunLoadReturn(ctx, in)
}

func (c *LocalClient) RunCancel(ctx context.Context, in *pblangsvc.RunCancelRequest, _ ...grpc.CallOption) (*pblangsvc.RunCancelResponse, error) {
	return c.Server.RunCancel(ctx, in)
}

func (c *LocalClient) ListRuns(ctx context.Context, in *pblangsvc.ListRunsRequest, _ ...grpc.CallOption) (*pblangsvc.ListRunsResponse, error) {
	return c.Server.ListRuns(ctx, in)
}

func (c *LocalClient) RunDiscard(ctx context.Context, in *pblangsvc.RunDiscardRequest, _ ...grpc.CallOption) (*pblangsvc.RunDiscardResponse, error) {
	return c.Server.RunDiscard(ctx, in)
}
