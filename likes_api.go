// Package groupme defines a client capable of executing API commands for the GroupMe chat service
package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GroupMe documentation: https://dev.groupme.com/docs/v3#likes
//
// Per-emoji reactions are not in the v3 docs; see
// https://groupme-js.github.io/GroupMeCommunityDocs/api/conversations/reactions/.
// Powerup emoji like icons ({"type":"emoji","pack_id":N,"pack_index":N}) are
// not implemented.

/*//////// Endpoints ////////*/
const (
	// Used to build other endpoints
	likesEndpointRoot = "/messages/%s/%s"

	createLikeEndpoint  = likesEndpointRoot + "/like"   // POST
	destroyLikeEndpoint = likesEndpointRoot + "/unlike" // POST
)

// UnicodeLikeIcons is the fixed set of unicode emoji GroupMe accepts as a like icon.
var UnicodeLikeIcons = map[string]bool{
	"❤️": true, "👍": true, "🤣": true, "🎉": true, "🔥": true,
	"😮": true, "👀": true, "😭": true, "🥺": true, "🙏": true,
	"💀": true, "🫶": true, "🤬": true, "💅": true, "🫠": true,
}

/*//////// API Requests ////////*/

// Create

type likeIcon struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

type createLikeBody struct {
	LikeIcon *likeIcon `json:"like_icon,omitempty"`
}

// CreateLike - Like a message. If emoji is one of UnicodeLikeIcons, reacts
// with that specific emoji; otherwise (including an empty string) falls
// back to the plain generic heart like, same as calling this before emoji
// support was added here.
func (c *Client) CreateLike(ctx context.Context, conversationID, messageID ID, emoji string) error {
	url := fmt.Sprintf(c.endpointBase+createLikeEndpoint, conversationID, messageID)

	var body *createLikeBody
	if UnicodeLikeIcons[emoji] {
		body = &createLikeBody{LikeIcon: &likeIcon{Type: "unicode", Code: emoji}}
	}

	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	httpReq, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	return c.doWithAuthToken(ctx, httpReq, nil)
}

// DestroyLike - Unlike a message.
func (c *Client) DestroyLike(ctx context.Context, conversationID, messageID ID) error {
	url := fmt.Sprintf(c.endpointBase+destroyLikeEndpoint, conversationID, messageID)

	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	return c.doWithAuthToken(ctx, httpReq, nil)
}
