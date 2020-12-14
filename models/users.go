package models

import "context"

type UserClient interface {
	InviteUserByEmail(ctx context.Context) error
}
