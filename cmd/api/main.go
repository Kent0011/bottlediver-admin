package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/Kent0011/bottlediver-admin/internal/handler"
	"github.com/Kent0011/bottlediver-admin/internal/repository"
	"github.com/Kent0011/bottlediver-admin/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Error("failed to load aws config", "error", err)
		os.Exit(1)
	}

	tableName, ok := os.LookupEnv("TABLE_NAME")
	if !ok || strings.TrimSpace(tableName) == "" {
		logger.Error("TABLE_NAME is required")
		os.Exit(1)
	}

	username, ok := os.LookupEnv("BASIC_AUTH_USERNAME")
	if !ok || strings.TrimSpace(username) == "" {
		logger.Error("BASIC_AUTH_USERNAME is required")
		os.Exit(1)
	}

	password, ok := os.LookupEnv("BASIC_AUTH_PASSWORD")
	if !ok || strings.TrimSpace(password) == "" {
		logger.Error("BASIC_AUTH_PASSWORD is required")
		os.Exit(1)
	}

	allowedOrigins := splitCSV(os.Getenv("ALLOWED_ORIGINS"))

	dynamoClient := dynamodb.NewFromConfig(cfg)
	repo := repository.NewDynamoRepository(dynamoClient, tableName, logger)
	service := usecase.NewContentService(repo)
	h := handler.New(service, logger, handler.Config{
		BasicUsername:  username,
		BasicPassword:  password,
		AllowedOrigins: allowedOrigins,
	})

	lambda.Start(func(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
		return h.Handle(ctx, req)
	})
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}

	return values
}
