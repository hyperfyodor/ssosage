package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	ssosage_proto "github.com/hyperfyodor/ssosage_proto"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	empty "google.golang.org/protobuf/types/known/emptypb"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const APP_SECRET = "hello"

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("token is invalid")
)

type Ssosage struct {
	logger  *slog.Logger
	storage *SqliteStorage
}

// type UserSaver interface {
// 	SaveUser(ctx context.Context, name string, passHash []byte) error
// }

// type UserProvider interface {
// 	User(ctx context.Context, name string) (User, error)
// }

type User struct {
	name     string
	passHash []byte
}

type SqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage(storagePath string) (*SqliteStorage, error) {

	const op = "SqliteStorage.New"

	db, err := sql.Open("sqlite", storagePath)

	if err != nil {
		return nil, wrap(op, err)
	}

	return &SqliteStorage{db}, nil

}

func (s *SqliteStorage) SaveUser(ctx context.Context, name string, passHash []byte) error {

	const op = "SqliteStorage.SaveUser"

	query, err := s.db.Prepare("INSERT INTO users(name,pass_hash) VALUES(?, ?)")

	if err != nil {
		return wrap(op, err)
	}

	_, err = query.ExecContext(ctx, name, passHash)

	if err != nil {
		if liteErr, ok := err.(*sqlite.Error); ok {
			code := liteErr.Code()
			if code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
				return ErrUserExists
			}
		}

		return wrap(op, err)
	}

	return nil

}

func (s *SqliteStorage) User(ctx context.Context, name string) (User, error) {
	const op = "SqliteStorage.SaveUser"

	query, err := s.db.Prepare("SELECT name, pass_hash FROM users WHERE name = ?")

	if err != nil {
		return User{}, wrap(op, err)
	}

	row := query.QueryRowContext(ctx, name)

	var user User

	err = row.Scan(&user.name, &user.passHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, wrap(op, ErrUserNotFound)
		}

		return User{}, wrap(op, err)
	}

	return user, nil
}

func (s *SqliteStorage) Stop() {
	s.db.Close()
}

func NewSsosage(
	log *slog.Logger,
	storage *SqliteStorage,
) *Ssosage {
	return &Ssosage{log, storage}
}

func (s *Ssosage) RegisterNewUser(ctx context.Context, name string, password string) error {

	const op = "Ssosage.RegisterNewUser"

	log := s.logWith(op, name)
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // TODO create interface for hashing + salt

	if err != nil {
		log.Error("failed to generate hash", slErr(err))

		return wrap(op, err)
	}

	err = s.storage.SaveUser(ctx, name, passHash)

	if err != nil {

		if errors.Is(err, ErrUserExists) {
			log.Error("user exists", slErr(err))

			return ErrUserExists
		}

		log.Error("failed to save user", slErr(err))

		return wrap(op, err)
	}

	return nil

}

func (s *Ssosage) Login(ctx context.Context, name string, password string) (string, error) {

	const op = "Ssosage.Login"

	log := s.logWith(op, name)

	log.Info("logging in")

	user, err := s.storage.User(ctx, name)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn("user not found", slErr(err))

			return "", ErrInvalidCredentials
		}

		log.Error("failed to get user", slErr(err))

		return "", wrap(op, err)
	}

	// TODO compare hash and pass should be interface
	if err := bcrypt.CompareHashAndPassword(user.passHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slErr(err))

		return "", ErrInvalidCredentials
	}

	token, err := newToken(name)

	if err != nil {
		log.Info("failed to generate token", slErr(err))

		return "", wrap(op, err)
	}

	return token, nil

}

type server struct {
	ssosage_proto.UnimplementedSsosageServer
	ssosage *Ssosage
}

/*
Register(context.Context, *RegisterRequest) (*emptypb.Empty, error)
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
*/

func (s *server) Register(ctx context.Context, request *ssosage_proto.RegisterRequest) (*empty.Empty, error) {
	if !nameIsValid(request.GetName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}

	if !passwirdIsValid(request.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	err := s.ssosage.RegisterNewUser(ctx, request.GetName(), request.GetPassword())

	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &empty.Empty{}, nil
}

func (s *server) Login(ctx context.Context, request *ssosage_proto.LoginRequest) (*ssosage_proto.LoginResponse, error) {
	if !nameIsValid(request.GetName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}

	if !passwirdIsValid(request.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	token, err := s.ssosage.Login(ctx, request.GetName(), request.GetPassword())

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssosage_proto.LoginResponse{Token: token}, nil
}

func wrap(op string, err error) error {
	return fmt.Errorf("%s : %w", op, err)
}

func (s *Ssosage) logWith(op string, name string) *slog.Logger {
	return s.logger.With(
		slog.String("op", op),
		slog.String("name", name),
	)
}

func slErr(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func newToken(name string) (string, error) { // jwt generation and validation should be interface
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = name
	claims["exp"] = 0

	tokenString, err := token.SignedString([]byte(APP_SECRET)) // each service should register in ssosage and provide its secret (should it?)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return APP_SECRET, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}

func nameIsValid(name string) bool {
	return len(name) > 0
}

func passwirdIsValid(password string) bool {
	return len(password) > 0
}
