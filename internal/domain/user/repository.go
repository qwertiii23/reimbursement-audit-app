package user

import "context"

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateLastLogin(ctx context.Context, userID string) error
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}
