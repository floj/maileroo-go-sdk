package maileroo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	APIKey  string
	Timeout time.Duration
	http    *http.Client
}

const (
	APIBaseURL                   = "https://smtp.maileroo.com/api/v2/"
	MaxAssociativeMapKeyLength   = 128
	MaxAssociativeMapValueLength = 768
	MaxSubjectLength             = 255
	ReferenceIDLength            = 24 // hex chars
	maxBulkMessages              = 500
	defaultUserAgent             = "maileroo-go-sdk/1.0"
)

type AssocValue = any
type AssocMap = map[string]AssocValue

type BasicEmailData struct {
	From        EmailAddress   `json:"-"`
	To          []EmailAddress `json:"-"`
	Cc          []EmailAddress `json:"-"`
	Bcc         []EmailAddress `json:"-"`
	ReplyTo     []EmailAddress `json:"-"`
	Subject     string         `json:"-"`
	HTML        *string        `json:"-"`
	Plain       *string        `json:"-"`
	Tracking    *bool          `json:"-"`
	Tags        AssocMap       `json:"-"`
	Headers     AssocMap       `json:"-"`
	Attachments []Attachment   `json:"-"`
	ScheduledAt *time.Time     `json:"-"`
	ReferenceID *string        `json:"-"`
}

type TemplatedEmailData struct {
	From         EmailAddress   `json:"-"`
	To           []EmailAddress `json:"-"`
	Cc           []EmailAddress `json:"-"`
	Bcc          []EmailAddress `json:"-"`
	ReplyTo      []EmailAddress `json:"-"`
	Subject      string         `json:"-"`
	TemplateID   int            `json:"-"`
	TemplateData map[string]any `json:"-"`
	Tracking     *bool          `json:"-"`
	Tags         AssocMap       `json:"-"`
	Headers      AssocMap       `json:"-"`
	Attachments  []Attachment   `json:"-"`
	ScheduledAt  *time.Time     `json:"-"`
	ReferenceID  *string        `json:"-"`
}

type BulkMessage struct {
	From         EmailAddress   `json:"-"`
	To           []EmailAddress `json:"-"`
	Cc           []EmailAddress `json:"-"`
	Bcc          []EmailAddress `json:"-"`
	ReplyTo      []EmailAddress `json:"-"`
	ReferenceID  *string        `json:"-"`
	TemplateData map[string]any `json:"-"`
}

type BulkEmailData struct {
	Subject     string        `json:"-"`
	HTML        *string       `json:"-"`
	Plain       *string       `json:"-"`
	TemplateID  *int          `json:"-"`
	Tracking    *bool         `json:"-"`
	Tags        AssocMap      `json:"-"`
	Headers     AssocMap      `json:"-"`
	Attachments []Attachment  `json:"-"`
	Messages    []BulkMessage `json:"-"`
}

type ScheduledEmailsResponse struct {
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalCount int           `json:"total_count"`
	TotalPages int           `json:"total_pages"`
	Items      []interface{} `json:"results"`
}

type BasePayload struct {
	Subject     string
	From        EmailAddress
	To          []EmailAddress
	Cc          []EmailAddress
	Bcc         []EmailAddress
	ReplyTo     []EmailAddress
	Tracking    *bool
	Tags        AssocMap
	Headers     AssocMap
	Attachments []Attachment
	ScheduledAt *time.Time
	ReferenceID *string
}

func NewClient(apiKey string, timeoutSeconds int) (*Client, error) {

	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("API key must be a non-empty string")
	}

	if timeoutSeconds <= 0 {
		return nil, errors.New("timeout must be a positive integer")
	}

	return &Client{
		APIKey:  apiKey,
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		http: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}, nil

}

func (c *Client) GetReferenceID() string {

	const byteLen = ReferenceIDLength / 2

	b := make([]byte, byteLen)
	_, err := rand.Read(b)

	if err != nil {

		for i := range b {
			b[i] = byte(mathrand.Intn(256))
		}

	}

	return hex.EncodeToString(b)

}

