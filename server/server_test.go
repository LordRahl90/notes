package server

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		os.Exit(code)
	}()
	code = m.Run()
}

func TestPing(t *testing.T) {
	svr := New()
	w, err := newTestRequest(svr.router, http.MethodGet, "/ping", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	res := struct {
		Version string `json:"version"`
		Message string `json:"message"`
		Time    string `json:"time"`
	}{}

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	assert.Equal(t, "v2", res.Version)
	assert.Equal(t, "PONG V2", res.Message)
	assert.Equal(t, time.Now().Format(time.DateOnly), strings.Split(res.Time, " ")[0])
}

func TestCreate(t *testing.T) {
	req := NoteReq{
		Title:   "title",
		Content: "content",
	}

	b, err := json.Marshal(req)
	require.NoError(t, err)
	svr := New()
	w, err := newTestRequest(svr.router, http.MethodPost, "/", b)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, w.Code)

	var (
		res, getRes Note
	)

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
	require.NotEmpty(t, res.ID)
	assert.Equal(t, req.Title, res.Title)
	assert.Equal(t, req.Content, res.Content)

	w, err = newTestRequest(svr.router, http.MethodGet, "/"+res.ID, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, w.Code)

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &getRes))
	require.Equal(t, res, getRes)
}

func TestCreate_BadRequest(t *testing.T) {
	svr := New()
	w, err := newTestRequest(svr.router, http.MethodPost, "/", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGet_FindNonExistentNote(t *testing.T) {
	svr := New()
	w, err := newTestRequest(svr.router, http.MethodGet, "/123", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGet_All(t *testing.T) {
	svr := New()
	database[uuid.NewString()] = Note{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		Title:     "title",
		Content:   "content",
	}
	w, err := newTestRequest(svr.router, http.MethodGet, "/", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func newTestRequest(router *gin.Engine, method, path string, payload []byte) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	var (
		req *http.Request
		err error
	)
	if len(payload) > 0 {
		req, err = http.NewRequest(method, path, bytes.NewBuffer(payload))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	if err != nil {
		return nil, err
	}
	router.ServeHTTP(w, req)
	return w, nil
}
