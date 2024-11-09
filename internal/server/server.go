package server

import (
	"context"
	"errors"
	"ssosage/internal/services/ssosage"
	"ssosage/internal/storage"
	"strings"

	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

/*
// registers new app - stores app name, secret and roles, if it already exists returns an error
	RegisterApp(context.Context, *RegisterAppRequest) (*RegisterAppResponse, error)
	// registers new client - stores client name and pass hash, if it already exists returns an error
	RegisterClient(context.Context, *RegisterClientRequest) (*RegisterClientResponse, error)
	// generates token for a specific app - token contains client name
	GenerateToken(context.Context, *GenerateTokenRequest) (*GenerateTokenResponse, error)
*/

type server struct {
	ssosage_proto.UnimplementedSsosageServer
	ssosage *ssosage.Ssosage
}

func (s *server) RegisterApp(ctx context.Context, request *ssosage_proto.RegisterAppRequest) (*ssosage_proto.RegisterAppResponse, error) {
	if !nameIsValid(request.GetAppName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid app name")
	}

	if !rolesAreValid(request.GetRoles()) {
		return nil, status.Error(codes.InvalidArgument, "invalid app roles")
	}

	if !secretIsValid(request.GetAppSecret()) {
		return nil, status.Error(codes.InvalidArgument, "invalid app secret")
	}

	_, err := s.ssosage.RegisterNewApp(ctx, request.GetAppName(), request.GetAppSecret(), strings.Join(request.GetRoles(), ","))

	if err != nil {
		if errors.Is(err, storage.ErrAppExists) {
			return nil, status.Error(codes.AlreadyExists, "app already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register app")
	}

	return &ssosage_proto.RegisterAppResponse{}, nil
}

func (s *server) RegisterClient(ctx context.Context, request *ssosage_proto.RegisterClientRequest) (*ssosage_proto.RegisterClientResponse, error) {
	if !nameIsValid(request.GetClientName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid client name")
	}

	if !passwordIsValid(request.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	_, err := s.ssosage.RegisterNewClient(ctx, request.GetClientName(), request.GetPassword())

	if err != nil {
		if errors.Is(err, storage.ErrClientExists) {
			return nil, status.Error(codes.AlreadyExists, "client already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register client")
	}

	return &ssosage_proto.RegisterClientResponse{}, nil
}

func (s *server) GenerateToken(ctx context.Context, request *ssosage_proto.GenerateTokenRequest) (*ssosage_proto.GenerateTokenResponse, error) {
	if !nameIsValid(request.GetAppName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid app name")
	}

	if !nameIsValid(request.GetClientName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid client name")
	}

	if !passwordIsValid(request.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	if !roleIsValid(request.GetRole()) {
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	token, err := s.ssosage.GenerateToken(ctx, request.GetClientName(), request.GetPassword(), request.GetAppName(), request.GetRole())

	if err != nil {
		if errors.Is(err, ssosage.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		if errors.Is(err, ssosage.ErrInvalidRole) {
			return nil, status.Error(codes.InvalidArgument, "invalid role")
		}

		if errors.Is(err, ssosage.ErrInvalidApp) {
			return nil, status.Error(codes.InvalidArgument, "invalid app")
		}

		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &ssosage_proto.GenerateTokenResponse{Token: token}, nil
}

func New(s *ssosage.Ssosage) *server {
	return &server{ssosage: s}
}
func nameIsValid(name string) bool {
	return len(name) > 0
}

func passwordIsValid(password string) bool {
	return len(password) > 0
}

func rolesAreValid(roles []string) bool {
	return len(roles) > 0
}

func secretIsValid(secret string) bool {
	return len(secret) > 0
}

func roleIsValid(role string) bool {
	return len(role) > 0
}
