// Package groupme defines a client capable of executing API commands for the GroupMe chat service
package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GroupMe documentation: https://dev.groupme.com/docs/image_service

const imageServiceUploadURL = "https://image.groupme.com/pictures"

type imageUploadResponse struct {
	Payload struct {
		URL string `json:"url"`
	} `json:"payload"`
}

// UploadImage returns the i.groupme.com URL to use in an image Attachment.
func (c *Client) UploadImage(ctx context.Context, data []byte, contentType string) (string, error) {
	if contentType == "" {
		contentType = "image/jpeg"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", imageServiceUploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", contentType)
	httpReq.Header.Set("X-Access-Token", c.authorizationToken)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("image upload request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read image upload response: %w", err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("image upload failed with status %d", httpResp.StatusCode)
	}

	var resp imageUploadResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse image upload response: %w", err)
	}
	if resp.Payload.URL == "" {
		return "", fmt.Errorf("image upload response had no payload.url")
	}
	return resp.Payload.URL, nil
}
