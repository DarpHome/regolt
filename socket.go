package regolt

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketDialer interface {
	Dial(wsUrl string, header http.Header) (*websocket.Conn, *http.Response, error)
}

type DefaultWebsocketDialer struct {
	Dialer *websocket.Dialer
}

func NewDefaultWebsocketDialer(dialer *websocket.Dialer) DefaultWebsocketDialer {
	return DefaultWebsocketDialer{Dialer: dialer}
}

func (w DefaultWebsocketDialer) Dial(wsUrl string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return w.Dialer.Dial(wsUrl, header)
}

// The server has authenticated your connection and you will shortly start receiving data.
type Authenticated struct{}

// Data for use by client, data structures match the API specification.
type Ready struct {
	Users    []*User        `json:"users"`
	Servers  []*Server      `json:"servers"`
	Channels []*Channel     `json:"channels"`
	Members  []*Member      `json:"members"`
	Emojis   *[]CustomEmoji `json:"emojis"`
}

type PartialMessage struct {
	// Message content
	Content *string `json:"content"`
	// Attached embeds to this message
	Embeds *[]*Embed `json:"embeds"`
}

// Message edited or otherwise updated.
type MessageUpdate struct {
	MessageID ULID            `json:"id"`
	Channel   ULID            `json:"channel"`
	Data      *PartialMessage `json:"data"`
}

// Message has data being appended to it.
type MessageAppend struct {
	MessageID ULID `json:"id"`
	Channel   ULID `json:"channel"`
	// will contain only `embeds`
	Append *PartialMessage `json:"append"`
}

// Message has been deleted.
type MessageDelete struct {
	MessageID ULID `json:"id"`
	Channel   ULID `json:"channel"`
}

// A reaction has been added to a message.
type MessageReact struct {
	MessageID ULID  `json:"id"`
	ChannelID ULID  `json:"channel_id"`
	UserID    ULID  `json:"user_id"`
	Emoji     Emoji `json:"emoji_id"`
}

// A reaction has been removed from a message.
type MessageUnreact struct {
	MessageID ULID  `json:"id"`
	ChannelID ULID  `json:"channel_id"`
	UserID    ULID  `json:"user_id"`
	Emoji     Emoji `json:"emoji_id"`
}

// A certain reaction has been removed from the message.
type MessageRemoveReaction struct {
	ID        ULID  `json:"id"`
	ChannelID ULID  `json:"channel_id"`
	Emoji     Emoji `json:"emoji_id"`
}

// Multiple messages were deleted.
type BulkDeleteMessage struct {
	// Where messages were deleted.
	ChannelID ULID `json:"channel_id"`
	// Affected message IDs.
	IDs []ULID `json:"ids"`
}

// Channel details updated.
type ChannelUpdate struct {
	ChannelID ULID            `json:"id"`
	Data      *PartialChannel `json:"data"`
	// Possible values: ["Description", "Icon", "DefaultPermissions"]
	Clear []string `json:"clear"`
}

// Whether description was cleared.
func (cu *ChannelUpdate) IsDescriptionCleared() bool {
	return slices.Contains(cu.Clear, "Description")
}

// Whether icon was removed.
func (cu *ChannelUpdate) IsIconRemoved() bool {
	return slices.Contains(cu.Clear, "Icon")
}

// Whether default permissions were removed.
func (cu *ChannelUpdate) IsDefaultPermissionsWereRemoved() bool {
	return slices.Contains(cu.Clear, "DefaultPermissions")
}

// Channel has been deleted.
type ChannelDelete struct {
	ChannelID ULID `json:"id"`
}

// A user has joined the group.
type ChannelGroupJoin struct {
	ChannelID ULID `json:"id"`
	User      ULID `json:"user"`
}

// A user has left the group.
type ChannelGroupLeave struct {
	ChannelID ULID `json:"id"`
	User      ULID `json:"user"`
}

// A user has started typing in this channel.
type ChannelStartTyping struct {
	ChannelID ULID `json:"id"`
	User      ULID `json:"user"`
}

// A user has stopped typing in this channel.
type ChannelStopTyping struct {
	ChannelID ULID `json:"id"`
	User      ULID `json:"user"`
}

// You have acknowledged new messages in this channel up to this message ID.
type ChannelAck struct {
	ChannelID ULID `json:"id"`
	User      ULID `json:"user"`
	MessageID ULID `json:"message_id"`
}

// Server has been created.
type ServerCreate struct {
	// Unique ID
	ID ULID `json:"id"`
	// Server
	Server *Server `json:"server"`
	// Channels within this server
	Channels []*Channel `json:"channels"`
	// Emojis within this server
	Emojis []*CustomEmoji `json:"emojis"`
}

// Server details updated.
type ServerUpdate struct {
	ServerID ULID           `json:"id"`
	Data     *PartialServer `json:"data"`
	// Possible values: ["Description", "Categories", "SystemMessages", "Icon", "Banner"]
	Clear []string `json:"clear"`
}

// Whether description was cleared.
func (su *ServerUpdate) IsDescriptionCleared() bool {
	return slices.Contains(su.Clear, "Description")
}

// Whether categories were removed.
func (su *ServerUpdate) IsCategoriesWereRemoved() bool {
	return slices.Contains(su.Clear, "Categories")
}

// Whether icon was removed.
func (su *ServerUpdate) IsIconRemoved() bool {
	return slices.Contains(su.Clear, "Icon")
}

// Whether banner was removed.
func (su *ServerUpdate) IsBannerRemoved() bool {
	return slices.Contains(su.Clear, "Banner")
}

// Whether system messages were removed.
func (su *ServerUpdate) IsSystemMessagesWereRemoved() bool {
	return slices.Contains(su.Clear, "SystemMessages")
}

// Server has been deleted.
type ServerDelete struct {
	ServerID ULID `json:"id"`
}

// Server member details updated.
type ServerMemberUpdate struct {
	ID   MemberID       `json:"id"`
	Data *PartialMember `json:"data"`
	// Possible values: ["Nickname", "Avatar", "Roles", "Timeout"]
	Clear []string `json:"clear"`
}

// Whether nickname was removed.
func (smu *ServerMemberUpdate) IsNicknameRemoved() bool {
	return slices.Contains(smu.Clear, "Nickname")
}

// Whether avatar was removed.
func (smu *ServerMemberUpdate) IsAvatarRemoved() bool {
	return slices.Contains(smu.Clear, "Avatar")
}

// Whether roles were cleared.
func (smu *ServerMemberUpdate) IsRolesWereCleared() bool {
	return slices.Contains(smu.Clear, "Roles")
}

