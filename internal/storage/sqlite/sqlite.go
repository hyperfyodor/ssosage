package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"ssosage/internal/helpers"
	"ssosage/internal/models"
	"ssosage/internal/storage"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

/*
implements UserSaver, UserProvider
*/
type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {

	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite", storagePath)

	if err != nil {
		return nil, helpers.WrapErr(op, err)
	}

	return &Storage{db}, nil

}

func (s *Storage) SaveUser(ctx context.Context, name string, passwordHash []byte) error {

	const op = "storage.sqlite.SaveUser"

	query, err := s.db.Prepare("INSERT INTO users(name,pass_hash) VALUES(?, ?)")

	if err != nil {
		return helpers.WrapErr(op, err)
	}

	_, err = query.ExecContext(ctx, name, passwordHash)

	if err != nil {
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return storage.ErrUserExists
			}
		}

		return helpers.WrapErr(op, err)
	}

	return nil

}

func (s *Storage) User(ctx context.Context, name string) (models.User, error) {
	const op = "storage.sqlite.User"

	query, err := s.db.Prepare("SELECT name, pass_hash FROM users WHERE name = ?")

	if err != nil {
		return models.User{}, helpers.WrapErr(op, err)
	}

	row := query.QueryRowContext(ctx, name)

	var user models.User

	err = row.Scan(&user.Name, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, helpers.WrapErr(op, storage.ErrUserNotFound)
		}

		return models.User{}, helpers.WrapErr(op, err)
	}

	return user, nil
}

func (s *Storage) Stop() {
	s.db.Close()
}
