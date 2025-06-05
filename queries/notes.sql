-- name: FindAllNotes :many
SELECT *
FROM notes
WHERE deleted_at IS NULL;

-- name: FindNote :one
SELECT *
FROM notes
WHERE id = ?
  AND deleted_at IS NULL;

-- name: FindNoteByTitle :one
SELECT *
FROM notes
WHERE title = ?
  AND deleted_at IS NULL;

-- name: FindNoteByIDs :many
SELECT *
FROM notes
WHERE id IN (?)
  AND deleted_at IS NULL;

-- name: CreateNote :exec
INSERT INTO notes (note_id,title, content, user_id, created_at, updated_at)
VALUES (?,?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: UpdateNote :exec
UPDATE notes
SET title      = ?,
    content    = ?,
    updated_at = now()
WHERE id = ?
  AND deleted_at IS NULL;

-- name: DeleteNote :exec
UPDATE notes
SET deleted_at = now()
WHERE id = ?
  AND deleted_at IS NULL;