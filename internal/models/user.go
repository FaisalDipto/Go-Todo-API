package models

type User struct {
	ID						int			`json:"id"`
	Username			string	`json:"username" validate:"required,alphanum,min=3,max=32"`
	Password 			string	`json:"password,omitempty" validate:"required,min=6"`
	PasswordHash	string	`json:"-"`
}