// Whether timeout was removed.
func (smu *ServerMemberUpdate) IsTimeoutRemoved() bool {
	return slices.Contains(smu.Clear, "Timeout")
}

// A user has joined the server.
type ServerMemberJoin struct {
	ServerID ULID `json:"id"`
	UserID   ULID `json:"user"`
}

// A user has left the server.
type ServerMemberLeave struct {
	ServerID ULID `json:"id"`
	UserID   ULID `json:"user"`
}

// Server role has been updated or created.
type ServerRoleUpdate struct {
	ServerID ULID         `json:"id"`
	RoleID   ULID         `json:"role_id"`
	Data     *PartialRole `json:"data"`
	// Possible values: ["Colour"]
	Clear     []string `json:"clear"`
	IsCreated bool     `json:"-"`
}

// Whether colour was removed.
func (sru *ServerRoleUpdate) IsColourRemoved() bool {
	return slices.Contains(sru.Clear, "Colour")
}

// Server role has been deleted.
type ServerRoleDelete struct {
	ServerID ULID `json:"id"`
	RoleID   ULID `json:"role_id"`
}

// User has been updated.
type UserUpdate struct {
	UserID ULID         `json:"id"`
	Data   *PartialUser `json:"data"`
	// Possible values: ["Avatar", "StatusText", "StatusPresence", "ProfileContent", "ProfileBackground", "DisplayName"]
	Clear   []string `json:"clear"`
	EventID ULID     `json:"event_id"`
}

// Whether avatar was removed.
func (uu *UserUpdate) IsAvatarRemoved() bool {
	return slices.Contains(uu.Clear, "Avatar")
}

// Whether status text was removed.
func (uu *UserUpdate) IsStatusTextRemoved() bool {
	return slices.Contains(uu.Clear, "StatusText")
}

// Whether status presence was removed.
func (uu *UserUpdate) IsStatusPresenceRemoved() bool {
	return slices.Contains(uu.Clear, "StatusPresence")
}

// Whether profile content was removed.
func (uu *UserUpdate) IsProfileContentRemoved() bool {
	return slices.Contains(uu.Clear, "ProfileContent")
}

// Whether profile background was removed.
func (uu *UserUpdate) IsProfileBackgroundRemoved() bool {
	return slices.Contains(uu.Clear, "ProfileBackground")
}

// Whether display name was removed.
func (uu *UserUpdate) IsDisplayNameRemoved() bool {
	return slices.Contains(uu.Clear, "Display")
}

// Your relationship with another user has changed.
type UserRelationship struct {
	ID   ULID  `json:"id"`
	User *User `json:"user"`
	// deprecated
	// Status RelationshipStatus `json:"status"`
}

func (ur *UserRelationship) Status() RelationshipStatus {
	return ur.User.Relationship
}

// Settings were updated remotely
type UserSettingsUpdate struct {
	ID     ULID                `json:"id"`
	Update map[string]*Setting `json:"update"`
}

// User has been platform banned or deleted their account.
// Clients should remove the following associated data:
// - Messages
// - DM Channels
// - Relationships
// - Server Memberships
// User flags are specified to explain why a wipe is occurring though not all reasons will necessarily ever appear.
type UserPlatformWipe struct {
	UserID ULID      `json:"user_id"`
	Flags  UserFlags `json:"flags"`
}

// Emoji has been deleted.
type EmojiDelete struct {
	EmojiID ULID `json:"id"`
}

// Webhook details updated.
type WebhookUpdate struct {
	ID   ULID            `json:"id"`
	Data *PartialWebhook `json:"data"`
	// Possible values: ["Avatar"]
	Remove []string `json:"remove"`
}

// Whether avatar was removed.
func (wu *WebhookUpdate) IsAvatarRemoved() bool {
	return slices.Contains(wu.Remove, "Avatar")
}

// Webhook has been deleted.
type WebhookDelete struct {
	ID ULID `json:"id"`
}

type AuthEventType string

const (
	AuthEventTypeCreateAccount     AuthEventType = "CreateAccount"
	AuthEventTypeCreateSession     AuthEventType = "CreateSession"
	AuthEventTypeDeleteSession     AuthEventType = "DeleteSession"
	AuthEventTypeDeleteAllSessions AuthEventType = "DeleteAllSessions"
)

type EmailVerificationStatus string

const (
	// Account is verified
	EmailVerificationStatusVerified EmailVerificationStatus = "Verified"
	// Pending email verification
	EmailVerificationStatusPending EmailVerificationStatus = "Pending"
	// Moving to a new email
	EmailVerificationStatusMoving EmailVerificationStatus = "Moving"
)

// Email verification status
type EmailVerification struct {
	Status EmailVerificationStatus `json:"status"`
	// [Present only on Moving] New email
	NewEmail string `json:"new_email,omitempty"`
	// [Present only on Pending/Moving] New token
	Token string `json:"token,omitempty"`
	// [Present only on Pending/Moving] Time at which this token expires
	Expiry *Time `json:"expiry,omitempty"`
}

// Password reset information
type PasswordReset struct {
	// Token required to change password
	Token string `json:"token"`
	// Time at which this token expires
	Expiry *Time `json:"expiry,omitempty"`
}

type DeletionInfoStatus string

const (
	// The user must confirm deletion by email
	DeletionInfoStatusWaitingForVerification DeletionInfoStatus = "WaitingForVerification"
	// The account is scheduled for deletion
	DeletionInfoStatusScheduled DeletionInfoStatus = "Scheduled"
	// This account was deleted
	DeletionInfoStatusDeleted DeletionInfoStatus = "Deleted"
)

// Account deletion information
type DeletionInfo struct {
	Token  string `json:"token,omitempty"`
	Expiry *Time  `json:"expiry,omitempty"`
	After  *Time  `json:"after,omitempty"`
}

// Lockout information
type Lockout struct {
	// Attempt counter
	Attempts int `json:"attempts"`
	// Time at which this lockout expires
	Expiry *Time `json:"expiry"`
}

type TOTPStatus string

const (
	// Disabled
	TOTPStatusDisabled TOTPStatus = "Disabled"
	// Waiting for user activation
	TOTPStatusPending TOTPStatus = "Pending"
	// Required on account
	TOTPStatusEnabled TOTPStatus = "Enabled"
)

// Time-based one-time password configuration
type TOTP struct {
	Status TOTPStatus `json:"status"`
	// [Present only on Pending/Enabled]
	Secret string `json:"secret,omitempty"`
}

