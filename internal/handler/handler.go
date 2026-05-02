package handler

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/Kent0011/bottlediver-admin/internal/domain"
)

type contentService interface {
	List(ctx context.Context, contentType domain.ContentType) ([]domain.Document, error)
	Create(ctx context.Context, contentType domain.ContentType, input domain.DocumentInput) (domain.Document, error)
	Update(ctx context.Context, contentType domain.ContentType, id string, input domain.DocumentInput) (domain.Document, error)
	Delete(ctx context.Context, contentType domain.ContentType, id string) error
}

type Config struct {
	BasicUsername  string
	BasicPassword  string
	AllowedOrigins []string
}

type Handler struct {
	service        contentService
	logger         *slog.Logger
	basicUsername  string
	basicPassword  string
	allowedOrigins map[string]struct{}
}

type errorResponse struct {
	Error     errorBody `json:"error"`
	RequestID string    `json:"request_id"`
}

type errorBody struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

func New(service contentService, logger *slog.Logger, cfg Config) *Handler {
	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}

	return &Handler{
		service:        service,
		logger:         logger,
		basicUsername:  cfg.BasicUsername,
		basicPassword:  cfg.BasicPassword,
		allowedOrigins: allowedOrigins,
	}
}

func (h *Handler) Handle(ctx context.Context, req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	requestID := req.RequestContext.RequestID
	path := normalizedPath(req)
	method := strings.ToUpper(req.RequestContext.HTTP.Method)
	if method == "" {
		method = strings.ToUpper(req.RequestContext.HTTP.Method)
	}

	h.logger.Info("incoming request", "endpoint", path, "method", method, "request_body", req.Body)

	if strings.Trim(path, "/") == "auth" {
		if method == http.MethodOptions {
			return h.emptyResponse(req, requestID, http.StatusNoContent), nil
		}
		if method != http.MethodGet {
			return h.errorResponse(req, requestID, domain.NotFound("route not found")), nil
		}
		if err := h.requireBasicAuth(req.Headers); err != nil {
			return h.errorResponse(req, requestID, err), nil
		}
		return h.jsonResponse(req, requestID, http.StatusOK, map[string]any{"authenticated": true}), nil
	}

	contentType, id, hasID, ok := parseRoute(path)
	if !ok {
		return h.errorResponse(req, requestID, domain.NotFound("route not found")), nil
	}

	if method == http.MethodOptions {
		return h.emptyResponse(req, requestID, http.StatusNoContent), nil
	}

	switch {
	case method == http.MethodGet && !hasID:
		items, err := h.service.List(ctx, contentType)
		if err != nil {
			return h.errorResponse(req, requestID, err), nil
		}
		return h.jsonResponse(req, requestID, http.StatusOK, map[string]any{"items": items}), nil
	case method == http.MethodPost && !hasID:
		if err := h.requireBasicAuth(req.Headers); err != nil {
			return h.errorResponse(req, requestID, err), nil
		}

		input, err := decodeInput(req.Body)
		if err != nil {
			return h.errorResponse(req, requestID, err), nil
		}

		doc, err := h.service.Create(ctx, contentType, input)
		if err != nil {
			return h.errorResponse(req, requestID, err), nil
		}
		return h.jsonResponse(req, requestID, http.StatusCreated, doc), nil
	case method == http.MethodPut && hasID:
		if err := h.requireBasicAuth(req.Headers); err != nil {
			return h.errorResponse(req, requestID, err), nil
		}

		input, err := decodeInput(req.Body)
		if err != nil {
			return h.errorResponse(req, requestID, err), nil
		}

		doc, err := h.service.Update(ctx, contentType, id, input)
		if err != nil {
			return h.errorResponse(req, requestID, err), nil
		}
		return h.jsonResponse(req, requestID, http.StatusOK, doc), nil
	case method == http.MethodDelete && hasID:
		if err := h.requireBasicAuth(req.Headers); err != nil {
			return h.errorResponse(req, requestID, err), nil
		}

		if err := h.service.Delete(ctx, contentType, id); err != nil {
			return h.errorResponse(req, requestID, err), nil
		}
		return h.emptyResponse(req, requestID, http.StatusNoContent), nil
	default:
		return h.errorResponse(req, requestID, domain.NotFound("route not found")), nil
	}
}

