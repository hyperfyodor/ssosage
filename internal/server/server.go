package server

import (
	"context"
	"errors"
	"ssosage/internal/services/ssosage"
	"ssosage/internal/storage"

	"github.com/hyperfyodor/ssosage_proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

/**
Register(context.Context, *RegisterRequest) (*emptypb.Empty, error)
Login(context.Context, *LoginRequest) (*LoginResponse, error)
*/

type server struct {
	ssosage_proto.UnimplementedSsosageServer
	ssosage *ssosage.Ssosage
}

func (s *server) Register(ctx context.Context, request *ssosage_proto.RegisterRequest) (*empty.Empty, error) {
	if !nameIsValid(request.GetName()) {
		return nil, status.Error(codes.InvalidArgument, "invalid name")
	}

	if !passwirdIsValid(request.GetPassword()) {
		return nil, status.Error(codes.InvalidArgument, "invalid password")
	}

	err := s.ssosage.RegisterNewUser(ctx, request.GetName(), request.GetPassword())

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
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
		if errors.Is(err, ssosage.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid credentials")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssosage_proto.LoginResponse{Token: token}, nil
}

func New(s *ssosage.Ssosage) *server {
	return &server{ssosage: s}
}
func nameIsValid(name string) bool {
	return len(name) > 0
}

func passwirdIsValid(password string) bool {
	return len(password) > 0
}