// MFA configuration
type MultiFactorAuthentication struct {
	// TOTP MFA token, enabled if present
	// (2-Factor)
	TOTPToken *TOTP `json:"totp_token,omitempty"`
	// Recovery codes
	RecoveryCodes []string `json:"recovery_codes,omitempty"`
}

// Account model
type SocketAccount struct {
	// Unique ID
	ID string `json:"_id"`
	// User's email
	Email string `json:"email"`
	// Normalised email
	//
	// (see https://github.com/insertish/authifier/#how-does-authifier-work)
	EmailNormalised string `json:"email_normalised"`
	// Argon2 hashed password
	Password string `json:"password"`
	// Whether the account is disabled
	Disabled bool `json:"disabled"`
	// Email verification status
	Verification *EmailVerification `json:"verification"`
	// Password reset information
	PasswordReset *PasswordReset `json:"password_reset"`
	// Account deletion information
	Deletion *DeletionInfo `json:"deletion"`
	// Account lockout
	Lockout *Lockout `json:"lockout"`
	// Multi-factor authentication information
	MFA *MultiFactorAuthentication `json:"mfa"`
}

type SocketSession struct {
	// Unique ID
	ID string `json:"_id"`
	// User ID
	UserID string `json:"user_id"`
	// Session token
	Token string `json:"token"`
	// Display name
	Name string `json:"name"`
	// Web Push subscription
	Subscription *WebPushSubscription `json:"subscription,omitempty"`
}

// Forwarded events from rAuth, currently only session deletion events are forwarded.
type Auth struct {
	EventType AuthEventType `json:"event_type"`
	// [Present only on CreateAccount]
	Account *SocketAccount `json:"account,omitempty"`
	// [Present only on CreateSession]
	Session *SocketSession `json:"session,omitempty"`
	// [Present only on DeleteSession/DeleteAllSessions]
	UserID ULID `json:"user_id,omitempty"`
	// [Present only on DeleteSession]
	SessionID string `json:"session_id,omitempty"`
	// [Present only on DeleteAllSessions]
	ExcludeSessionID string `json:"exclude_session_id,omitempty"`
}

type Events struct {
	Error         *EventController[error]
	RevoltError   *EventController[string]
	Authenticated *EventController[*Authenticated]
	Raw           *EventController[map[string]any]
	Ready         *EventController[*Ready]
	// Message received, the event object has the same schema as the Message object in the API with the addition of an event type.
	Message               *EventController[*Message]
	MessageUpdate         *EventController[*MessageUpdate]
	MessageAppend         *EventController[*MessageAppend]
	MessageDelete         *EventController[*MessageDelete]
	MessageReact          *EventController[*MessageReact]
	MessageUnreact        *EventController[*MessageUnreact]
	MessageRemoveReaction *EventController[*MessageRemoveReaction]
	BulkDeleteMessage     *EventController[*BulkDeleteMessage]
	// Channel created, the event object has the same schema as the Channel object in the API with the addition of an event type.
	ChannelCreate      *EventController[*Channel]
	ChannelUpdate      *EventController[*ChannelUpdate]
	ChannelDelete      *EventController[*ChannelDelete]
	ChannelGroupJoin   *EventController[*ChannelGroupJoin]
	ChannelGroupLeave  *EventController[*ChannelGroupLeave]
	ChannelStartTyping *EventController[*ChannelStartTyping]
	ChannelStopTyping  *EventController[*ChannelStopTyping]
	ChannelAck         *EventController[*ChannelAck]
	ServerCreate       *EventController[*ServerCreate]
	ServerUpdate       *EventController[*ServerUpdate]
	ServerDelete       *EventController[*ServerDelete]
	ServerMemberUpdate *EventController[*ServerMemberUpdate]
	ServerMemberJoin   *EventController[*ServerMemberJoin]
	ServerMemberLeave  *EventController[*ServerMemberLeave]
	ServerRoleUpdate   *EventController[*ServerRoleUpdate]
	ServerRoleDelete   *EventController[*ServerRoleDelete]
	UserUpdate         *EventController[*UserUpdate]
	UserRelationship   *EventController[*UserRelationship]
	UserSettingsUpdate *EventController[*UserSettingsUpdate]
	UserPlatformWipe   *EventController[*UserPlatformWipe]
	// Emoji created, the event object has the same schema as the Emoji object in the API with the addition of an event type.
	EmojiCreate *EventController[*CustomEmoji]
	EmojiDelete *EventController[*EmojiDelete]
	// Webhook created, the event object has the same schema as the Webhook object in the API with the addition of an event type.
	WebhookCreate *EventController[*Webhook]
	WebhookUpdate *EventController[*WebhookUpdate]
	WebhookDelete *EventController[*WebhookDelete]
	// Report created, the event object has the same schema as the Report object in the API with the addition of an event type.
	ReportCreate *EventController[*Report]
	Auth         *EventController[*Auth]
}

func (e *Events) init() {
	e.Error = NewEventController[error]()
	e.RevoltError = NewEventController[string]()
	e.Authenticated = NewEventController[*Authenticated]()
	e.Raw = NewEventController[map[string]any]()
	e.Ready = NewEventController[*Ready]()
	e.Message = NewEventController[*Message]()
	e.MessageUpdate = NewEventController[*MessageUpdate]()
	e.MessageAppend = NewEventController[*MessageAppend]()
	e.MessageDelete = NewEventController[*MessageDelete]()
	e.MessageReact = NewEventController[*MessageReact]()
	e.MessageUnreact = NewEventController[*MessageUnreact]()
	e.MessageRemoveReaction = NewEventController[*MessageRemoveReaction]()
	e.BulkDeleteMessage = NewEventController[*BulkDeleteMessage]()
	e.ChannelCreate = NewEventController[*Channel]()
	e.ChannelUpdate = NewEventController[*ChannelUpdate]()
	e.ChannelDelete = NewEventController[*ChannelDelete]()
	e.ChannelGroupJoin = NewEventController[*ChannelGroupJoin]()
	e.ChannelGroupLeave = NewEventController[*ChannelGroupLeave]()
	e.ChannelStartTyping = NewEventController[*ChannelStartTyping]()
	e.ChannelStopTyping = NewEventController[*ChannelStopTyping]()
	e.ChannelAck = NewEventController[*ChannelAck]()
	e.ServerCreate = NewEventController[*ServerCreate]()
	e.ServerUpdate = NewEventController[*ServerUpdate]()
	e.ServerDelete = NewEventController[*ServerDelete]()
	e.ServerMemberUpdate = NewEventController[*ServerMemberUpdate]()
	e.ServerMemberJoin = NewEventController[*ServerMemberJoin]()
	e.ServerMemberLeave = NewEventController[*ServerMemberLeave]()
	e.ServerRoleUpdate = NewEventController[*ServerRoleUpdate]()
	e.ServerRoleDelete = NewEventController[*ServerRoleDelete]()
	e.UserUpdate = NewEventController[*UserUpdate]()
	e.UserRelationship = NewEventController[*UserRelationship]()
	e.UserSettingsUpdate = NewEventController[*UserSettingsUpdate]()
	e.UserPlatformWipe = NewEventController[*UserPlatformWipe]()
	e.EmojiCreate = NewEventController[*CustomEmoji]()
	e.EmojiDelete = NewEventController[*EmojiDelete]()
	e.WebhookCreate = NewEventController[*Webhook]()
	e.WebhookUpdate = NewEventController[*WebhookUpdate]()
	e.WebhookDelete = NewEventController[*WebhookDelete]()
	e.ReportCreate = NewEventController[*Report]()
	e.Auth = NewEventController[*Auth]()
}

