package markets_test

import (
	"context"

	dusers "socialpredict/internal/domain/users"
)

type noOpCreatorProfile struct{}

func (noOpCreatorProfile) ValidateUserExists(context.Context, string) error { return nil }

func (noOpCreatorProfile) GetPublicUser(context.Context, string) (*dusers.PublicUser, error) {
	return nil, nil
}

type noOpWallet struct{}

func (noOpWallet) ValidateBalance(context.Context, string, int64, int64) error { return nil }
func (noOpWallet) Debit(context.Context, string, int64, int64, string) error   { return nil }
func (noOpWallet) Credit(context.Context, string, int64, string) error         { return nil }
