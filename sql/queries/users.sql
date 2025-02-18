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

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT * FROM posts 
INNER JOIN feed_follows ON feed_follows.feed_id=feeds.id
INNER JOIN feeds ON posts.feed_id=feeds.id
WHERE feed_follows.user_id=$1
order by posts.published_at desc
limit $2;

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

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE feed_id=(
    SELECT id FROM feeds
    WHERE url=$2
) AND feed_follows.user_id=$1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id=$1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;
