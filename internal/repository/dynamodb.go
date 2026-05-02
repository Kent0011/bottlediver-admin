package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
)

type DynamoRepository struct {
	client    *dynamodb.Client
	tableName string
	logger    *slog.Logger
}

type dynamoItem struct {
	Type             int      `dynamodbav:"type"`
	Timestamp        int64    `dynamodbav:"timestamp"`
	Title            string   `dynamodbav:"title"`
	Image            *string  `dynamodbav:"image,omitempty"`
	Content          *string  `dynamodbav:"content,omitempty"`
	Musics           []string `dynamodbav:"musics,omitempty"`
	AppleMusicLink   *string  `dynamodbav:"applemusic_link,omitempty"`
	SpotifyLink      *string  `dynamodbav:"spotify_link,omitempty"`
	YouTubeMusicLink *string  `dynamodbav:"youtubemusic_link,omitempty"`
	LineMusicLink    *string  `dynamodbav:"linemusic_link,omitempty"`
	AmazonMusicLink  *string  `dynamodbav:"amazonmusic_link,omitempty"`
	Where            *string  `dynamodbav:"where,omitempty"`
	With             []string `dynamodbav:"with,omitempty"`
	Ticket           *string  `dynamodbav:"ticket,omitempty"`
	Time             *string  `dynamodbav:"time,omitempty"`
	Link             *string  `dynamodbav:"link,omitempty"`
}

func NewDynamoRepository(client *dynamodb.Client, tableName string, logger *slog.Logger) *DynamoRepository {
	return &DynamoRepository{
		client:    client,
		tableName: tableName,
		logger:    logger,
	}
}

func (r *DynamoRepository) List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("#type = :type"),
		ExpressionAttributeNames: map[string]string{
			"#type": "type",
		},
		ExpressionAttributeValues: map[string]dynamodbTypes.AttributeValue{
			":type": &dynamodbTypes.AttributeValueMemberN{Value: strconv.Itoa(int(contentType))},
		},
		ScanIndexForward: aws.Bool(false),
	}

	r.logger.Info("dynamodb query", "operation", "Query", "query", input)

	output, err := r.client.Query(ctx, input)
	if err != nil {
		r.logger.Error("dynamodb query failed", "operation", "Query", "error", err)
		return nil, domain.Internal("failed to list documents")
	}

	r.logger.Info("dynamodb response", "operation", "Query", "count", len(output.Items))

	var items []dynamoItem
	if err := attributevalue.UnmarshalListOfMaps(output.Items, &items); err != nil {
		return nil, domain.Internal("failed to decode documents")
	}

	documents := make([]domain.Document, 0, len(items))
	for _, item := range items {
		documents = append(documents, item.toDomain())
	}

	return documents, nil
}

func (r *DynamoRepository) Create(ctx context.Context, doc domain.Document) error {
	item, err := attributevalue.MarshalMap(newDynamoItem(doc))
	if err != nil {
		return domain.Internal("failed to encode document")
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		ConditionExpression: aws.String(
			"attribute_not_exists(#type) AND attribute_not_exists(#timestamp)",
		),
		ExpressionAttributeNames: map[string]string{
			"#type":      "type",
			"#timestamp": "timestamp",
		},
	}

	r.logger.Info("dynamodb query", "operation", "PutItem", "query", input)

	if _, err := r.client.PutItem(ctx, input); err != nil {
		r.logger.Error("dynamodb query failed", "operation", "PutItem", "error", err)
		return mapConditionalError(err, domain.Conflict("document already exists"), domain.Internal("failed to create document"))
	}

	r.logger.Info("dynamodb response", "operation", "PutItem", "result", "ok")
	return nil
}

func (r *DynamoRepository) Update(ctx context.Context, doc domain.Document) error {
	item, err := attributevalue.MarshalMap(newDynamoItem(doc))
	if err != nil {
		return domain.Internal("failed to encode document")
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
		ConditionExpression: aws.String(
			"attribute_exists(#type) AND attribute_exists(#timestamp)",
		),
		ExpressionAttributeNames: map[string]string{
			"#type":      "type",
			"#timestamp": "timestamp",
		},
	}

	r.logger.Info("dynamodb query", "operation", "PutItem", "query", input)

	if _, err := r.client.PutItem(ctx, input); err != nil {
		r.logger.Error("dynamodb query failed", "operation", "PutItem", "error", err)
		return mapConditionalError(err, domain.NotFound("document not found"), domain.Internal("failed to update document"))
	}

	r.logger.Info("dynamodb response", "operation", "PutItem", "result", "ok")
	return nil
}

func (r *DynamoRepository) Delete(ctx context.Context, contentType domain.ContentType, timestamp int64) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]dynamodbTypes.AttributeValue{
			"type":      &dynamodbTypes.AttributeValueMemberN{Value: strconv.Itoa(int(contentType))},
			"timestamp": &dynamodbTypes.AttributeValueMemberN{Value: strconv.FormatInt(timestamp, 10)},
		},
		ConditionExpression: aws.String(
			"attribute_exists(#type) AND attribute_exists(#timestamp)",
		),
		ExpressionAttributeNames: map[string]string{
			"#type":      "type",
			"#timestamp": "timestamp",
		},
	}

	r.logger.Info("dynamodb query", "operation", "DeleteItem", "query", input)

	if _, err := r.client.DeleteItem(ctx, input); err != nil {
		r.logger.Error("dynamodb query failed", "operation", "DeleteItem", "error", err)
		return mapConditionalError(err, domain.NotFound("document not found"), domain.Internal("failed to delete document"))
	}

	r.logger.Info("dynamodb response", "operation", "DeleteItem", "result", "ok")
	return nil
}

func mapConditionalError(err error, conditionalErr, fallback error) error {
	var conditionalCheckFailed *dynamodbTypes.ConditionalCheckFailedException
	if errors.As(err, &conditionalCheckFailed) {
		return conditionalErr
	}

	return fallback
}

func newDynamoItem(doc domain.Document) dynamoItem {
	return dynamoItem{
		Type:             int(doc.Type),
		Timestamp:        doc.Timestamp,
		Title:            doc.Title,
		Image:            doc.Image,
		Content:          doc.Content,
		Musics:           doc.Musics,
		AppleMusicLink:   doc.AppleMusicLink,
		SpotifyLink:      doc.SpotifyLink,
		YouTubeMusicLink: doc.YouTubeMusicLink,
		LineMusicLink:    doc.LineMusicLink,
		AmazonMusicLink:  doc.AmazonMusicLink,
		Where:            doc.Where,
		With:             doc.With,
		Ticket:           doc.Ticket,
		Time:             doc.Time,
		Link:             doc.Link,
	}
}

func (i dynamoItem) toDomain() domain.Document {
	return domain.Document{
		ID:               strconv.FormatInt(i.Timestamp, 10),
		Type:             domain.ContentType(i.Type),
		Timestamp:        i.Timestamp,
		Title:            i.Title,
		Image:            i.Image,
		Content:          i.Content,
		Musics:           i.Musics,
		AppleMusicLink:   i.AppleMusicLink,
		SpotifyLink:      i.SpotifyLink,
		YouTubeMusicLink: i.YouTubeMusicLink,
		LineMusicLink:    i.LineMusicLink,
		AmazonMusicLink:  i.AmazonMusicLink,
		Where:            i.Where,
		With:             i.With,
		Ticket:           i.Ticket,
		Time:             i.Time,
		Link:             i.Link,
	}
}
