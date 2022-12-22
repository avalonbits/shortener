-- name: SetShort :exec
INSERT INTO ShortLong(short, longn) VALUES (?, ?);

-- name: GetLong :one
SELECT longn FROM ShortLong WHERE short = ?;
