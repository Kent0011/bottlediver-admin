package handler

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
)

type mockContentService struct {
	listFn   func(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error)
	createFn func(ctx context.Context, contentType domain.ContentType, input domain.DocumentInput) (domain.Document, error)
	updateFn func(ctx context.Context, contentType domain.ContentType, id string, input domain.DocumentInput) (domain.Document, error)
	deleteFn func(ctx context.Context, contentType domain.ContentType, id string) error
}

func (m mockContentService) List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error) {
	return m.listFn(ctx, contentType)
}

func (m mockContentService) Create(ctx context.Context, contentType domain.ContentType, input domain.DocumentInput) (domain.Document, error) {
	return m.createFn(ctx, contentType, input)
}

func (m mockContentService) Update(ctx context.Context, contentType domain.ContentType, id string, input domain.DocumentInput) (domain.Document, error) {
	return m.updateFn(ctx, contentType, id, input)
}

func (m mockContentService) Delete(ctx context.Context, contentType domain.ContentType, id string) error {
	return m.deleteFn(ctx, contentType, id)
}

func TestHandleGetNews(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	service := mockContentService{
		listFn: func(_ context.Context, contentType domain.ContentType) ([]domain.Document, error) {
			if contentType != domain.ContentTypeNews {
				t.Fatalf("unexpected content type: %v", contentType)
			}
			return []domain.Document{
				{ID: "1746187200000", Title: "news title"},
			}, nil
		},
	}

	handler := New(service, logger, Config{BasicUsername: "admin", BasicPassword: "secret"})
	response, err := handler.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/news",
		RequestContext: events.LambdaFunctionURLRequestContext{
			RequestID: "req-1",
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
				Path:   "/news",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func TestHandlePostRequiresAuth(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := New(mockContentService{}, logger, Config{BasicUsername: "admin", BasicPassword: "secret"})

	response, err := handler.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/news",
		Body:    `{"title":"title","content":"content"}`,
		RequestContext: events.LambdaFunctionURLRequestContext{
			RequestID: "req-2",
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodPost,
				Path:   "/news",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func TestHandlePostVideo(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	service := mockContentService{
		createFn: func(_ context.Context, contentType domain.ContentType, input domain.DocumentInput) (domain.Document, error) {
			if contentType != domain.ContentTypeVideo {
				t.Fatalf("unexpected content type: %v", contentType)
			}
			return domain.Document{
				ID:    "1746187200000",
				Title: "[Live video] sample",
				Link:  stringPtr("https://www.youtube.com/watch?v=test"),
			}, nil
		},
	}

	handler := New(service, logger, Config{BasicUsername: "admin", BasicPassword: "secret"})
	response, err := handler.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/video",
		Body:    `{"title":"[Live video] sample","link":"https://www.youtube.com/watch?v=test"}`,
		Headers: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
		},
		RequestContext: events.LambdaFunctionURLRequestContext{
			RequestID: "req-3",
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodPost,
				Path:   "/video",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func TestHandleAuthSuccess(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := New(mockContentService{}, logger, Config{BasicUsername: "admin", BasicPassword: "secret"})

	response, err := handler.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/auth",
		Headers: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
		},
		RequestContext: events.LambdaFunctionURLRequestContext{
			RequestID: "req-auth-1",
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
				Path:   "/auth",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func TestHandleAuthUnauthorized(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := New(mockContentService{}, logger, Config{BasicUsername: "admin", BasicPassword: "secret"})

	response, err := handler.Handle(context.Background(), events.LambdaFunctionURLRequest{
		RawPath: "/auth",
		Headers: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:wrong")),
		},
		RequestContext: events.LambdaFunctionURLRequestContext{
			RequestID: "req-auth-2",
			HTTP: events.LambdaFunctionURLRequestContextHTTPDescription{
				Method: http.MethodGet,
				Path:   "/auth",
			},
		},
	})
	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", response.StatusCode)
	}
}

func stringPtr(value string) *string {
	return &value
}