func (c *Client) SendBasicEmail(ctx context.Context, data BasicEmailData) (string, error) {

	payload := BasePayload{
		Subject:     data.Subject,
		From:        data.From,
		To:          data.To,
		Cc:          data.Cc,
		Bcc:         data.Bcc,
		ReplyTo:     data.ReplyTo,
		Tracking:    data.Tracking,
		Tags:        data.Tags,
		Headers:     data.Headers,
		Attachments: data.Attachments,
		ScheduledAt: data.ScheduledAt,
		ReferenceID: data.ReferenceID,
	}

	basePayload, err := c.buildBasePayload(payload)

	if err != nil {
		return "", err
	}

	if data.HTML == nil && data.Plain == nil {
		return "", errors.New("either html or plain body is required")
	}

	basePayload["html"] = data.HTML
	basePayload["plain"] = data.Plain

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			ReferenceID string `json:"reference_id"`
		} `json:"data"`
	}

	if err := c.sendRequest(ctx, http.MethodPost, "emails", basePayload, &out); err != nil {
		return "", err
	}

	if out.Success {
		return out.Data.ReferenceID, nil
	}

	if out.Message == "" {
		out.Message = "Unknown"
	}

	return "", fmt.Errorf("the API returned an error: %s", out.Message)

}

func (c *Client) SendTemplatedEmail(ctx context.Context, data TemplatedEmailData) (string, error) {

	payload := BasePayload{
		Subject:     data.Subject,
		From:        data.From,
		To:          data.To,
		Cc:          data.Cc,
		Bcc:         data.Bcc,
		ReplyTo:     data.ReplyTo,
		Tracking:    data.Tracking,
		Tags:        data.Tags,
		Headers:     data.Headers,
		Attachments: data.Attachments,
		ScheduledAt: data.ScheduledAt,
		ReferenceID: data.ReferenceID,
	}

	basePayload, err := c.buildBasePayload(payload)

	if err != nil {
		return "", err
	}

	basePayload["template_id"] = data.TemplateID

	if data.TemplateData != nil {

		if err := validateTemplateData(data.TemplateData); err != nil {
			return "", err
		}

		basePayload["template_data"] = data.TemplateData

	}

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			ReferenceID string `json:"reference_id"`
		} `json:"data"`
	}

	if err := c.sendRequest(ctx, http.MethodPost, "emails/template", basePayload, &out); err != nil {
		return "", err
	}

	if out.Success {
		return out.Data.ReferenceID, nil
	}

	if out.Message == "" {
		out.Message = "Unknown"
	}

	return "", fmt.Errorf("the API returned an error: %s", out.Message)

}

func (c *Client) SendBulkEmails(ctx context.Context, data BulkEmailData) ([]string, error) {

	if err := requireSubject(data.Subject); err != nil {
		return nil, err
	}

	hasHTML := data.HTML != nil
	hasPlain := data.Plain != nil
	hasTemplateID := data.TemplateID != nil

	if (!hasHTML && !hasPlain) && !hasTemplateID {
		return nil, errors.New("you must provide either html, plain, or template_id")
	}

	if data.TemplateID != nil && (hasHTML || hasPlain) {
		return nil, errors.New("template_id cannot be combined with html or plain")
	}

	if len(data.Messages) == 0 {
		return nil, errors.New("messages must be a non-empty array")
	}

	if len(data.Messages) > maxBulkMessages {
		return nil, fmt.Errorf("messages cannot contain more than %d items", maxBulkMessages)
	}

	payload := map[string]any{
		"subject": data.Subject,
	}

	if hasHTML {
		payload["html"] = data.HTML
	}

	if hasPlain {
		payload["plain"] = data.Plain
	}

	if hasTemplateID {
		payload["template_id"] = data.TemplateID
	}

	if data.Tracking != nil {
		payload["tracking"] = *data.Tracking
	}

	if data.Tags != nil {

		if err := validateAssociativeMap(data.Tags, "tags"); err != nil {
			return nil, err
		}

		payload["tags"] = data.Tags

	}

	if data.Headers != nil {

		if err := validateAssociativeMap(data.Headers, "headers"); err != nil {
			return nil, err
		}

		payload["headers"] = data.Headers

	}

	if len(data.Attachments) > 0 {

		arr := make([]Attachment, 0, len(data.Attachments))

		for _, att := range data.Attachments {

			if err := att.validate(); err != nil {
				return nil, err
			}

			arr = append(arr, att)

		}

		payload["attachments"] = arr

	}

	msgs, err := c.normalizeBulkMessages(data.Messages)

	if err != nil {
		return nil, err
	}

	payload["messages"] = msgs

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			ReferenceIDs []string `json:"reference_ids"`
		} `json:"data"`
	}

	if err := c.sendRequest(ctx, http.MethodPost, "emails/bulk", payload, &out); err != nil {
		return nil, err
	}

	if out.Success {
		return out.Data.ReferenceIDs, nil
	}

	if out.Message == "" {
		out.Message = "Unknown"
	}

	return nil, fmt.Errorf("the API returned an error: %s", out.Message)

}

