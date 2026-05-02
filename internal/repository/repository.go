package repository

import (
	"context"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
)

type DocumentRepository interface {
	List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error)
	Create(ctx context.Context, doc domain.Document) error
	Update(ctx context.Context, doc domain.Document) error
	Delete(ctx context.Context, contentType domain.ContentType, timestamp int64) error
}
