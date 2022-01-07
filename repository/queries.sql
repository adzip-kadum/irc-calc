-- name: GetCalcs :many
SELECT *
FROM irc_calcs
WHERE channel = $1
  AND "key" = $2
ORDER BY "when" ASC;

-- name: AddCalc :one
INSERT INTO irc_calcs (channel, "key", "by", "when", content)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
