package domain

import (
	"net/url"
	"strconv"
	"strings"
)

type ContentType int

const (
	ContentTypeNews ContentType = 1 + iota
	ContentTypeDiscography
	ContentTypeLive
	ContentTypeVideo
)

type Document struct {
	ID               string      `json:"id"`
	Type             ContentType `json:"-"`
	Timestamp        int64       `json:"-"`
	Title            string      `json:"title"`
	Image            *string     `json:"image,omitempty"`
	Content          *string     `json:"content,omitempty"`
	Musics           []string    `json:"musics,omitempty"`
	AppleMusicLink   *string     `json:"applemusic_link,omitempty"`
	SpotifyLink      *string     `json:"spotify_link,omitempty"`
	YouTubeMusicLink *string     `json:"youtubemusic_link,omitempty"`
	LineMusicLink    *string     `json:"linemusic_link,omitempty"`
	AmazonMusicLink  *string     `json:"amazonmusic_link,omitempty"`
	Where            *string     `json:"where,omitempty"`
	With             []string    `json:"with,omitempty"`
	Ticket           *string     `json:"ticket,omitempty"`
	Time             *string     `json:"time,omitempty"`
	Link             *string     `json:"link,omitempty"`
}

type DocumentInput struct {
	Title            *string   `json:"title"`
	Image            *string   `json:"image"`
	Content          *string   `json:"content"`
	Musics           *[]string `json:"musics"`
	AppleMusicLink   *string   `json:"applemusic_link"`
	SpotifyLink      *string   `json:"spotify_link"`
	YouTubeMusicLink *string   `json:"youtubemusic_link"`
	LineMusicLink    *string   `json:"linemusic_link"`
	AmazonMusicLink  *string   `json:"amazonmusic_link"`
	Where            *string   `json:"where"`
	With             *[]string `json:"with"`
	Ticket           *string   `json:"ticket"`
	Time             *string   `json:"time"`
	Link             *string   `json:"link"`
}

func ParseContentType(resource string) (ContentType, bool) {
	switch resource {
	case "news":
		return ContentTypeNews, true
	case "discography":
		return ContentTypeDiscography, true
	case "live":
		return ContentTypeLive, true
	case "video":
		return ContentTypeVideo, true
	default:
		return 0, false
	}
}

func ParseDocumentID(raw string) (int64, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, InvalidArgument("id is required")
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, InvalidArgument("id must be a positive unix millisecond string")
	}

	return value, nil
}

func BuildDocument(contentType ContentType, timestamp int64, input DocumentInput) (Document, error) {
	title, err := normalizeRequiredString("title", input.Title, 1, 200)
	if err != nil {
		return Document{}, err
	}

	image, err := normalizeOptionalURL("image", input.Image)
	if err != nil {
		return Document{}, err
	}

	doc := Document{
		ID:        strconv.FormatInt(timestamp, 10),
		Type:      contentType,
		Timestamp: timestamp,
		Title:     title,
		Image:     image,
	}

	switch contentType {
	case ContentTypeNews:
		content, err := normalizeRequiredString("content", input.Content, 1, 10000)
		if err != nil {
			return Document{}, err
		}
		doc.Content = &content
	case ContentTypeDiscography:
		musics, err := normalizeRequiredStringList("musics", input.Musics, 1, 200)
		if err != nil {
			return Document{}, err
		}

		appleMusicLink, err := normalizeRequiredURL("applemusic_link", input.AppleMusicLink)
		if err != nil {
			return Document{}, err
		}
		spotifyLink, err := normalizeRequiredURL("spotify_link", input.SpotifyLink)
		if err != nil {
			return Document{}, err
		}
		youTubeMusicLink, err := normalizeRequiredURL("youtubemusic_link", input.YouTubeMusicLink)
		if err != nil {
			return Document{}, err
		}
		lineMusicLink, err := normalizeRequiredURL("linemusic_link", input.LineMusicLink)
		if err != nil {
			return Document{}, err
		}
		amazonMusicLink, err := normalizeRequiredURL("amazonmusic_link", input.AmazonMusicLink)
		if err != nil {
			return Document{}, err
		}

		doc.Musics = musics
		doc.AppleMusicLink = &appleMusicLink
		doc.SpotifyLink = &spotifyLink
		doc.YouTubeMusicLink = &youTubeMusicLink
		doc.LineMusicLink = &lineMusicLink
		doc.AmazonMusicLink = &amazonMusicLink
	case ContentTypeLive:
		where, err := normalizeRequiredString("where", input.Where, 1, 200)
		if err != nil {
			return Document{}, err
		}
		with, err := normalizeRequiredStringList("with", input.With, 1, 200)
		if err != nil {
			return Document{}, err
		}
		ticket, err := normalizeRequiredString("ticket", input.Ticket, 1, 200)
		if err != nil {
			return Document{}, err
		}
		timeValue, err := normalizeRequiredString("time", input.Time, 1, 100)
		if err != nil {
			return Document{}, err
		}
		link, err := normalizeRequiredURL("link", input.Link)
		if err != nil {
			return Document{}, err
		}

		doc.Where = &where
		doc.With = with
		doc.Ticket = &ticket
		doc.Time = &timeValue
		doc.Link = &link
	case ContentTypeVideo:
		link, err := normalizeRequiredURL("link", input.Link)
		if err != nil {
			return Document{}, err
		}
		doc.Link = &link
	default:
		return Document{}, Internal("unsupported content type")
	}

	return doc, nil
}

func normalizeRequiredString(name string, value *string, minLength, maxLength int) (string, error) {
	if value == nil {
		return "", InvalidArgument(name + " is required")
	}

	normalized := strings.TrimSpace(*value)
	if len(normalized) < minLength || len(normalized) > maxLength {
		return "", InvalidArgument(name + " is invalid")
	}

	return normalized, nil
}

func normalizeRequiredStringList(name string, values *[]string, minLength, maxLength int) ([]string, error) {
	if values == nil || len(*values) == 0 {
		return nil, InvalidArgument(name + " is required")
	}

	normalized := make([]string, 0, len(*values))
	for _, value := range *values {
		trimmed := strings.TrimSpace(value)
		if len(trimmed) < minLength || len(trimmed) > maxLength {
			return nil, InvalidArgument(name + " contains an invalid value")
		}
		normalized = append(normalized, trimmed)
	}

	return normalized, nil
}

func normalizeRequiredURL(name string, value *string) (string, error) {
	if value == nil {
		return "", InvalidArgument(name + " is required")
	}

	normalized, err := normalizeURL(*value)
	if err != nil {
		return "", InvalidArgument(name + " must be a valid url")
	}
	if normalized == "" {
		return "", InvalidArgument(name + " is required")
	}

	return normalized, nil
}

func normalizeOptionalURL(name string, value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}

	normalized, err := normalizeURL(*value)
	if err != nil {
		return nil, InvalidArgument(name + " must be a valid url")
	}

	if normalized == "" {
		return nil, nil
	}

	return &normalized, nil
}

func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", url.InvalidHostError(parsed.Host)
	}

	if parsed.Host == "" {
		return "", url.InvalidHostError(parsed.Host)
	}

	return parsed.String(), nil
}
