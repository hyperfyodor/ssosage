package models

type Client struct {
	ID           uint64
	Name         string
	PasswordHash []byte
}

type App struct {
	ID     uint64
	Name   string
	Secret string
	Roles  string
}
