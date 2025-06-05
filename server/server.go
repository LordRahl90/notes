package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"notes/services/entities"
	"time"

	"notes/services/tracing"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	//"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
)

var database map[string]entities.Note

// Server is the server :)
type Server struct {
	router *gin.Engine
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
		//otelgin.Middleware("notes"),
		logMiddleware(),
	)
	database = make(map[string]entities.Note)

	s := &Server{
		router: router,
	}

	router.GET("/ping", func(c *gin.Context) {
		_, span := tracing.Tracer().Start(c.Request.Context(), "ping")
		defer span.End()
		sentry.CaptureException(errors.New("sentry error handling"))
		span.SetAttributes(attribute.String("client", c.ClientIP()))
		slog.InfoContext(c.Request.Context(), "pong", "time", time.Now().String(), "client", c.ClientIP())

		if err := waitService(c.Request.Context()); err != nil {
			slog.ErrorContext(c.Request.Context(), "wait service failed", "time", time.Now().String(), "client", c.ClientIP(), "error", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"version": "v2",
			"message": "PONG V2 Redeployed",
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

func waitService(ctx context.Context) error {
	ctx, span := tracing.Tracer().Start(ctx, "wait")
	defer span.End()
	time.Sleep(4 * time.Second)

	slog.InfoContext(ctx, "Wait service has completed")

	innerWaitService(ctx)

	return nil
}

func innerWaitService(ctx context.Context) {
	_, span := tracing.Tracer().Start(ctx, "inner")
	defer span.End()

	time.Sleep(500 * time.Microsecond)
}

func (s *Server) create(ctx *gin.Context) {
	var req entities.NoteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	note := entities.Note{
		ID:        uuid.NewString(),
		UserID:    req.UserID,
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