type Socket struct {
	Token       string
	Dialer      WebsocketDialer
	Connection  *websocket.Conn
	URL         *url.URL
	Me          *User
	Events      Events
	closed      bool
	lastPingReq time.Time
	lastPingRes time.Time
	ticker      *time.Ticker
	closeEvent  chan struct{}
	mu          sync.Mutex
	Cache       *GenericCache
	Logger      *slog.Logger
	Arshaler    JSONArshaler
}

func (socket *Socket) Latency() time.Duration {
	if socket.lastPingReq.IsZero() || socket.lastPingRes.IsZero() {
		return 0
	}
	return socket.lastPingRes.Sub(socket.lastPingReq)
}

func (socket *Socket) LatencyMs() float64 {
	l := socket.Latency()
	if l == 0 {
		return math.NaN()
	}
	return float64(l/time.Millisecond) / 1000
}

func (socket *Socket) Close() error {
	return socket.CloseWithCode(websocket.CloseNormalClosure)
}

func (socket *Socket) close() error {
	socket.closed = true
	socket.closeEvent <- struct{}{}
	socket.ticker.Stop()
	return nil
}

func (socket *Socket) CloseWithCode(closeCode int) error {
	if err := socket.close(); err != nil {
		return err
	}
	return socket.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(closeCode, ""))
}

// Generate by:
// typ = 'type Events struct { ... }'
// print('\n'.join(map(lambda p: "func (socket *Socket) On{0}(f func({1})) *Subscription[{1}] {\n\treturn socket.Events.{0}.Listen(f)\n}\n".replace('{0}', p[0]).replace('{1}', p[1]), re.compile(r'(\w+)\s+\*EventController\[([0-9A-Za-z_\[\]]+)\]', re.IGNORECASE).findall(typ))))

func (socket *Socket) OnError(f func(error)) *Subscription[error] {
	return socket.Events.Error.Listen(f)
}

func (socket *Socket) OnRevoltError(f func(string)) *Subscription[string] {
	return socket.Events.RevoltError.Listen(f)
}

func (socket *Socket) OnAuthenticated(f func(*Authenticated)) *Subscription[*Authenticated] {
	return socket.Events.Authenticated.Listen(f)
}

func (socket *Socket) OnRaw(f func(map[string]any)) *Subscription[map[string]any] {
	return socket.Events.Raw.Listen(f)
}

func (socket *Socket) OnReady(f func(*Ready)) *Subscription[*Ready] {
	return socket.Events.Ready.Listen(f)
}

func (socket *Socket) OnMessage(f func(*Message)) *Subscription[*Message] {
	return socket.Events.Message.Listen(f)
}

func (socket *Socket) OnMessageUpdate(f func(*MessageUpdate)) *Subscription[*MessageUpdate] {
	return socket.Events.MessageUpdate.Listen(f)
}

func (socket *Socket) OnMessageAppend(f func(*MessageAppend)) *Subscription[*MessageAppend] {
	return socket.Events.MessageAppend.Listen(f)
}

func (socket *Socket) OnMessageDelete(f func(*MessageDelete)) *Subscription[*MessageDelete] {
	return socket.Events.MessageDelete.Listen(f)
}

func (socket *Socket) OnMessageReact(f func(*MessageReact)) *Subscription[*MessageReact] {
	return socket.Events.MessageReact.Listen(f)
}

func (socket *Socket) OnMessageUnreact(f func(*MessageUnreact)) *Subscription[*MessageUnreact] {
	return socket.Events.MessageUnreact.Listen(f)
}

func (socket *Socket) OnMessageRemoveReaction(f func(*MessageRemoveReaction)) *Subscription[*MessageRemoveReaction] {
	return socket.Events.MessageRemoveReaction.Listen(f)
}

func (socket *Socket) OnBulkDeleteMessage(f func(*BulkDeleteMessage)) *Subscription[*BulkDeleteMessage] {
	return socket.Events.BulkDeleteMessage.Listen(f)
}

func (socket *Socket) OnChannelCreate(f func(*Channel)) *Subscription[*Channel] {
	return socket.Events.ChannelCreate.Listen(f)
}

func (socket *Socket) OnChannelUpdate(f func(*ChannelUpdate)) *Subscription[*ChannelUpdate] {
	return socket.Events.ChannelUpdate.Listen(f)
}

func (socket *Socket) OnChannelDelete(f func(*ChannelDelete)) *Subscription[*ChannelDelete] {
	return socket.Events.ChannelDelete.Listen(f)
}

func (socket *Socket) OnChannelGroupJoin(f func(*ChannelGroupJoin)) *Subscription[*ChannelGroupJoin] {
	return socket.Events.ChannelGroupJoin.Listen(f)
}

func (socket *Socket) OnChannelGroupLeave(f func(*ChannelGroupLeave)) *Subscription[*ChannelGroupLeave] {
	return socket.Events.ChannelGroupLeave.Listen(f)
}

func (socket *Socket) OnChannelStartTyping(f func(*ChannelStartTyping)) *Subscription[*ChannelStartTyping] {
	return socket.Events.ChannelStartTyping.Listen(f)
}

func (socket *Socket) OnChannelStopTyping(f func(*ChannelStopTyping)) *Subscription[*ChannelStopTyping] {
	return socket.Events.ChannelStopTyping.Listen(f)
}

