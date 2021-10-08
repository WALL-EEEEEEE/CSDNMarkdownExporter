package inter

import "time"

type Blog struct {
	Desc       string
	Url        string
	View       int
	Id         string
	Comment    int
	Title      string
	CreateTime time.Time
}
