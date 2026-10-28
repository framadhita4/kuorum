-- name: GetCitationGraph :one
SELECT anchor_paper_id, nodes_json, edges_json, traversed_at
FROM citation_graphs
WHERE anchor_paper_id = ?;

-- name: UpsertCitationGraph :exec
INSERT INTO citation_graphs (anchor_paper_id, nodes_json, edges_json, traversed_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(anchor_paper_id) DO UPDATE SET
    nodes_json = excluded.nodes_json,
    edges_json = excluded.edges_json,
    traversed_at = excluded.traversed_at;

-- name: DeleteCitationGraph :exec
DELETE FROM citation_graphs
WHERE anchor_paper_id = ?;