func (socket *Socket) OnChannelAck(f func(*ChannelAck)) *Subscription[*ChannelAck] {
	return socket.Events.ChannelAck.Listen(f)
}

func (socket *Socket) OnServerCreate(f func(*ServerCreate)) *Subscription[*ServerCreate] {
	return socket.Events.ServerCreate.Listen(f)
}

func (socket *Socket) OnServerUpdate(f func(*ServerUpdate)) *Subscription[*ServerUpdate] {
	return socket.Events.ServerUpdate.Listen(f)
}

func (socket *Socket) OnServerDelete(f func(*ServerDelete)) *Subscription[*ServerDelete] {
	return socket.Events.ServerDelete.Listen(f)
}

func (socket *Socket) OnServerMemberUpdate(f func(*ServerMemberUpdate)) *Subscription[*ServerMemberUpdate] {
	return socket.Events.ServerMemberUpdate.Listen(f)
}

func (socket *Socket) OnServerMemberJoin(f func(*ServerMemberJoin)) *Subscription[*ServerMemberJoin] {
	return socket.Events.ServerMemberJoin.Listen(f)
}

func (socket *Socket) OnServerMemberLeave(f func(*ServerMemberLeave)) *Subscription[*ServerMemberLeave] {
	return socket.Events.ServerMemberLeave.Listen(f)
}

func (socket *Socket) OnServerRoleUpdate(f func(*ServerRoleUpdate)) *Subscription[*ServerRoleUpdate] {
	return socket.Events.ServerRoleUpdate.Listen(f)
}

func (socket *Socket) OnServerRoleDelete(f func(*ServerRoleDelete)) *Subscription[*ServerRoleDelete] {
	return socket.Events.ServerRoleDelete.Listen(f)
}

func (socket *Socket) OnUserUpdate(f func(*UserUpdate)) *Subscription[*UserUpdate] {
	return socket.Events.UserUpdate.Listen(f)
}

func (socket *Socket) OnUserRelationship(f func(*UserRelationship)) *Subscription[*UserRelationship] {
	return socket.Events.UserRelationship.Listen(f)
}

func (socket *Socket) OnUserSettingsUpdate(f func(*UserSettingsUpdate)) *Subscription[*UserSettingsUpdate] {
	return socket.Events.UserSettingsUpdate.Listen(f)
}

func (socket *Socket) OnUserPlatformWipe(f func(*UserPlatformWipe)) *Subscription[*UserPlatformWipe] {
	return socket.Events.UserPlatformWipe.Listen(f)
}

func (socket *Socket) OnEmojiCreate(f func(*CustomEmoji)) *Subscription[*CustomEmoji] {
	return socket.Events.EmojiCreate.Listen(f)
}

func (socket *Socket) OnEmojiDelete(f func(*EmojiDelete)) *Subscription[*EmojiDelete] {
	return socket.Events.EmojiDelete.Listen(f)
}

func (socket *Socket) OnWebhookCreate(f func(*Webhook)) *Subscription[*Webhook] {
	return socket.Events.WebhookCreate.Listen(f)
}

func (socket *Socket) OnWebhookUpdate(f func(*WebhookUpdate)) *Subscription[*WebhookUpdate] {
	return socket.Events.WebhookUpdate.Listen(f)
}

func (socket *Socket) OnWebhookDelete(f func(*WebhookDelete)) *Subscription[*WebhookDelete] {
	return socket.Events.WebhookDelete.Listen(f)
}

func (socket *Socket) OnReportCreate(f func(*Report)) *Subscription[*Report] {
	return socket.Events.ReportCreate.Listen(f)
}

func (socket *Socket) OnAuth(f func(*Auth)) *Subscription[*Auth] {
	return socket.Events.Auth.Listen(f)
}

func (socket *Socket) init() {
	socket.Events.init()
	socket.Cache.init()
}

type SocketConfig struct {
	Cache          *GenericCache
	Dialer         WebsocketDialer
	URL            *url.URL
	AddLogSource   bool
	Logger         *slog.Logger
	DisableLogging bool
	LoggerLevel    slog.Leveler
	Arshaler       JSONArshaler
}

func NewSocket(token string, config *SocketConfig) (socket *Socket, err error) {
	if config == nil {
		config = &SocketConfig{}
	}
	d := config.Dialer
	if d == nil {
		d = NewDefaultWebsocketDialer(&websocket.Dialer{})
	}
	wsUrl := config.URL
	if wsUrl == nil {
		wsUrl, err = url.Parse("wss://ws.revolt.chat/?version=1&format=json")
		if err != nil {
			return
		}
	}
	cache := config.Cache
	if cache == nil {
		cache = &GenericCache{}
	}
	logger := config.Logger
	if logger == nil && !config.DisableLogging {
		level := config.LoggerLevel
		if level == nil {
			level = slog.LevelInfo
		}
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.AddLogSource,
			Level:     level,
		}).WithGroup("regolt"))
	}
	arshaler := config.Arshaler
	if arshaler == nil {
		arshaler = NewJSONArshaler(json.Marshal, json.Unmarshal)
	}
	socket = &Socket{
		Token:      token,
		Dialer:     d,
		closeEvent: make(chan struct{}),
		URL:        wsUrl,
		Cache:      cache,
		Logger:     logger,
		Arshaler:   arshaler,
	}
	socket.init()
	return
}

func (s *Socket) logDebug(msg string, args ...any) {
	if s.Logger != nil {
		s.Logger.Debug(msg, args...)
	}
}

func (s *Socket) logInfo(msg string, args ...any) {
	if s.Logger != nil {
		s.Logger.Info(msg, args...)
	}
}

func (s *Socket) logError(msg string, args ...any) {
	if s.Logger != nil {
		s.Logger.Error(msg, args...)
	}
}

func (s *Socket) logWarn(msg string, args ...any) {
	if s.Logger != nil {
		s.Logger.Warn(msg, args...)
	}
}

type SocketError struct {
	ErrorID string
}

func (se SocketError) Error() string {
	t := ""
	switch se.ErrorID {
	case "LabelMe":
		t = "uncategorised error"
	case "InternalServer":
		t = "the server ran into an issue"
	case "InvalidSession":
		t = "authentication details are incorrect"
	case "OnboardingNotFinished":
		t = "user has not chosen a username"
	case "AlreadyAuthenticated":
		t = "this connection is already authenticated"
	}
	if len(t) != 0 {
		return se.ErrorID + ": " + t
	}
	return se.ErrorID
}

