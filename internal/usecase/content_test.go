package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
)

type mockRepository struct {
	listFn   func(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error)
	createFn func(ctx context.Context, doc domain.Document) error
	updateFn func(ctx context.Context, doc domain.Document) error
	deleteFn func(ctx context.Context, contentType domain.ContentType, timestamp int64) error
}

func (m mockRepository) List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error) {
	return m.listFn(ctx, contentType)
}

func (m mockRepository) Create(ctx context.Context, doc domain.Document) error {
	return m.createFn(ctx, doc)
}

func (m mockRepository) Update(ctx context.Context, doc domain.Document) error {
	return m.updateFn(ctx, doc)
}

func (m mockRepository) Delete(ctx context.Context, contentType domain.ContentType, timestamp int64) error {
	return m.deleteFn(ctx, contentType, timestamp)
}

func TestCreateNews(t *testing.T) {
	var saved domain.Document
	service := NewContentService(mockRepository{
		createFn: func(_ context.Context, doc domain.Document) error {
			saved = doc
			return nil
		},
	})
	service.now = func() time.Time {
		return time.UnixMilli(1746187200000)
	}

	title := "news title"
	content := "news content"
	doc, err := service.Create(context.Background(), domain.ContentTypeNews, domain.DocumentInput{
		Title:   &title,
		Content: &content,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if doc.ID != "1746187200000" {
		t.Fatalf("unexpected id: %s", doc.ID)
	}

	if saved.Title != "news title" {
		t.Fatalf("unexpected title: %s", saved.Title)
	}
}

func TestUpdateRejectsInvalidID(t *testing.T) {
	service := NewContentService(mockRepository{})
	title := "video title"
	link := "https://example.com/video"

	_, err := service.Update(context.Background(), domain.ContentTypeVideo, "invalid", domain.DocumentInput{
		Title: &title,
		Link:  &link,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INVALID_ARGUMENT" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUsesParsedID(t *testing.T) {
	called := false
	service := NewContentService(mockRepository{
		deleteFn: func(_ context.Context, contentType domain.ContentType, timestamp int64) error {
			called = true
			if contentType != domain.ContentTypeLive {
				t.Fatalf("unexpected content type: %v", contentType)
			}
			if timestamp != 1746187200000 {
				t.Fatalf("unexpected timestamp: %d", timestamp)
			}
			return nil
		},
	})

	if err := service.Delete(context.Background(), domain.ContentTypeLive, "1746187200000"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	if !called {
		t.Fatal("delete repository was not called")
	}
}

func TestCreateVideoRejectsEmptyLink(t *testing.T) {
	service := NewContentService(mockRepository{})
	title := "video title"
	link := "   "

	_, err := service.Create(context.Background(), domain.ContentTypeVideo, domain.DocumentInput{
		Title: &title,
		Link:  &link,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	var appErr *domain.AppError
	if !errors.As(err, &appErr) || appErr.Code != "INVALID_ARGUMENT" {
		t.Fatalf("unexpected error: %v", err)
	}
}
