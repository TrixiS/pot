package models

type Connection struct {
	ID   int    `storm:"id,increment" json:"id"`
	Name string `storm:"index"        json:"name"`
	Host string `storm:"index"        json:"host"`
	User string `                     json:"user"`
	Port uint16 `                     json:"port"`
}
