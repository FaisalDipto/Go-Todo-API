package models

type Todo struct {
	Id	int					`json:"id"`
	Title string		`json:"title"`
	Status bool			`json:"status"`
}