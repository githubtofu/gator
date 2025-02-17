-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name=$1;

-- name: GetUserById :one
SELECT * FROM users WHERE id=$1;

-- name: GetUsers :many
SELECT * FROM users;

-- name: Reset :exec
DELETE FROM users WHERE true;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url=$1;

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

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, feeds.name AS feed_name, users.name AS user_name
FROM feed_follows 
INNER JOIN users ON feed_follows.user_id=users.id
INNER JOIN feeds ON feed_follows.feed_id=feeds.id
WHERE feed_follows.user_id=$1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)

SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN users ON inserted_feed_follow.user_id=users.id
INNER JOIN feeds ON inserted_feed_follow.feed_id=feeds.id;
