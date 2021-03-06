package users

import "context"

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Avatar   string `json:"-"`
}

type UserRepository interface {
	AddUser(ctx context.Context, user User) (int, error)
	UpdateUser(ctx context.Context, user User) error
	DeleteUser(ctx context.Context, userID int) error
	GetUserByID(ctx context.Context, userID int) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}

type UserUsecase interface {
	Registration(ctx context.Context, user User) (int, error)
	UpdateUser(ctx context.Context, changerID int, changedUser User) error
	DeleteUser(ctx context.Context, changerID int, userID int) error
	GetUserByID(ctx context.Context, userID int) (User, error)
	Authentication(ctx context.Context, user User) (int, error)
	UpdateAvatar(ctx context.Context, userID int, path string) error
}
