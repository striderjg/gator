-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeeds :many
SELECT f.name, f.url, u.name AS username FROM feeds f INNER JOIN users u ON u.id = f.user_id;

-- name: MarkFeedFetched :exec
UPDATE feeds SET updated_at=sqlc.arg(time), last_fetched_at=sqlc.arg(time) WHERE id=$1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds ORDER BY last_fetched_at NULLS FIRST LIMIT 1;