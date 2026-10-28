package groupme

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCreatePoll(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		status         int
		reason         string
	}{
		{"created", `{"meta":{"code":20100},"response":{"poll":{"data":{"id":"poll-1","conversation_id":"88","options":[{"id":"7","title":"Yes"},{"id":"9","title":"No"}]}},"message":{"id":"announcement-1"}}}`, 201, ""},
		{"invalid expiration", `{"meta":{"code":40002,"errors":["Poll expiration is invalid"]}}`, 400, "Poll expiration is invalid"},
		{"invalid options", `{"meta":{"code":40003,"errors":{"options":"the length must be between 2 and 30"}}}`, 400, "options: the length must be between 2 and 30"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := &CreatePollRequest{Subject: "Choose", Options: []PollOption{{Title: "Yes"}, {Title: "No"}}, Expiration: 1794414187, Visibility: "public", Type: "single"}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/poll/88" || r.Header.Get("X-Access-Token") != "test-token" || r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				var got CreatePollRequest
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil || !reflect.DeepEqual(&got, input) {
					t.Errorf("unexpected request body: %+v, %v", got, err)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.response))
			}))
			defer server.Close()
			client := NewClient("test-token")
			client.endpointBase = server.URL
			poll, messageID, err := client.CreatePoll(context.Background(), "88", input)
			if tc.reason != "" {
				var meta *Meta
				if !errors.As(err, &meta) || meta.Code != 400 || !reflect.DeepEqual(meta.Errors, []string{tc.reason}) {
					t.Fatalf("unexpected rejection: %v", err)
				}
			} else if err != nil || messageID != "announcement-1" || poll.ID != "poll-1" || poll.Options[0].ID != "7" || poll.Options[1].ID != "9" {
				t.Fatalf("unexpected create result: %+v, %s, %v", poll, messageID, err)
			}
		})
	}
}