func decodeInput(body string) (domain.DocumentInput, error) {
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()

	var input domain.DocumentInput
	if err := decoder.Decode(&input); err != nil {
		return domain.DocumentInput{}, domain.InvalidArgument("request body is invalid")
	}

	return input, nil
}

func parseRoute(path string) (domain.ContentType, string, bool, bool) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return 0, "", false, false
	}

	parts := strings.Split(trimmed, "/")
	contentType, ok := domain.ParseContentType(parts[0])
	if !ok {
		return 0, "", false, false
	}

	if len(parts) == 1 {
		return contentType, "", false, true
	}

	if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		return contentType, parts[1], true, true
	}

	return 0, "", false, false
}

func normalizedPath(req events.LambdaFunctionURLRequest) string {
	if strings.TrimSpace(req.RawPath) != "" {
		return req.RawPath
	}

	if strings.TrimSpace(req.RequestContext.HTTP.Path) != "" {
		return req.RequestContext.HTTP.Path
	}

	return "/"
}

func (h *Handler) requireBasicAuth(headers map[string]string) error {
	raw := headerValue(headers, "Authorization")
	if !strings.HasPrefix(strings.ToLower(raw), "basic ") {
		return domain.Unauthenticated("basic authorization is required")
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw[6:]))
	if err != nil {
		return domain.Unauthenticated("basic authorization is invalid")
	}

	username, password, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return domain.Unauthenticated("basic authorization is invalid")
	}

	if subtle.ConstantTimeCompare([]byte(username), []byte(h.basicUsername)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(h.basicPassword)) != 1 {
		return domain.Unauthenticated("basic authorization is invalid")
	}

	return nil
}

func headerValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}

	return ""
}

func (h *Handler) jsonResponse(req events.LambdaFunctionURLRequest, requestID string, status int, payload any) events.LambdaFunctionURLResponse {
	body, err := json.Marshal(payload)
	if err != nil {
		return h.errorResponse(req, requestID, domain.Internal("failed to encode response"))
	}

	h.logger.Info("outgoing response", "endpoint", normalizedPath(req), "status", status, "response", string(body))

	headers := h.baseHeaders(req, requestID)
	headers["Content-Type"] = "application/json; charset=utf-8"

	return events.LambdaFunctionURLResponse{
		StatusCode: status,
		Headers:    headers,
		Body:       string(body),
	}
}

func (h *Handler) emptyResponse(req events.LambdaFunctionURLRequest, requestID string, status int) events.LambdaFunctionURLResponse {
	h.logger.Info("outgoing response", "endpoint", normalizedPath(req), "status", status, "response", "")

	return events.LambdaFunctionURLResponse{
		StatusCode: status,
		Headers:    h.baseHeaders(req, requestID),
	}
}

func (h *Handler) errorResponse(req events.LambdaFunctionURLRequest, requestID string, err error) events.LambdaFunctionURLResponse {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		appErr = domain.Internal("internal server error")
	}

	payload := errorResponse{
		Error: errorBody{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: []string{},
		},
		RequestID: requestID,
	}

	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		body = []byte(`{"error":{"code":"INTERNAL","message":"internal server error","details":[]},"request_id":"` + requestID + `"}`)
		appErr = domain.Internal("internal server error")
	}

	h.logger.Error("request failed", "endpoint", normalizedPath(req), "status", appErr.HTTPStatus, "error", appErr.Message, "response", string(body))

	headers := h.baseHeaders(req, requestID)
	headers["Content-Type"] = "application/json; charset=utf-8"

	return events.LambdaFunctionURLResponse{
		StatusCode: appErr.HTTPStatus,
		Headers:    headers,
		Body:       string(body),
	}
}

func (h *Handler) baseHeaders(req events.LambdaFunctionURLRequest, requestID string) map[string]string {
	headers := map[string]string{
		"X-Request-Id":                 requestID,
		"Access-Control-Allow-Headers": "Authorization, Content-Type",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
	}

	if origin := h.allowedOrigin(headerValue(req.Headers, "Origin")); origin != "" {
		headers["Access-Control-Allow-Origin"] = origin
		if origin != "*" {
			headers["Access-Control-Allow-Credentials"] = "true"
		}
	}

	return headers
}

func (h *Handler) allowedOrigin(requestOrigin string) string {
	if len(h.allowedOrigins) == 0 {
		return "*"
	}

	if _, ok := h.allowedOrigins[requestOrigin]; ok {
		return requestOrigin
	}

	return ""
}
