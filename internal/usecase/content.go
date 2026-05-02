package usecase

import (
	"context"
	"time"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
	"github.com/Kent0011/bottlediver-admin/internal/repository"
)

type ContentService struct {
	repo repository.DocumentRepository
	now  func() time.Time
}

func NewContentService(repo repository.DocumentRepository) *ContentService {
	return &ContentService{
		repo: repo,
		now:  time.Now,
	}
}

func (s *ContentService) List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error) {
	return s.repo.List(ctx, contentType)
}

func (s *ContentService) Create(ctx context.Context, contentType domain.ContentType, input domain.DocumentInput) (domain.Document, error) {
	doc, err := domain.BuildDocument(contentType, s.now().UTC().UnixMilli(), input)
	if err != nil {
		return domain.Document{}, err
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return domain.Document{}, err
	}

	return doc, nil
}

func (s *ContentService) Update(ctx context.Context, contentType domain.ContentType, id string, input domain.DocumentInput) (domain.Document, error) {
	timestamp, err := domain.ParseDocumentID(id)
	if err != nil {
		return domain.Document{}, err
	}

	doc, err := domain.BuildDocument(contentType, timestamp, input)
	if err != nil {
		return domain.Document{}, err
	}

	if err := s.repo.Update(ctx, doc); err != nil {
		return domain.Document{}, err
	}

	return doc, nil
}

func (s *ContentService) Delete(ctx context.Context, contentType domain.ContentType, id string) error {
	timestamp, err := domain.ParseDocumentID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, contentType, timestamp)
}
