package models

type User struct {
	ID           string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}
