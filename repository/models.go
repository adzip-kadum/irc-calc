// Code generated by sqlc. DO NOT EDIT.

package repository

import (
	"time"
)

type IrcCalc struct {
	ID      int64     `json:"id"`
	Channel string    `json:"channel"`
	Key     string    `json:"key"`
	By      string    `json:"by"`
	When    time.Time `json:"when"`
	Content string    `json:"content"`
}

type Migration struct {
	Version int32 `json:"version"`
}
