package models

type User struct {
	Name         string
	PasswordHash []byte
}
