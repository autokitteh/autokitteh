package accountsstoregrpcsvc

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbaccount "github.com/autokitteh/autokitteh/api/gen/stubs/go/account"
	pbaccountsvc "github.com/autokitteh/autokitteh/api/gen/stubs/go/accountsvc"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/sdk/api/apiaccount"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Svc struct {
	pbaccountsvc.UnimplementedAccountsServer

	Store accountsstore.Store

	L L.Nullable
}

var _ pbaccountsvc.AccountsServer = &Svc{}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux) {
	pbaccountsvc.RegisterAccountsServer(srv, s)

	if gw != nil {
		if err := pbaccountsvc.RegisterAccountsHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) CreateAccount(ctx context.Context, req *pbaccountsvc.CreateAccountRequest) (*pbaccountsvc.CreateAccountResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apiaccount.AccountSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	if err := s.Store.Create(ctx, apiaccount.AccountName(req.Name), d); err != nil {
		if errors.Is(err, accountsstore.ErrAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "create: %v", err)
		}

		return nil, status.Errorf(codes.Unknown, "create: %v", err)
	}

	return &pbaccountsvc.CreateAccountResponse{}, nil
}

func (s *Svc) UpdateAccount(ctx context.Context, req *pbaccountsvc.UpdateAccountRequest) (*pbaccountsvc.UpdateAccountResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	d, err := apiaccount.AccountSettingsFromProto(req.Settings)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "data: %v", err)
	}

	name := apiaccount.AccountName(req.Name)

	if err := s.Store.Update(ctx, name, d); err != nil {
		return nil, status.Errorf(codes.Unknown, "update: %v", err)
	}

	return &pbaccountsvc.UpdateAccountResponse{}, nil
}

func (s *Svc) GetAccount(ctx context.Context, req *pbaccountsvc.GetAccountRequest) (*pbaccountsvc.GetAccountResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	name := apiaccount.AccountName(req.Name)

	a, err := s.Store.Get(ctx, name)
	if err != nil {
		if errors.Is(err, accountsstore.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "not found")
		}

		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	return &pbaccountsvc.GetAccountResponse{Account: a.PB()}, nil
}

func (s *Svc) GetAccounts(ctx context.Context, req *pbaccountsvc.GetAccountsRequest) (*pbaccountsvc.GetAccountsResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	names := make([]apiaccount.AccountName, len(req.Names))
	for i, name := range req.Names {
		names[i] = apiaccount.AccountName(name)
	}

	as, err := s.Store.BatchGet(ctx, names)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "get: %v", err)
	}

	pbas := make([]*pbaccount.Account, 0, len(as))
	for _, v := range as {
		if v != nil {
			pbas = append(pbas, v.PB())
		}
	}

	return &pbaccountsvc.GetAccountsResponse{Accounts: pbas}, nil
}
