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

func (s *Storage) SaveClient(ctx context.Context, name string, passwordHash []byte) (int64, error) {

	const op = "storage.sqlite.SaveClient"

	query, err := s.db.Prepare("INSERT INTO clients(name,password_hash) VALUES(?, ?)")

	if err != nil {
		return 0, helpers.WrapErr(op, err)
	}

	res, err := query.ExecContext(ctx, name, passwordHash)

	if err != nil {
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return 0, helpers.WrapErr(op, storage.ErrClientExists)
			}
		}

		return 0, helpers.WrapErr(op, err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, helpers.WrapErr(op, err)
	}

	return id, nil

}

func (s *Storage) Client(ctx context.Context, name string) (models.Client, error) {
	const op = "storage.sqlite.Client"

	query, err := s.db.Prepare("SELECT id, name, password_hash FROM clients WHERE name = ?")

	if err != nil {
		return models.Client{}, helpers.WrapErr(op, err)
	}

	row := query.QueryRowContext(ctx, name)

	var client models.Client

	err = row.Scan(&client.ID, &client.Name, &client.PasswordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Client{}, helpers.WrapErr(op, storage.ErrClientNotFound)
		}

		return models.Client{}, helpers.WrapErr(op, err)
	}

	return client, nil
}

func (s *Storage) SaveApp(ctx context.Context, name string, secret string, roles string) (int64, error) {

	const op = "storage.sqlite.SaveApp"

	query, err := s.db.Prepare("INSERT INTO apps(name,secret,roles) VALUES(?, ?, ?)")

	if err != nil {
		return 0, helpers.WrapErr(op, err)
	}

	res, err := query.ExecContext(ctx, name, secret, roles)

	if err != nil {
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return 0, helpers.WrapErr(op, storage.ErrAppExists)
			}
		}

		return 0, helpers.WrapErr(op, err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, helpers.WrapErr(op, err)
	}

	return id, nil
}

func (s *Storage) App(ctx context.Context, name string) (models.App, error) {
	const op = "storage.sqlite.App"

	query, err := s.db.Prepare("SELECT id, name, secret, roles FROM apps WHERE name = ?")

	if err != nil {
		return models.App{}, helpers.WrapErr(op, err)
	}

	row := query.QueryRowContext(ctx, name)

	var app models.App

	err = row.Scan(&app.ID, &app.Name, &app.Secret, &app.Roles)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, helpers.WrapErr(op, storage.ErrAppNotFound)
		}

		return models.App{}, helpers.WrapErr(op, err)
	}

	return app, nil
}

func (s *Storage) Stop() {
	s.db.Close()
}
