package server

import "time"

// Note a basic notes struct
type Note struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// NoteReq request for creating notes
type NoteReq struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title" binding:"required"`
	Content string `json:"note" binding:"required"`
}
