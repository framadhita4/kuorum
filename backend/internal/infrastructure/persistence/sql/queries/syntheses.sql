-- name: GetGroundedSynthesis :one
SELECT anchor_paper_id, concept, definitions_json, mechanisms_json, case_studies_json, generated_at
FROM grounded_syntheses
WHERE anchor_paper_id = ? AND concept = ?;

-- name: UpsertGroundedSynthesis :exec
INSERT INTO grounded_syntheses (
    anchor_paper_id, concept, definitions_json,
    mechanisms_json, case_studies_json, generated_at
) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(anchor_paper_id, concept) DO UPDATE SET
    definitions_json = excluded.definitions_json,
    mechanisms_json = excluded.mechanisms_json,
    case_studies_json = excluded.case_studies_json,
    generated_at = excluded.generated_at;
