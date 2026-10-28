-- name: GetPaper :one
SELECT id, title, abstract, authors_json, year, citation_count, influential_citation_count, tldr, fields_of_study_json, open_access_pdf_url, doi, venue, fetched_at
FROM papers
WHERE id = ?;

-- name: UpsertPaper :exec
INSERT INTO papers (
    id, title, abstract, authors_json, year, citation_count,
    influential_citation_count, tldr, fields_of_study_json,
    open_access_pdf_url, doi, venue, fetched_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
    title = excluded.title,
    abstract = excluded.abstract,
    authors_json = excluded.authors_json,
    year = excluded.year,
    citation_count = excluded.citation_count,
    influential_citation_count = excluded.influential_citation_count,
    tldr = excluded.tldr,
    fields_of_study_json = excluded.fields_of_study_json,
    open_access_pdf_url = excluded.open_access_pdf_url,
    doi = excluded.doi,
    venue = excluded.venue,
    fetched_at = excluded.fetched_at;

-- name: SearchPapers :many
SELECT id, title, abstract, authors_json, year, citation_count, influential_citation_count, tldr, fields_of_study_json, open_access_pdf_url, doi, venue, fetched_at
FROM papers
WHERE title LIKE '%' || ? || '%' OR abstract LIKE '%' || ? || '%'
ORDER BY citation_count DESC
LIMIT ?;
