// Package groupme defines a client capable of executing API commands for the GroupMe chat service
package groupme

import (
	"context"
	"fmt"
	"net/http"
)

// GroupMe's polls feature (and its API) postdates this library and isn't
// covered by the old dev.groupme.com v3 docs it was written against;
// confirmed instead against real live API responses. A poll is addressed
// by its containing conversation's ID (a group ID in every case observed
// live) plus its own poll ID -- both of which arrive on every poll.*
// message's Event.Data (see json.go's PollEventData), so callers
// shouldn't need to look either up separately.

const pollEndpoint = "/poll/%s/%s" // GET

// PollOption is one selectable choice in a poll, as returned by GetPoll.
// Distinct from PollEventOption (json.go), which is the shape embedded in
// a poll.finished message's Event.Data and carries a final vote tally,
// not a poll definition.
type PollOption struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

// PollDetail is a poll's full definition, as returned by GetPoll. Only
// fields confirmed against a real live poll are documented as such below;
// GroupMe's polls API isn't otherwise documented anywhere.
type PollDetail struct {
	ID             string       `json:"id,omitempty"`
	Subject        string       `json:"subject,omitempty"`
	OwnerID        string       `json:"owner_id,omitempty"`
	ConversationID string       `json:"conversation_id,omitempty"`
	CreatedAt      Timestamp    `json:"created_at,omitempty"`
	Expiration     Timestamp    `json:"expiration,omitempty"`
	// Status is "active" confirmed live; "finished" or similar for an
	// ended poll is expected but not separately confirmed (a finished
	// poll's own GetPoll response wasn't checked -- the poll.finished
	// message's Event.Data already carries the final results directly,
	// so there was no need to).
	Status       string       `json:"status,omitempty"`
	Options      []PollOption `json:"options,omitempty"`
	LastModified Timestamp    `json:"last_modified,omitempty"`
	// Type is "single" confirmed live (single-select); "multiple" for a
	// multi-select poll is expected (GroupMe's UI supports creating one)
	// but not confirmed against a real API response.
	Type string `json:"type,omitempty"`
	// Visibility is "anonymous" confirmed live; a named-voting equivalent
	// is expected but not confirmed.
	Visibility string `json:"visibility,omitempty"`
}

type getPollResponse struct {
	Poll struct {
		Data PollDetail `json:"data"`
	} `json:"poll"`
}

// GetPoll fetches a poll's full definition (subject, options, status,
// expiration, etc). conversationID and pollID both come directly off a
// poll.* message's Event.Data (PollEventData in json.go) -- callers
// shouldn't need to derive either independently.
func (c *Client) GetPoll(ctx context.Context, conversationID, pollID string) (*PollDetail, error) {
	url := fmt.Sprintf(c.endpointBase+pollEndpoint, conversationID, pollID)

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var resp getPollResponse
	if err := c.doWithAuthToken(ctx, httpReq, &resp); err != nil {
		return nil, err
	}
	return &resp.Poll.Data, nil
}
