// Code generated by sqlc. DO NOT EDIT.
// source: queries.sql

package repository

import (
	"context"
	"time"
)

const addCalc = `-- name: AddCalc :one
INSERT INTO irc_calcs (channel, "key", "by", "when", content)
VALUES ($1, $2, $3, $4, $5) RETURNING id
`

type AddCalcParams struct {
	Channel string    `json:"channel"`
	Key     string    `json:"key"`
	By      string    `json:"by"`
	When    time.Time `json:"when"`
	Content string    `json:"content"`
}

func (q *Queries) AddCalc(ctx context.Context, arg AddCalcParams) (int64, error) {
	row := q.db.QueryRow(ctx, addCalc,
		arg.Channel,
		arg.Key,
		arg.By,
		arg.When,
		arg.Content,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getCalcs = `-- name: GetCalcs :many
SELECT id, channel, key, by, "when", content
FROM irc_calcs
WHERE channel = $1
  AND "key" = $2
ORDER BY "when" ASC
`

type GetCalcsParams struct {
	Channel string `json:"channel"`
	Key     string `json:"key"`
}

func (q *Queries) GetCalcs(ctx context.Context, arg GetCalcsParams) ([]IrcCalc, error) {
	rows, err := q.db.Query(ctx, getCalcs, arg.Channel, arg.Key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IrcCalc
	for rows.Next() {
		var i IrcCalc
		if err := rows.Scan(
			&i.ID,
			&i.Channel,
			&i.Key,
			&i.By,
			&i.When,
			&i.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
