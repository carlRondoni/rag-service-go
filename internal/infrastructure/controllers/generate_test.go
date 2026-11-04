//go:build unit || !integration

package controllers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"rag-service-go/internal/application"
	"rag-service-go/internal/domain"
	"rag-service-go/internal/infrastructure/controllers"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateController_Success(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	llmMock := new(mockLLMForController)
	storeMock := new(mockVectorStore)

	llmMock.On("Embed", mock.Anything, []string{"what is go?"}).
		Return([]domain.Embedding{{Vector: []float32{0.1, 0.2}}}, nil)
	storeMock.On("Search", mock.Anything, domain.Embedding{Vector: []float32{0.1, 0.2}}, 3).
		Return([]domain.SearchResult{}, nil)
	llmMock.On("Generate", mock.Anything, "what is go?").
		Return("Go is an open source language", nil)

	handler := application.NewGenerateHandler(llmMock, storeMock)
	controller := controllers.NewGenerateController(handler, logger)

	body := map[string]string{
		"prompt": "what is go?",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(res.Body).Decode(&respBody)
	assert.Equal(t, "Go is an open source language", respBody["response"])
}

func TestGenerateController_MethodNotAllowed(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	controller := controllers.NewGenerateController(application.GenerateHandler{}, logger)

	req := httptest.NewRequest(http.MethodGet, "/generate", nil)
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}

func TestGenerateController_MissingPrompt(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	controller := controllers.NewGenerateController(application.GenerateHandler{}, logger)

	body := map[string]string{}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGenerateController_HandlerError(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	llmMock := new(mockLLMForController)
	storeMock := new(mockVectorStore)

	llmMock.On("Embed", mock.Anything, []string{"what is go?"}).
		Return([]domain.Embedding{}, errors.New("embed failed"))

	handler := application.NewGenerateHandler(llmMock, storeMock)
	controller := controllers.NewGenerateController(handler, logger)

	body := map[string]string{
		"prompt": "what is go?",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(b))
	w := httptest.NewRecorder()

	controller.Execute(w, req)

	res := w.Result()
	defer func() { assert.NoError(t, res.Body.Close()) }()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

