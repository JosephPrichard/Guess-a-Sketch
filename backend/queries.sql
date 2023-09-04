-- name: Get :one
SELECT * FROM players 
WHERE id = $1 LIMIT 1;

-- name: PointsLeaderboard :many
SELECT * FROM players 
ORDER BY points DESC LIMIT $1;

-- name: WinsLeaderboard :many
SELECT * FROM players 
ORDER BY wins DESC LIMIT $1;

-- name: WordsLeaderboard :many
SELECT * FROM players 
ORDER BY words_guessed DESC LIMIT $1;

-- name: DrawingsLeaderboard :many
SELECT * FROM players 
ORDER BY drawings_guessed DESC LIMIT $1;

