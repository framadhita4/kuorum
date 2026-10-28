-- name: GetExploration :one
SELECT prompt, normalized_term, angles_json, created_at
FROM explorations
WHERE prompt = ?;

-- name: UpsertExploration :exec
INSERT INTO explorations (prompt, normalized_term, angles_json, created_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(prompt) DO UPDATE SET
    normalized_term = excluded.normalized_term,
    angles_json = excluded.angles_json,
    created_at = excluded.created_at;