func (socket *Socket) Connect() error {
	conn, _, err := socket.Dialer.Dial(socket.URL.String(), http.Header{})
	socket.logInfo("connected to websocket")
	if err != nil {
		return err
	}
	conn.SetCloseHandler(func(code int, message string) error {
		socket.logDebug("received close message with code/message", slog.Int("code", code), slog.String("message", message))
		if socket.closed {
			socket.logDebug("socket was closed, returning")
			return nil
		}
		socket.logDebug("closing socket internals")
		if err := socket.close(); err != nil {
			socket.logError("caught an error when closing", slog.Any("err", err))
			return err
		}
		// reinitialize
		socket.lastPingReq = time.Time{}
		socket.lastPingRes = time.Time{}
		socket.logDebug("reopening")
		return socket.Open()
	})
	socket.Connection = conn
	return nil
}

func (socket *Socket) Write(typ string, d map[string]any) error {
	e := d
	e["type"] = typ
	socket.mu.Lock()
	defer socket.mu.Unlock()
	b, err := socket.marshal(e)
	if err != nil {
		return err
	}
	return socket.Connection.WriteMessage(websocket.TextMessage, b)
}

func (socket *Socket) Authenticate() error {
	return socket.Write("Authenticate", map[string]any{"token": socket.Token})
}

func (socket *Socket) BeginTyping(channel ULID) error {
	return socket.Write("BeginTyping", map[string]any{"channel": channel})
}

func (socket *Socket) EndTyping(channel ULID) error {
	return socket.Write("EndTyping", map[string]any{"channel": channel})
}

func (socket *Socket) Ping() error {
	socket.lastPingReq = time.Now()
	return socket.Write("Ping", map[string]any{"data": 0})
}

func (socket *Socket) emitError(err error) {
	socket.logError("caught an error", slog.Any("err", err))
	socket.Events.Error.EmitInGoroutines(err)
}

func (socket *Socket) pinger() {
	for {
		select {
		case <-socket.closeEvent:
			return
		case _, d := <-socket.ticker.C:
			if !d {
				return
			}
		}
		if socket.lastPingRes.Before(socket.lastPingReq) {
			return
		}
		if err := socket.Ping(); err != nil {
			socket.Events.Error.Emit(err)
			return
		}
	}
}

func (socket *Socket) Open() (err error) {
	err = socket.Connect()
	if err != nil {
		return
	}
	readyEvent := make(chan struct{})
	errorEvent := make(chan error)
	sub1 := socket.Events.Ready.Listen(func(_ *Ready) {
		readyEvent <- struct{}{}
	})
	defer sub1.Delete()
	sub2 := socket.Events.RevoltError.Listen(func(errorID string) {
		errorEvent <- &SocketError{ErrorID: errorID}
	})
	defer sub2.Delete()
	sub3 := socket.Events.Error.Listen(func(err error) {
		errorEvent <- err
	})
	defer sub3.Delete()
	socket.logInfo("authenticating")
	err = socket.Authenticate()
	if err != nil {
		return
	}
	socket.Listen()
	select {
	case err = <-errorEvent:
		return
	case <-readyEvent:
		socket.logInfo("ready")
	}
	return
}

func (socket *Socket) marshal(v any) ([]byte, error) {
	if socket.Arshaler != nil {
		b, err := socket.Arshaler.Marshal(v)
		if _, ok := err.(*ArshalNotImplemented); !ok {
			return b, err
		}
	}
	return json.Marshal(v)
}

func (socket *Socket) unmarshal(d []byte, v any) error {
	if socket.Arshaler != nil {
		err := socket.Arshaler.Unmarshal(d, v)
		if _, ok := err.(*ArshalNotImplemented); !ok {
			return err
		}
	}
	return json.Unmarshal(d, v)
}

