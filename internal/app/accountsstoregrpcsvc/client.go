package accountsstoregrpcsvc

import (
	"context"

	"google.golang.org/grpc"

	pbaccountsvc "go.autokitteh.dev/idl/go/accountsvc"
)

type LocalClient struct {
	Server pbaccountsvc.AccountsServer
}

var _ pbaccountsvc.AccountsClient = &LocalClient{}

func (c *LocalClient) CreateAccount(ctx context.Context, in *pbaccountsvc.CreateAccountRequest, _ ...grpc.CallOption) (*pbaccountsvc.CreateAccountResponse, error) {
	return c.Server.CreateAccount(ctx, in)
}

func (c *LocalClient) UpdateAccount(ctx context.Context, in *pbaccountsvc.UpdateAccountRequest, _ ...grpc.CallOption) (*pbaccountsvc.UpdateAccountResponse, error) {
	return c.Server.UpdateAccount(ctx, in)
}

func (c *LocalClient) GetAccount(ctx context.Context, in *pbaccountsvc.GetAccountRequest, _ ...grpc.CallOption) (*pbaccountsvc.GetAccountResponse, error) {
	return c.Server.GetAccount(ctx, in)
}

func (c *LocalClient) GetAccounts(ctx context.Context, in *pbaccountsvc.GetAccountsRequest, _ ...grpc.CallOption) (*pbaccountsvc.GetAccountsResponse, error) {
	return c.Server.GetAccounts(ctx, in)
}