func (c *Client) DeleteScheduledEmail(ctx context.Context, referenceID string) error {

	if err := validateReferenceID(referenceID); err != nil {
		return err
	}

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	path := "emails/scheduled/" + referenceID

	if err := c.sendRequest(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return err
	}

	if out.Success {
		return nil
	}

	if out.Message == "" {
		out.Message = "Unknown"
	}

	return fmt.Errorf("the API returned an error: %s", out.Message)

}

func (c *Client) GetScheduledEmails(ctx context.Context, page, perPage int) (*ScheduledEmailsResponse, error) {

	if page < 1 {
		return nil, errors.New("page must be a positive integer (>= 1)")
	}

	if perPage < 1 {
		return nil, errors.New("per_page must be a positive integer (>= 1)")
	}

	if perPage > 100 {
		return nil, errors.New("per_page cannot be greater than 100")
	}

	q := url.Values{}

	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("per_page", fmt.Sprintf("%d", perPage))

	var out struct {
		Success bool                     `json:"success"`
		Message string                   `json:"message"`
		Data    *ScheduledEmailsResponse `json:"data"`
	}

	if err := c.sendRequest(ctx, http.MethodGet, "emails/scheduled?"+q.Encode(), nil, &out); err != nil {
		return nil, err
	}

	if out.Success && out.Data != nil {
		return out.Data, nil
	}

	if out.Message == "" {
		out.Message = "Unknown"
	}

	return nil, fmt.Errorf("the API returned an error: %s", out.Message)

}

func (c *Client) buildBasePayload(payload BasePayload) (map[string]any, error) {

	if err := requireSubject(payload.Subject); err != nil {
		return nil, err
	}

	if len(payload.To) == 0 {
		return nil, errors.New("field to is required and must have at least one recipient")
	}

	result := map[string]any{
		"subject": payload.Subject,
	}

	result["from"] = payload.From.ToJSON()
	result["to"] = emailAddressesToJSON(payload.To)

	if len(payload.Cc) > 0 {
		result["cc"] = emailAddressesToJSON(payload.Cc)
	}

	if len(payload.Bcc) > 0 {
		result["bcc"] = emailAddressesToJSON(payload.Bcc)
	}

	if len(payload.ReplyTo) > 0 {
		result["reply_to"] = emailAddressesToJSON(payload.ReplyTo)
	}

	if payload.Tracking != nil {
		result["tracking"] = *payload.Tracking
	}

	if payload.Tags != nil {

		if err := validateAssociativeMap(payload.Tags, "tags"); err != nil {
			return nil, err
		}

		result["tags"] = payload.Tags

	}
	if payload.Headers != nil {

		if err := validateAssociativeMap(payload.Headers, "headers"); err != nil {
			return nil, err
		}

		result["headers"] = payload.Headers

	}
	if len(payload.Attachments) > 0 {

		arr := make([]Attachment, 0, len(payload.Attachments))

		for _, att := range payload.Attachments {

			if err := att.validate(); err != nil {
				return nil, err
			}

			arr = append(arr, att)

		}

		result["attachments"] = arr

	}

	if payload.ScheduledAt != nil {
		result["scheduled_at"] = *payload.ScheduledAt
	}

	if payload.ReferenceID != nil {

		if err := validateReferenceID(*payload.ReferenceID); err != nil {
			return nil, err
		}

		result["reference_id"] = *payload.ReferenceID

	} else {

		result["reference_id"] = c.GetReferenceID()

	}

	return result, nil

}