func (socket *Socket) process(s []byte) {
	socket.logDebug("processing", slog.String("payload", string(s)))
	var a map[string]any
	if err := json.Unmarshal(s, &a); err != nil {
		return
	}
	socket.Events.Raw.EmitInGoroutines(a)
	typ := a["type"].(string)
	socket.logDebug("received", slog.String("type", typ))
	switch typ {
	case "Error":
		socket.Events.RevoltError.Emit(a["error"].(string))
	case "NotFound":
		socket.Events.RevoltError.Emit("InvalidSession")
	case "Authenticated":
		socket.Events.Authenticated.EmitInGoroutines(&Authenticated{})
	case "Bulk":
		t := struct {
			V []json.RawMessage `json:"v"`
		}{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		for _, u := range t.V {
			socket.process(u)
		}
	case "Pong":
		t := struct {
			Data int `json:"data"`
		}{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.lastPingRes = time.Now()
	case "Ready":
		t := &Ready{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.Ready.EmitAndCall(t, func(r *Ready) {
			for _, u := range r.Users {
				socket.Cache.Users.Set(u.ToOptimized())
			}
			for _, s := range r.Servers {
				socket.Cache.Servers.Set(s.ToOptimized())
				for i, r := range s.Roles {
					socket.Cache.Roles.Set(s.ID, r.ToOptimized(i))
				}
			}
			for _, c := range r.Channels {
				socket.Cache.Channels.Set(c.ToOptimized())
			}
			if r.Emojis != nil {
				for _, e := range *r.Emojis {
					socket.Cache.Emojis.Set(e.ToOptimized())
				}
			}
		})
	case "Message":
		t := &Message{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.Message.EmitAndCall(t, func(r *Message) {
			socket.Cache.Messages.Set(r.Channel, r.ToOptimized())
		})
	case "MessageUpdate":
		t := &MessageUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageUpdate.EmitAndCall(t, func(r *MessageUpdate) {
			socket.Cache.Messages.PartiallyUpdate(r.Channel, r.MessageID, func(m *OptimizedMessage) {
				if r.Data.Content != nil {
					m.Content = *r.Data.Content
				}
				if r.Data.Embeds != nil {
					var embeds []*OptimizedEmbed
					for _, e := range *r.Data.Embeds {
						embeds = append(embeds, e.ToOptimized())
					}
					m.Embeds = embeds
				}
			})
		})
	case "MessageAppend":
		t := &MessageAppend{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageAppend.EmitAndCall(t, func(r *MessageAppend) {
			socket.Cache.Messages.PartiallyUpdate(r.Channel, r.MessageID, func(m *OptimizedMessage) {
				if r.Append.Embeds != nil {
					var embeds []*OptimizedEmbed
					for _, e := range *r.Append.Embeds {
						embeds = append(embeds, e.ToOptimized())
					}
					m.Embeds = append(m.Embeds, embeds...)
				}
			})
		})
	case "MessageDelete":
		t := &MessageDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageDelete.EmitAndCall(t, func(r *MessageDelete) {
			socket.Cache.Messages.Del(r.Channel, r.MessageID)
		})
	case "MessageReact":
		t := &MessageReact{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageReact.EmitInGoroutines(t)
	case "MessageUnreact":
		t := &MessageUnreact{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageUnreact.EmitInGoroutines(t)
	case "MessageRemoveReaction":
		t := &MessageRemoveReaction{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.MessageRemoveReaction.EmitInGoroutines(t)
	case "BulkDeleteMessage":
		t := &BulkDeleteMessage{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.BulkDeleteMessage.EmitAndCall(t, func(r *BulkDeleteMessage) {
			for _, i := range r.IDs {
				socket.Cache.Messages.Del(r.ChannelID, i)
			}
		})
	case "ChannelCreate":
		t := &Channel{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelCreate.EmitAndCall(t, func(r *Channel) {
			socket.Cache.Channels.Set(r.ToOptimized())
		})
	case "ChannelUpdate":
		t := &ChannelUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelUpdate.EmitAndCall(t, func(r *ChannelUpdate) {
			socket.Cache.Channels.PartiallyUpdate(t.ChannelID, func(c *OptimizedChannel) {
				if len(r.Data.Name) != 0 {
					c.Name = r.Data.Name
				}
				if r.Data.Description != nil {
					c.Description = *r.Data.Description
				}
				if r.Data.Active != nil {
					if *r.Data.Active {
						c.Flags |= OptimizedChannelFlagActive
					} else {
						c.Flags &= ^OptimizedChannelFlagActive
					}
				}
				if r.Data.NSFW != nil {
					if *r.Data.NSFW {
						c.Flags |= OptimizedChannelFlagNSFW
					} else {
						c.Flags &= ^OptimizedChannelFlagNSFW
					}
				}
				if r.Data.DefaultPermissions != nil {
					c.DefaultPermissions = r.Data.DefaultPermissions
				}
				if r.Data.RolePermissions != nil {
					c.RolePermissions = *r.Data.RolePermissions
				}
				if len(r.Data.Owner) != 0 {
					c.Owner = r.Data.Owner
				}
				if r.IsDescriptionCleared() {
					c.Description = ""
				}
				if r.IsIconRemoved() {
					c.Icon = nil
				}
				if r.IsDefaultPermissionsWereRemoved() {
					c.DefaultPermissions = nil
				}
			})
		})
	case "ChannelDelete":
		t := &ChannelDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelDelete.EmitAndCall(t, func(r *ChannelDelete) {
			socket.Cache.Channels.Del(r.ChannelID)
		})
	case "ChannelGroupJoin":
		t := &ChannelGroupJoin{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelGroupJoin.EmitInGoroutines(t)
	case "ChannelGroupLeave":
		t := &ChannelGroupLeave{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelGroupLeave.EmitInGoroutines(t)
	case "ChannelStartTyping":
		t := &ChannelStartTyping{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelStartTyping.EmitInGoroutines(t)
	case "ChannelStopTyping":
		t := &ChannelStopTyping{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelStopTyping.EmitInGoroutines(t)
	case "ChannelAck":
		t := &ChannelAck{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ChannelAck.EmitInGoroutines(t)
	case "ServerCreate":
		t := &ServerCreate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerCreate.EmitAndCall(t, func(r *ServerCreate) {
			socket.Cache.Servers.Set(r.Server.ToOptimized())
			for i, o := range r.Server.Roles {
				socket.Cache.Roles.Set(i, o.ToOptimized(i))
			}
			for _, c := range r.Channels {
				socket.Cache.Channels.Set(c.ToOptimized())
			}
		})
	case "ServerUpdate":
		t := &ServerUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerUpdate.EmitAndCall(t, func(r *ServerUpdate) {
			socket.Cache.Servers.PartiallyUpdate(r.ServerID, func(s *OptimizedServer) {
				if r.Data.Name != nil {
					s.Name = *r.Data.Name
				}
				if r.Data.Analytics != nil {
					if *r.Data.Analytics {
						s.Flags |= OptimizedServerFlagsAnalytics
					} else {
						s.Flags &= ^OptimizedServerFlagsAnalytics
					}
				}
				if r.Data.NSFW != nil {
					if *r.Data.NSFW {
						s.Flags |= OptimizedServerFlagsNSFW
					} else {
						s.Flags &= ^OptimizedServerFlagsNSFW
					}
				}
				if r.IsDescriptionCleared() {
					s.Description = ""
				}
				if r.IsCategoriesWereRemoved() {
					s.Categories = []*Category{}
				}
				if r.IsSystemMessagesWereRemoved() {
					s.SystemMessages = nil
				}
				if r.IsIconRemoved() {
					s.Icon = nil
				}
				if r.IsBannerRemoved() {
					s.Banner = nil
				}
			})
		})
	case "ServerDelete":
		t := &ServerDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerDelete.EmitAndCall(t, func(r *ServerDelete) {
			socket.Cache.Roles.DelGroup(r.ServerID)
			socket.Cache.Servers.Del(r.ServerID)
		})
	case "ServerMemberUpdate":
		t := &ServerMemberUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerMemberUpdate.EmitInGoroutines(t)
	case "ServerMemberJoin":
		t := &ServerMemberJoin{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerMemberJoin.EmitAndCall(t, func(smj *ServerMemberJoin) {
			socket.Cache.Members.Set(smj.ServerID, &Member{
				ID: MemberID{
					Server: smj.ServerID,
					User:   smj.UserID,
				},
				JoinedAt: Time(time.Now()),
			})
		})
	case "ServerMemberLeave":
		t := &ServerMemberLeave{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerMemberLeave.EmitInGoroutines(t)
	case "ServerRoleUpdate":
		t := &ServerRoleUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		if len(t.Data.Name) != 0 && t.Data.Permissions != nil && t.Data.Hoist != nil && t.Data.Rank != nil {
			t.IsCreated = true
		}
		socket.Events.ServerRoleUpdate.EmitAndCall(t, func(r *ServerRoleUpdate) {
			if r.IsCreated {
				u := &Role{
					Name:        r.Data.Name,
					Permissions: *r.Data.Permissions,
					Hoist:       *r.Data.Hoist,
					Rank:        *r.Data.Rank,
				}
				if len(r.Data.Colour) != 0 {
					u.Colour = r.Data.Colour
				}
				socket.Cache.Roles.Set(r.ServerID, u.ToOptimized(r.RoleID))
			} else {
				socket.Cache.Roles.PartiallyUpdate(r.ServerID, r.RoleID, func(o *OptimizedRole) {
					if len(r.Data.Name) != 0 {
						o.Name = r.Data.Name
					}
					if r.Data.Permissions != nil {
						o.Permissions = *r.Data.Permissions
					}
					if len(r.Data.Colour) != 0 {
						o.Colour = r.Data.Colour
					}
					if r.Data.Hoist != nil {
						if *r.Data.Hoist {
							o.Flags |= OptimizedRoleFlagsHoist
						} else {
							o.Flags &= ^OptimizedRoleFlagsHoist
						}
					}
					if r.Data.Rank != nil {
						o.Rank = *r.Data.Rank
					}
					if r.IsColourRemoved() {
						o.Colour = ""
					}
				})
			}
		})
	case "ServerRoleDelete":
		t := &ServerRoleDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ServerRoleDelete.EmitAndCall(t, func(r *ServerRoleDelete) {
			socket.Cache.Roles.Del(r.ServerID, r.RoleID)
		})
	case "UserUpdate":
		t := &UserUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.UserUpdate.EmitAndCall(t, func(r *UserUpdate) {
			socket.Cache.Users.PartiallyUpdate(r.UserID, func(u *OptimizedUser) {
				if r.Data.Badges != nil {
					u.Flags.updateBadges(*r.Data.Badges)
				}
				if r.Data.Flags != nil {
					u.Flags.updateFlags(*r.Data.Flags)
				}
				if r.Data.Online != nil {
					if *r.Data.Online {
						u.Flags |= OptimizedUserFlagsOnline
					} else {
						u.Flags &= ^OptimizedUserFlagsOnline
					}
				}
				if r.Data.Username != nil {
					u.Username = *r.Data.Username
				}
				if r.Data.Status != nil {
					if u.Status == nil {
						u.Status = &OptimizedUserStatus{}
					}
					if len(r.Data.Status.Presence) != 0 {
						u.Status.Presence = r.Data.Status.Presence.ToOptimized()
					}
					if len(r.Data.Status.Text) != 0 {
						u.Status.Text = r.Data.Status.Text
					}
				}
				if r.Data.Profile != nil {
					if u.Profile == nil {
						u.Profile = &OptimizedUserProfile{}
					}
					if r.Data.Profile.Background != nil {
						u.Profile.Background = r.Data.Profile.Background.ToOptimized()
					}
				}
				if r.Data.Relations != nil {
					u.Relations = *r.Data.Relations
				}
				if r.Data.Relationship != nil {
					u.Relationship = (*r.Data.Relationship).ToOptimized()
				}
				if r.IsAvatarRemoved() {
					u.Avatar = nil
				}
				if r.IsStatusTextRemoved() && u.Status != nil {
					u.Status.Text = ""
				}
				if r.IsStatusPresenceRemoved() && u.Status != nil {
					u.Status.Presence = OptimizedPresenceInvisible
				}

				if r.IsProfileContentRemoved() && u.Profile != nil {
					u.Profile.Content = ""
				}
				if r.IsProfileBackgroundRemoved() && u.Profile != nil {
					u.Profile.Background = nil
				}
				if r.IsDisplayNameRemoved() {
					u.DisplayName = ""
				}
			})
		})
	case "UserRelationship":
		t := &UserRelationship{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.UserRelationship.EmitAndCall(t, func(r *UserRelationship) {
			socket.Cache.Users.PartiallyUpdate(t.ID, func(u *OptimizedUser) {
				u.Relationship = r.Status().ToOptimized()
			})
		})
	case "UserSettingsUpdate":
		t := &UserSettingsUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.UserSettingsUpdate.EmitInGoroutines(t)
	case "UserPlatformWipe":
		t := &UserPlatformWipe{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.UserPlatformWipe.EmitInGoroutines(t)
	case "EmojiCreate":
		t := &CustomEmoji{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.EmojiCreate.EmitAndCall(t, func(r *CustomEmoji) {
			socket.Cache.Emojis.Set(r.ToOptimized())
		})
	case "EmojiDelete":
		t := &EmojiDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.EmojiDelete.EmitAndCall(t, func(r *EmojiDelete) {
			socket.Cache.Emojis.Del(r.EmojiID)
		})
	case "WebhookCreate":
		t := &Webhook{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.WebhookCreate.EmitAndCall(t, func(r *Webhook) {
			socket.Cache.Webhooks.Set(r.ToOptimized())
		})
	case "WebhookUpdate":
		t := &WebhookUpdate{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.WebhookUpdate.EmitAndCall(t, func(r *WebhookUpdate) {
			socket.Cache.Webhooks.PartiallyUpdate(r.ID, func(w *OptimizedWebhook) {
				if len(r.Data.Name) != 0 {
					w.Name = r.Data.Name
				}
				if r.Data.Avatar != nil {
					w.Avatar = r.Data.Avatar.ToOptimized()
				}
				if r.Data.Permissions != nil {
					w.Permissions = *r.Data.Permissions
				}
				if r.IsAvatarRemoved() {
					w.Avatar = nil
				}
			})
		})
	case "WebhookDelete":
		t := &WebhookDelete{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.WebhookDelete.EmitAndCall(t, func(r *WebhookDelete) {
			socket.Cache.Webhooks.Del(r.ID)
		})
	case "ReportCreate":
		t := &Report{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.ReportCreate.EmitInGoroutines(t)
	case "Auth":
		t := &Auth{}
		if err := socket.unmarshal(s, &t); err != nil {
			socket.emitError(err)
			return
		}
		socket.Events.Auth.EmitInGoroutines(t)
	default:
		socket.logWarn("unknown event received", slog.String("type", typ), slog.String("payload", string(s)))
	}
}

func (socket *Socket) listener() {
	for {
		if socket.closed {
			return
		}
		m, p, err := socket.Connection.ReadMessage()
		if err != nil {
			socket.logError("caught an error when reading message, attempting reconnect in 3s", slog.Any("err", err))
			socket.Events.Error.Emit(err)
			time.Sleep(3 * time.Second)
			socket.logInfo("reconnecting")
			if err = socket.Open(); err != nil {
				socket.Events.Error.Emit(err)
			}
			return
		}
		if m != websocket.TextMessage {
			continue
		}
		socket.process(p)
	}
}

func (socket *Socket) Listen() {
	socket.ticker = time.NewTicker(30 * time.Second)
	go socket.pinger()
	go socket.listener()
}
