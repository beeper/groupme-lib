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
// The /like endpoint also accepts an optional JSON body specifying which
// emoji to react with -- not present in the old dev.groupme.com v3 docs
// this package was originally written against (GroupMe added per-emoji
// reactions after this library was last touched), confirmed against the
// community-maintained docs at
// https://groupme-js.github.io/GroupMeCommunityDocs/api/conversations/reactions/
// and by observing real {"reactions": [{"type":"unicode","code":"...","user_ids":[...]}]}
// data in live API responses. Without a body it's the old generic heart
// like. GroupMe restricts the unicode option to a fixed set of 15 emoji
// (see UnicodeLikeIcons in this file); a GroupMe-powerup emoji ({"type":
// "emoji","pack_id":N,"pack_index":N}) is also accepted but not
// implemented here -- no sensible way to map an arbitrary Matrix custom
// emoji to a specific GroupMe powerup pack/index.

/*//////// Endpoints ////////*/
const (
	// Used to build other endpoints
	likesEndpointRoot = "/messages/%s/%s"

	createLikeEndpoint  = likesEndpointRoot + "/like"   // POST
	destroyLikeEndpoint = likesEndpointRoot + "/unlike" // POST
)

// UnicodeLikeIcons is the fixed set of unicode emoji GroupMe accepts for a
// like_icon of type "unicode" (confirmed against the community docs cited
// above). CreateLike with any other emoji falls back to the plain heart
// like (no body) rather than sending a code GroupMe would reject.
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