func (c *Client) normalizeBulkMessages(in []BulkMessage) ([]map[string]any, error) {

	out := make([]map[string]any, 0, len(in))

	for i, m := range in {

		if len(m.To) == 0 {
			return nil, fmt.Errorf("messages[%d].to must have at least one recipient", i)
		}

		item := map[string]any{
			"from": m.From.ToJSON(),
			"to":   emailAddressesToJSON(m.To),
		}

		if len(m.Cc) > 0 {
			item["cc"] = emailAddressesToJSON(m.Cc)
		}

		if len(m.Bcc) > 0 {
			item["bcc"] = emailAddressesToJSON(m.Bcc)
		}

		if len(m.ReplyTo) > 0 {
			item["reply_to"] = emailAddressesToJSON(m.ReplyTo)
		}

		if m.ReferenceID != nil {

			if err := validateReferenceID(*m.ReferenceID); err != nil {
				return nil, fmt.Errorf("messages[%d].reference_id: %w", i, err)
			}

			item["reference_id"] = *m.ReferenceID

		} else {

			item["reference_id"] = c.GetReferenceID()

		}

		if m.TemplateData != nil {

			if err := validateTemplateData(m.TemplateData); err != nil {
				return nil, fmt.Errorf("messages[%d].template_data: %w", i, err)
			}

			item["template_data"] = m.TemplateData

		}

		out = append(out, item)

	}

	return out, nil

}

func (c *Client) sendRequest(ctx context.Context, method, endpoint string, body any, out any) error {

	if !strings.HasPrefix(endpoint, "http") {
		endpoint = APIBaseURL + strings.TrimLeft(endpoint, "/")
	}

	var r io.Reader

	if method == http.MethodGet || body == nil {

		r = nil

	} else {

		b, err := json.Marshal(body)

		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		r = strings.NewReader(string(b))

	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, r)

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := c.http.Do(req)

	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}

	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("the API response is not valid JSON: %v", err)
	}

	return nil

}

var refIDRe = regexp.MustCompile(`^[0-9a-fA-F]{24}$`)

func validateReferenceID(s string) error {

	if s != strings.TrimSpace(s) {
		return errors.New("reference_id must not contain whitespace")
	}

	if !refIDRe.MatchString(s) {
		return fmt.Errorf("reference_id must be a %d-character hexadecimal string", ReferenceIDLength)
	}

	return nil

}

func requireSubject(s string) error {

	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("subject must be a non-empty string with a maximum length of %d characters", MaxSubjectLength)
	}

	if runeLen(s) > MaxSubjectLength {
		return fmt.Errorf("subject must be a non-empty string with a maximum length of %d characters", MaxSubjectLength)
	}

	return nil

}

func validateAssociativeMap(m AssocMap, label string) error {

	for k, v := range m {

		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("%s keys must be non-empty strings", label)
		}

		if runeLen(k) > MaxAssociativeMapKeyLength {
			return fmt.Errorf("%s key must not exceed %d characters", label, MaxAssociativeMapKeyLength)
		}

		if !isAcceptableAssocValue(v) {
			return fmt.Errorf("%s must be an associative map with string keys and values (string/number/bool)", label)
		}

		if valLen(v) > MaxAssociativeMapValueLength {
			return fmt.Errorf("%s value must not exceed %d characters", label, MaxAssociativeMapValueLength)
		}

	}

	return nil

}

func validateTemplateData(m map[string]any) error {

	for k := range m {

		if strings.TrimSpace(k) == "" {
			return errors.New("template_data keys must be strings and non-empty")
		}

	}

	return nil

}

func isAcceptableAssocValue(v any) bool {

	switch v.(type) {

	case string, bool, int, int32, int64, uint, uint32, uint64, float32, float64:
		return true

	default:
		return false

	}

}

func valLen(v any) int {

	switch t := v.(type) {

	case string:
		return runeLen(t)

	case bool:

		if t {
			return 4
		}

		return 5

	case int, int32, int64, uint, uint32, uint64, float32, float64:
		return len(fmt.Sprintf("%v", t))

	default:
		return len(fmt.Sprintf("%v", t))

	}

}

func runeLen(s string) int {
	return len([]rune(s))
}

func StrPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func IntPtr(i int) *int {
	return &i
}
