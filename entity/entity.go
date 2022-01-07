package entity

import "time"

type Calc struct {
	Id       int64     `json:"id"`
	Channel  string    `json:"channel"`
	Encoding string    `json:"encoding"`
	Who      string    `json:"who"`
	By       string    `json:"by"`
	When     time.Time `json:"when"`
	Calc     string    `json:"calc"`
}
