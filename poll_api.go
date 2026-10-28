// Package groupme defines a client capable of executing API commands for the GroupMe chat service
package groupme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Polls are not in the v3 docs; see
// https://groupme-js.github.io/GroupMeCommunityDocs/api/groups/polls/.
// A poll is addressed by its conversation ID and poll ID, both of which are
// carried on every poll.* message's PollEventData.

const pollEndpoint = "/poll/%s/%s" // GET

type PollOption struct {
	ID    string `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
}

type PollDetail struct {
	ID             string       `json:"id,omitempty"`
	Subject        string       `json:"subject,omitempty"`
	OwnerID        string       `json:"owner_id,omitempty"`
	ConversationID string       `json:"conversation_id,omitempty"`
	CreatedAt      Timestamp    `json:"created_at,omitempty"`
	Expiration     Timestamp    `json:"expiration,omitempty"`
	Status         string       `json:"status,omitempty"`
	Options        []PollOption `json:"options,omitempty"`
	LastModified   Timestamp    `json:"last_modified,omitempty"`
	// Type is "single" or "multi".
	Type string `json:"type,omitempty"`
	// Visibility is "public" or "anonymous".
	Visibility string `json:"visibility,omitempty"`
	// MyVotes is the authenticated user's selection from the poll envelope.
	MyVotes []string `json:"-"`
}

type getPollResponse struct {
	Poll struct {
		Data      PollDetail `json:"data"`
		UserVote  string     `json:"user_vote"`
		UserVotes []string   `json:"user_votes"`
	} `json:"poll"`
}

type CreatePollRequest struct {
	Subject    string       `json:"subject"`
	Options    []PollOption `json:"options"`
	Expiration Timestamp    `json:"expiration"`
	Visibility string       `json:"visibility"`
	Type       string       `json:"type"`
}

// CreatePoll returns the poll and the ID of its poll.created announcement.
func (c *Client) CreatePoll(ctx context.Context, conversationID string, poll *CreatePollRequest) (*PollDetail, ID, error) {
	body, err := json.Marshal(poll)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointBase+"/poll/"+url.PathEscape(conversationID), bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	var resp struct {
		getPollResponse
		Message Message `json:"message"`
	}
	if err := c.doWithAuthToken(ctx, req, &resp); err != nil {
		return nil, "", err
	}
	if resp.Poll.Data.ID == "" || resp.Message.ID == "" {
		return nil, "", fmt.Errorf("GroupMe poll creation response has a missing poll or message ID")
	}
	result, err := resp.result(conversationID, resp.Poll.Data.ID)
	return result, resp.Message.ID, err
}

func (c *Client) GetPoll(ctx context.Context, conversationID, pollID string) (*PollDetail, error) {
	url := fmt.Sprintf(c.endpointBase+pollEndpoint, url.PathEscape(conversationID), url.PathEscape(pollID))

	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var resp getPollResponse
	if err := c.doWithAuthToken(ctx, httpReq, &resp); err != nil {
		return nil, err
	}
	return resp.result(conversationID, pollID)
}

func (resp *getPollResponse) result(conversationID, pollID string) (*PollDetail, error) {
	if resp.Poll.Data.ID != pollID || resp.Poll.Data.ConversationID != conversationID {
		return nil, fmt.Errorf("GroupMe poll response has a missing or mismatched identity")
	}
	resp.Poll.Data.MyVotes = resp.Poll.UserVotes
	if resp.Poll.Data.MyVotes == nil && resp.Poll.UserVote != "" {
		resp.Poll.Data.MyVotes = []string{resp.Poll.UserVote}
	}
	return &resp.Poll.Data, nil
}

// VotePoll replaces the current user's selection. Single-choice polls put the
// option ID in the path; multi-choice polls accept the complete votes array.
// The multi-choice endpoint has no trailing slash, as used by GroupMe Web.
// The response is the updated poll and user selection, not a chat message.
func (c *Client) VotePoll(ctx context.Context, conversationID, pollID string, votes []string, multiple bool) (*PollDetail, error) {
	endpoint := fmt.Sprintf(c.endpointBase+pollEndpoint, url.PathEscape(conversationID), url.PathEscape(pollID))
	var body []byte
	if multiple {
		var err error
		body, err = json.Marshal(struct {
			Votes []string `json:"votes"`
		}{Votes: votes})
		if err != nil {
			return nil, err
		}
	} else {
		if len(votes) != 1 || votes[0] == "" {
			return nil, fmt.Errorf("a single-choice GroupMe poll requires one vote")
		}
		endpoint += "/" + url.PathEscape(votes[0])
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if multiple {
		req.Header.Set("Content-Type", "application/json")
	}
	var resp getPollResponse
	if err := c.doWithAuthToken(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.result(conversationID, pollID)
}
