// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: notes.sql

package repositories

import (
	"context"
)

const createNote = `-- name: CreateNote :exec
INSERT INTO notes (title, content, user_id)
VALUES ($1, $2, $3)
`

func (q *Queries) CreateNote(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, createNote)
	return err
}

const deleteNote = `-- name: DeleteNote :exec
UPDATE notes
SET deleted_at = now()
WHERE id = $1
  AND deleted_at IS NULL
`

func (q *Queries) DeleteNote(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteNote)
	return err
}

const findAllNotes = `-- name: FindAllNotes :many
SELECT id, note_id, title, content, user_id, created_at, updated_at, deleted_at
FROM notes
WHERE deleted_at IS NULL
`

func (q *Queries) FindAllNotes(ctx context.Context) ([]Note, error) {
	rows, err := q.db.QueryContext(ctx, findAllNotes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Note
	for rows.Next() {
		var i Note
		if err := rows.Scan(
			&i.ID,
			&i.NoteID,
			&i.Title,
			&i.Content,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findNote = `-- name: FindNote :one
SELECT id, note_id, title, content, user_id, created_at, updated_at, deleted_at
FROM notes
WHERE id = $1
  AND deleted_at IS NULL
`

func (q *Queries) FindNote(ctx context.Context) (Note, error) {
	row := q.db.QueryRowContext(ctx, findNote)
	var i Note
	err := row.Scan(
		&i.ID,
		&i.NoteID,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const findNoteByIDs = `-- name: FindNoteByIDs :many
SELECT id, note_id, title, content, user_id, created_at, updated_at, deleted_at
FROM notes
WHERE id IN ($1)
  AND deleted_at IS NULL
`

func (q *Queries) FindNoteByIDs(ctx context.Context) ([]Note, error) {
	rows, err := q.db.QueryContext(ctx, findNoteByIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Note
	for rows.Next() {
		var i Note
		if err := rows.Scan(
			&i.ID,
			&i.NoteID,
			&i.Title,
			&i.Content,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findNoteByTitle = `-- name: FindNoteByTitle :one
SELECT id, note_id, title, content, user_id, created_at, updated_at, deleted_at
FROM notes
WHERE title = $1
  AND deleted_at IS NULL
`

func (q *Queries) FindNoteByTitle(ctx context.Context) (Note, error) {
	row := q.db.QueryRowContext(ctx, findNoteByTitle)
	var i Note
	err := row.Scan(
		&i.ID,
		&i.NoteID,
		&i.Title,
		&i.Content,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const updateNote = `-- name: UpdateNote :exec
UPDATE notes
SET title      = $2,
    content    = $3,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
`

func (q *Queries) UpdateNote(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, updateNote)
	return err
}
