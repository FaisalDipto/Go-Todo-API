package models

type Todo struct {
	Id	int					`json:"id"`
	Title string		`json:"title" validate:"required,min=3,max=100"`
	Status bool			`json:"status"`
}