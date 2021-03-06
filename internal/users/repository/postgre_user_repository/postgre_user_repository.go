package postgre_user_repository

import (
	"2021_1_Execute/internal/domain"
	"2021_1_Execute/internal/files"
	"2021_1_Execute/internal/users"
	"context"
	"regexp"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type PostgreUserRepository struct {
	Pool     *pgxpool.Pool
	FileUtil files.FileUtil
}

func NewPostgreUserRepository(pool *pgxpool.Pool, fileUtil files.FileUtil) users.UserRepository {
	return &PostgreUserRepository{
		Pool:     pool,
		FileUtil: fileUtil,
	}
}

func (repo *PostgreUserRepository) AddUser(ctx context.Context, user users.User) (int, error) {
	err := repo.insertUser(ctx, user)
	if err != nil {
		return -1, errors.Wrap(err, "Error while inserting user")
	}

	user.ID, err = repo.getIDByEmail(ctx, user.Email)
	if err != nil {
		return -1, errors.Wrap(err, "Error while getting id")
	}

	return user.ID, nil
}

func (repo *PostgreUserRepository) UpdateUser(ctx context.Context, user users.User) error {
	outdatedUser, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		return errors.Wrap(err, "Unable to get outdated user")
	}
	if outdatedUser.Email == "" {
		return domain.DBNotFoundError
	}

	newUser, err := repo.createUserUpdateObject(outdatedUser, user)
	if err != nil {
		return errors.Wrap(err, "Unable to create user update object")
	}

	err = repo.updateUserQuery(ctx, newUser)
	if err != nil {
		return errors.Wrap(err, "Error while updating user")
	}

	return nil
}

func (repo *PostgreUserRepository) DeleteUser(ctx context.Context, userID int) error {
	user, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Unable to get user by id")
	}
	if user.Email == "" {
		return domain.DBNotFoundError
	}

	rows, err := repo.Pool.Query(ctx, "delete from users where id = $1::int", user.ID)
	if err != nil {
		return errors.Wrap(err, "Unable to delete user")
	}
	rows.Close()

	return nil
}

func (repo *PostgreUserRepository) GetUserByID(ctx context.Context, userID int) (users.User, error) {
	rows, err := repo.Pool.Query(ctx, "select id, email, username, hashed_password, path_to_avatar from users where id = $1::int", userID)

	if err != nil {
		return users.User{}, errors.Wrap(err, "Error while getting user by id")
	}
	defer rows.Close()

	var result users.User

	for rows.Next() {
		err := rows.Scan(&result.ID, &result.Email, &result.Username, &result.Password, &result.Avatar)
		if err != nil {
			return users.User{}, errors.Wrap(err, "Error while reading user")
		}
	}
	return result, nil
}

func (repo *PostgreUserRepository) GetUserByEmail(ctx context.Context, email string) (users.User, error) {
	rows, err := repo.Pool.Query(ctx, "select id, email, username, hashed_password, path_to_avatar from users where email = $1::text", email)

	if err != nil {
		return users.User{}, errors.Wrap(err, "Error while getting user by email")
	}
	defer rows.Close()

	var result users.User

	for rows.Next() {
		err := rows.Scan(&result.ID, &result.Email, &result.Username, &result.Password, &result.Avatar)
		if err != nil {
			return users.User{}, errors.Wrap(err, "Error while reading user")
		}
	}
	return result, nil
}

func (repo *PostgreUserRepository) insertUser(ctx context.Context, user users.User) error {
	rows, err := repo.Pool.Query(ctx, "insert into users (email, username, hashed_password, path_to_avatar) values ($1::text, $2::text, $3::text, $4::text)",
		user.Email,
		user.Username,
		user.Password,
		user.Avatar,
	)
	if err != nil {
		return errors.Wrap(err, "Error while query insertUser")
	}
	rows.Close()
	err = rows.Err()
	if err != nil {
		re, reErr := regexp.MatchString(`duplicate key value violates unique constraint`, err.Error())
		switch {
		case reErr != nil:
			return errors.Wrap(reErr, "Invalid regexp")
		case re:
			return domain.DBConflictError
		default:
			return errors.Wrap(err, "postgreSQL error")
		}
	}

	return nil
}

func (repo *PostgreUserRepository) getIDByEmail(ctx context.Context, email string) (int, error) {
	rows, err := repo.Pool.Query(ctx, "select id from users where email = $1::text", email)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var id int = -1

	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return -1, errors.Wrap(err, "Unable to get user_id by email")
		}
	}

	if id == -1 {
		return -1, domain.DBNotFoundError
	}

	return id, nil
}

func (repo *PostgreUserRepository) updateUserQuery(ctx context.Context, user users.User) error {
	rows, err := repo.Pool.Query(ctx, "update users set email = $1::text, username = $2::text, hashed_password = $3::text, path_to_avatar = $4::text where id = $5::int",
		user.Email,
		user.Username,
		user.Password,
		user.Avatar,
		user.ID,
	)

	if err != nil {
		return errors.Wrap(err, "Error while query updateUser")
	}
	rows.Close()
	if err != nil {
		re, reErr := regexp.MatchString(`duplicate key value violates unique constraint`, err.Error())
		switch {
		case reErr != nil:
			return errors.Wrap(reErr, "Invalid regexp")
		case re:
			return domain.DBConflictError
		default:
			return errors.Wrap(err, "postgreSQL error")
		}
	}

	return nil
}

func (repo *PostgreUserRepository) createUserUpdateObject(outdatedUser, newUser users.User) (users.User, error) {
	var result users.User

	result.ID = outdatedUser.ID
	if newUser.Username != "" {
		result.Username = newUser.Username
	} else {
		result.Username = outdatedUser.Username
	}

	if newUser.Email != "" {
		result.Email = newUser.Email
	} else {
		result.Email = outdatedUser.Email
	}

	if newUser.Password != "" {
		result.Password = newUser.Password
	} else {
		result.Password = outdatedUser.Password
	}

	if newUser.Avatar != "" {
		if outdatedUser.Avatar != "" {
			err := repo.FileUtil.DeleteFile(outdatedUser.Avatar)
			if err != nil {
				return users.User{}, errors.Wrap(err, "Unable to delete outdated avatar")
			}
		}
		result.Avatar = newUser.Avatar
	} else {
		result.Avatar = outdatedUser.Avatar
	}

	return result, nil
}
