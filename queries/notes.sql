-- name: FindAllNotes :many
SELECT *
FROM notes
WHERE deleted_at IS NULL;

-- name: FindNote :one
SELECT *
FROM notes
WHERE id = $1
  AND deleted_at IS NULL;

-- name: FindNoteByTitle :one
SELECT *
FROM notes
WHERE title = $1
  AND deleted_at IS NULL;

-- name: FindNoteByIDs :many
SELECT *
FROM notes
WHERE id IN ($1)
  AND deleted_at IS NULL;

-- name: CreateNote :exec
INSERT INTO notes (title, content, user_id)
VALUES ($1, $2, $3);

-- name: UpdateNote :exec
UPDATE notes
SET title      = $2,
    content    = $3,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: DeleteNote :exec
UPDATE notes
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL;