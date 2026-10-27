// Package groupme defines a client capable of executing API commands for the GroupMe chat service
package groupme

import (
	"encoding/json"
	"fmt"
)

// Meta is the error type returned in the GroupMe response.
// Meant for clients that can't read HTTP status codes
type Meta struct {
	Code   HTTPStatusCode `json:"code,omitempty"`
	Errors []string       `json:"errors,omitempty"`
}

// Error returns the code and the error list as a string.
// Satisfies the error interface
func (m Meta) Error() string {
	return fmt.Sprintf("Error Code %d: %v", m.Code, m.Errors)
}

// Group is a GroupMe group, returned in JSON API responses
type Group struct {
	ID   ID     `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Type of group (private|public)
	Type          string        `json:"type,omitempty"`
	Description   string        `json:"description,omitempty"`
	ImageURL      string        `json:"image_url,omitempty"`
	CreatorUserID ID            `json:"creator_user_id,omitempty"`
	CreatedAt     Timestamp     `json:"created_at,omitempty"`
	UpdatedAt     Timestamp     `json:"updated_at,omitempty"`
	Members       []*Member     `json:"members,omitempty"`
	ShareURL      string        `json:"share_url,omitempty"`
	Messages      GroupMessages `json:"messages,omitempty"`
}

// GroupMessages is a Group field, only returned in Group JSON API responses
type GroupMessages struct {
	Count                uint           `json:"count,omitempty"`
	LastMessageID        ID             `json:"last_message_id,omitempty"`
	LastMessageCreatedAt Timestamp      `json:"last_message_created_at,omitempty"`
	Preview              MessagePreview `json:"preview,omitempty"`
}

// MessagePreview is a GroupMessages field, only returned in Group JSON API responses.
// Abbreviated form of Message type
type MessagePreview struct {
	Nickname    string        `json:"nickname,omitempty"`
	Text        string        `json:"text,omitempty"`
	ImageURL    string        `json:"image_url,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

// GetMemberByUserID gets the group member by their UserID,
// nil if no member matches
func (g *Group) GetMemberByUserID(userID ID) *Member {
	for _, member := range g.Members {
		if member.UserID == userID {
			return member
		}
	}

	return nil
}

// GetMemberByNickname gets the group member by their Nickname,
// nil if no member matches
func (g *Group) GetMemberByNickname(nickname string) *Member {
	for _, member := range g.Members {
		if member.Nickname == nickname {
			return member
		}
	}

	return nil
}

func (g *Group) String() string {
	return marshal(g)
}

// Member is a GroupMe group member, returned in JSON API responses
type Member struct {
	ID           ID     `json:"id,omitempty"`
	UserID       ID     `json:"user_id,omitempty"`
	Nickname     string `json:"nickname,omitempty"`
	Name         string `json:"name,omitempty"` // Account-wide name; Nickname is per-group.
	Muted        bool   `json:"muted,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	AutoKicked   bool   `json:"autokicked,omitempty"`
	AppInstalled bool   `json:"app_installed,omitempty"`
	GUID         string `json:"guid,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"` // Only used when searching for the member to add to a group.
	Email        string `json:"email,omitempty"`        // Only used when searching for the member to add to a group.
}

func (m *Member) String() string {
	return marshal(m)
}

// Message is a GroupMe group message, returned in JSON API responses
type Message struct {
	ID          ID         `json:"id,omitempty"`
	SourceGUID  string     `json:"source_guid,omitempty"`
	CreatedAt   Timestamp  `json:"created_at,omitempty"`
	GroupID     ID         `json:"group_id,omitempty"`
	UserID      ID         `json:"user_id,omitempty"`
	BotID       ID         `json:"bot_id,omitempty"`
	SenderID    ID         `json:"sender_id,omitempty"`
	SenderType  senderType `json:"sender_type,omitempty"`
	System      bool       `json:"system,omitempty"`
	Name        string     `json:"name,omitempty"`
	RecipientID ID         `json:"recipient_id,omitempty"`
	//ChatID - over push ConversationID seems to be called ChatID
	ChatID         ID     `json:"chat_id,omitempty"`
	ConversationID ID     `json:"conversation_id,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	// Maximum length of 1000 characters
	Text string `json:"text,omitempty"`
	// Must be an image service URL (i.groupme.com)
	ImageURL    string        `json:"image_url,omitempty"`
	FavoritedBy []string      `json:"favorited_by,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
	// Reactions holds per-emoji reactions; FavoritedBy only records that a
	// user reacted, not which emoji they used.
	Reactions []Reaction `json:"reactions,omitempty"`
	Event     *Event     `json:"event,omitempty"`
}

// Reaction is one emoji and every user who reacted with it.
type Reaction struct {
	Type    string   `json:"type,omitempty"`
	Code    string   `json:"code,omitempty"`
	UserIDs []string `json:"user_ids,omitempty"`
}

// Event is the structured payload of a system message. Data's shape depends
// on Type; poll.* events decode into PollEventData.
type Event struct {
	Type string          `json:"type,omitempty"`
	Data json.RawMessage `json:"data,omitempty"`
}

type PollEventEntity struct {
	ID       string `json:"id,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

// PollEventPoll is a poll summary. Expiration is only set on poll.reminder;
// use GetPoll for the full definition.
type PollEventPoll struct {
	ID         string    `json:"id,omitempty"`
	Subject    string    `json:"subject,omitempty"`
	Expiration Timestamp `json:"expiration,omitempty"`
}

// PollEventOption is one option's final tally from poll.finished. VoterIDs is
// only expected on non-anonymous polls.
type PollEventOption struct {
	ID       string   `json:"id,omitempty"`
	Title    string   `json:"title,omitempty"`
	Votes    int      `json:"votes,omitempty"`
	VoterIDs []string `json:"voter_ids,omitempty"`
}

// PollEventData is a poll.* Event's Data: poll.created sets Poll and User,
// poll.reminder sets Poll, and poll.finished sets Poll and Options.
type PollEventData struct {
	Conversation struct {
		ID string `json:"id,omitempty"`
	} `json:"conversation,omitempty"`
	Poll    PollEventPoll     `json:"poll,omitempty"`
	User    PollEventEntity   `json:"user,omitempty"`
	Options []PollEventOption `json:"options,omitempty"`
}

func (m *Message) String() string {
	return marshal(m)
}

type senderType string

// SenderType constants
const (
	SenderTypeUser   senderType = "user"
	SenderTypeBot    senderType = "bot"
	SenderTypeSystem senderType = "system"
)

type attachmentType string

// AttachmentType constants
const (
	Mentions attachmentType = "mentions"
	Image    attachmentType = "image"
	Location attachmentType = "location"
	Emoji    attachmentType = "emoji"
	Video    attachmentType = "video"
	Audio    attachmentType = "audio"
	File     attachmentType = "file"
	Reply    attachmentType = "reply"
	// Poll only carries PollID; the details are in the message's poll.created Event.
	Poll attachmentType = "poll"
)

// Attachment is a GroupMe message attachment, returned in JSON API responses
type Attachment struct {
	Type            attachmentType `json:"type,omitempty"`
	Loci            [][]int        `json:"loci,omitempty"`
	UserIDs         []ID           `json:"user_ids,omitempty"`
	URL             string         `json:"url,omitempty"`
	FileID          string         `json:"file_id,omitempty"`
	VideoPreviewURL string         `json:"preview_url,omitempty"`
	Name            string         `json:"name,omitempty"`
	Latitude        string         `json:"lat,omitempty"`
	Longitude       string         `json:"lng,omitempty"`
	Placeholder     string         `json:"placeholder,omitempty"`
	Charmap         [][]int        `json:"charmap,omitempty"`
	ReplyID         ID             `json:"reply_id,omitempty"`
	BaseReplyID     ID             `json:"base_reply_id,omitempty"`
	UserID          ID             `json:"user_id,omitempty"`
	PollID          string         `json:"poll_id,omitempty"`

	// Duration is in seconds. Peaks holds normalized waveform samples, which
	// GroupMe clients send either as a JSON array or as a string containing one.
	Duration      float64         `json:"duration,omitempty"`
	Peaks         json.RawMessage `json:"peaks,omitempty"`
	TranscriptURL string          `json:"transcript_url,omitempty"`
}

func (a *Attachment) String() string {
	return marshal(a)
}

// User is a GroupMe user, returned in JSON API responses
type User struct {
	ID          ID          `json:"id,omitempty"`
	PhoneNumber PhoneNumber `json:"phone_number,omitempty"`
	ImageURL    string      `json:"image_url,omitempty"`
	Name        string      `json:"name,omitempty"`
	CreatedAt   Timestamp   `json:"created_at,omitempty"`
	UpdatedAt   Timestamp   `json:"updated_at,omitempty"`
	AvatarURL   string      `json:"avatar_url,omitempty"`
	Email       string      `json:"email,omitempty"`
	SMS         bool        `json:"sms,omitempty"`
}

func (u *User) String() string {
	return marshal(u)
}

// Chat is a GroupMe direct message conversation between two users,
// returned in JSON API responses
type Chat struct {
	CreatedAt     Timestamp `json:"created_at,omitempty"`
	UpdatedAt     Timestamp `json:"updated_at,omitempty"`
	LastMessage   *Message  `json:"last_message,omitempty"`
	MessagesCount int       `json:"messages_count,omitempty"`
	OtherUser     User      `json:"other_user,omitempty"`
}

func (c *Chat) String() string {
	return marshal(c)
}

// Bot is a GroupMe bot, it is connected to a specific group which it can send messages to
type Bot struct {
	BotID          ID     `json:"bot_id,omitempty"`
	GroupID        ID     `json:"group_id,omitempty"`
	Name           string `json:"name,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	CallbackURL    string `json:"callback_url,omitempty"`
	DMNotification bool   `json:"dm_notification,omitempty"`
}

func (b *Bot) String() string {
	return marshal(b)
}

// Block is a GroupMe block between two users, direct messages are not allowed
type Block struct {
	UserID        ID        `json:"user_id,omitempty"`
	BlockedUserID ID        `json:"blocked_user_id,omitempty"`
	CreatedAT     Timestamp `json:"created_at,omitempty"`
}

func (b Block) String() string {
	return marshal(&b)
}

// Superficially increases test coverage
func marshal(i interface{}) string {
	bytes, err := json.MarshalIndent(i, "", "\t")
	if err != nil {
		return ""
	}

	return string(bytes)
}
