package notes

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"notes/repositories"
	"notes/services/entities"
	"notes/services/tracing"
)

type Service struct {
	db         *sql.DB
	repository *repositories.Queries
}

func New(db *sql.DB) *Service {
	return &Service{
		db:         db,
		repository: repositories.New(db),
	}
}

func (s *Service) GetNotes(ctx context.Context) ([]entities.Note, error) {
	ctx, span := tracing.Tracer().Start(ctx, "svc.GetNotes")
	defer span.End()

	notes, err := s.repository.FindAllNotes(ctx)
	if err != nil {
		return nil, err
	}
	if len(notes) == 0 {
		return nil, sql.ErrNoRows
	}

	result := make([]entities.Note, 0, len(notes))
	for i := range notes {
		result = append(result, entities.Note{
			ID:        notes[i].NoteID,
			UserID:    notes[i].UserID,
			Title:     notes[i].Title,
			Content:   notes[i].Content,
			CreatedAt: notes[i].CreatedAt.Time,
		})
	}
	return result, nil
}

func (s *Service) CreateNote(ctx context.Context, noteReq entities.NoteReq) error {
	ctx, span := tracing.Tracer().Start(ctx, "svc.CreateNote")
	defer span.End()

	if noteReq.UserID == "" || noteReq.Title == "" || noteReq.Content == "" {
		return sql.ErrNoRows // or a custom error
	}

	err := s.repository.CreateNote(ctx, repositories.CreateNoteParams{
		NoteID:  uuid.NewString(),
		Title:   noteReq.Title,
		Content: noteReq.Content,
		UserID:  noteReq.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}
