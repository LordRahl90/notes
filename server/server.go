package server

import (
	"errors"

	"log/slog"
	"net/http"
	"notes/tracing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
)

var database map[string]Note

type Server struct {
	router *gin.Engine
}

// Note a basic notes struct
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// NoteReq request for creating notes
type NoteReq struct {
	Title   string `json:"title"`
	Content string `json:"note"`
}

func logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.InfoContext(c.Request.Context(), "request started", "time", time.Now().String(), "client", c.ClientIP(), "method", c.Request.Method)
		c.Next()
		slog.InfoContext(c.Request.Context(), "request completed", "time", time.Now().String(), "client", c.ClientIP(), "method", c.Request.Method)
	}
}

// New returns a new
func New() *Server {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		otelgin.Middleware("notes"),
		logMiddleware(),
	)
	database = make(map[string]Note)

	s := &Server{
		router: router,
	}

	router.GET("/ping", func(c *gin.Context) {
		_, span := tracing.Tracer().Start(c.Request.Context(), "ping")
		defer span.End()
		sentry.CaptureException(errors.New("sentry error handling"))
		span.SetAttributes(attribute.String("client", c.ClientIP()))
		slog.InfoContext(c.Request.Context(), "pong", "time", time.Now().String(), "client", c.ClientIP())
		c.JSON(http.StatusOK, gin.H{
			"version": "v2",
			"message": "PONG V2",
			"time":    time.Now().String(),
		})
		slog.InfoContext(c.Request.Context(), "all completed!", "time", time.Now().String(), "client", c.ClientIP(), "method", c.Request.Method)
	})
	router.POST("/", s.create)
	router.GET("/", s.all)
	router.GET("/:id", s.single)
	return s
}

// Start starts the server
func (s *Server) Start(port string) error {
	return s.router.Run(port)
}

func (s *Server) create(ctx *gin.Context) {
	var req NoteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	note := Note{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		Title:     req.Title,
		Content:   req.Content,
	}
	database[note.ID] = note
	ctx.JSON(http.StatusCreated, note)
}

func (s *Server) all(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, database)
}

func (s *Server) single(ctx *gin.Context) {
	id := ctx.Param("id")
	v, ok := database[id]
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "key not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, v)
}
