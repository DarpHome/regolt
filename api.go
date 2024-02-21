package regolt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Route struct {
	Method string
	Path   string
}

var (
	// API.QueryNode() -> GET /
	RouteQueryNode = func() Route { return Route{"GET", "/"} }
	// API.FetchSelf() -> GET /users/@me
	RouteFetchSelf = func() Route { return Route{"GET", "/users/@me"} }
	// API.FetchUser(user) -> GET /users/{user}
	RouteFetchUser = func(user ULID) Route { return Route{"GET", "/users/" + user.EncodeFP()} }
	// API.EditUser(user, params) -> PATCH /users/{user}
	RouteEditUser = func(user ULID) Route { return Route{"PATCH", "/users/" + user.EncodeFP()} }
	// API.FetchUserFlags(user) -> GET /users/{user}/flags
	RouteFetchUserFlags = func(user ULID) Route { return Route{"GET", "/users/" + user.EncodeFP() + "/flags"} }
	// API.ChangeUsername(params) -> PATCH /users/@me/username
	RouteChangeUsername = func() Route { return Route{"PATCH", "/users/@me/username"} }
	// API.FetchDefaultAvatar(user) -> GET /users/{user}/default_avatar
	RouteFetchDefaultAvatar = func(user ULID) Route {
		return Route{"GET", "/users/" + user.EncodeFP() + "/default_avatar"}
	}
	// API.FetchUserProfile(user) -> GET /users/{user}/profile
	RouteFetchUserProfile = func(user ULID) Route {
		return Route{"GET", "/users/" + user.EncodeFP() + "/profile"}
	}
	// API.FetchDirectMessageChannels() -> GET /users/dms
	RouteFetchDirectMessageChannels = func() Route {
		return Route{"GET", "/users/dms"}
	}
	// API.OpenDirectMessage(user) -> GET /users/{user}/dm
	RouteOpenDirectMessage = func(user ULID) Route {
		return Route{"GET", "/users/" + user.EncodeFP() + "/dm"}
	}
	// API.FetchMutualFriendsAndServers(user) -> GET /users/{user}/mutual
	RouteFetchMutualFriendsAndServers = func(user ULID) Route {
		return Route{"GET", "/users/" + user.EncodeFP() + "/mutual"}
	}
	// API.AcceptFriendRequest(user) -> PUT /users/{user}/friend
	RouteAcceptFriendRequest = func(user ULID) Route {
		return Route{"PUT", "/users/" + user.EncodeFP() + "/friend"}
	}
	// API.DenyFriendRequest(user) -> DELETE /users/{user}/friend
	RouteDenyFriendRequest = func(user ULID) Route {
		return Route{"DELETE", "/users/" + user.EncodeFP() + "/friend"}
	}
	// API.RemoveFriend(user) -> DELETE /users/{user}/friend
	RouteRemoveFriend = func(user ULID) Route {
		return Route{"DELETE", "/users/" + user.EncodeFP() + "/friend"}
	}
	// API.BlockUser(user) -> PUT /users/{user}/block
	RouteBlockUser = func(user ULID) Route {
		return Route{"PUT", "/users/" + user.EncodeFP() + "/block"}
	}
	// API.UnblockUser(user) -> DELETE /users/{user}/block
	RouteUnblockUser = func(user ULID) Route {
		return Route{"DELETE", "/users/" + user.EncodeFP() + "/block"}
	}
	// API.SendFriendRequest(username) -> POST /users/friend
	RouteSendFriendRequest = func() Route {
		return Route{"POST", "/users/friend"}
	}
	// API.CreateBot(name) -> POST /bots/create
	RouteCreateBot = func() Route {
		return Route{"POST", "/bots/create"}
	}
	// API.FetchPublicBot(bot) -> GET /bots/{bot}/invite
	RouteFetchPublicBot = func(bot ULID) Route {
		return Route{"GET", "/bots/" + bot.EncodeFP() + "/invite"}
	}
	// API.InviteBot(bot, params) -> POST /bots/{bot}/invite
	RouteInviteBot = func(bot ULID) Route {
		return Route{"POST", "/bots/" + bot.EncodeFP() + "/invite"}
	}
	// API.FetchBot(bot) -> GET /bots/{bot}
	RouteFetchBot = func(bot ULID) Route {
		return Route{"GET", "/bots/" + bot.EncodeFP()}
	}
	// API.FetchOwnedBots() -> GET /bots/@me
	RouteFetchOwnedBots = func() Route {
		return Route{"GET", "/bots/@me"}
	}
	// API.DeleteBot(bot) -> DELETE /bots/{bot}
	RouteDeleteBot = func(bot ULID) Route {
		return Route{"DELETE", "/bots/" + bot.EncodeFP()}
	}
	// API.EditBot(bot, params) -> PATCH /bots/{bot}
	RouteEditBot = func(bot ULID) Route {
		return Route{"PATCH", "/bots/" + bot.EncodeFP()}
	}
	// API.FetchChannel(channel) -> GET /channels/{channel}
	RouteFetchChannel = func(channel ULID) Route {
		return Route{"GET", "/channels/" + channel.EncodeFP()}
	}
	// API.CloseChannel(channel) -> DELETE /channels/{channel}
	RouteCloseChannel = func(channel ULID) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP()}
	}
	// API.EditChannel(channel, params) -> PATCH /channels/{channel}
	RouteEditChannel = func(channel ULID) Route {
		return Route{"PATCH", "/channels/" + channel.EncodeFP()}
	}
	// API.CreateInvite(channel) -> POST /channels/{channel}/invites
	RouteCreateInvite = func(channel ULID) Route {
		return Route{"POST", "/channels/" + channel.EncodeFP() + "/invites"}
	}
	// API.SetRoleChannelPermission(channel, role, allow, deny) -> PUT /channels/{channel}/permissions/{role}
	RouteSetRoleChannelPermission = func(channel, role ULID) Route {
		return Route{"PUT", "/channels/" + channel.EncodeFP() + "/permissions/" + role.EncodeFP()}
	}
	// API.SetDefaultChannelPermission(channel, allow, deny) -> PUT /channels/{channel}/permissions/default
	RouteSetDefaultChannelPermission = func(channel ULID) Route {
		return Route{"PUT", "/channels/" + channel.EncodeFP() + "/permissions/default"}
	}
	// API.AcknowledgeMessage(channel, message) -> PUT /channels/{channel}/ack/{message}
	RouteAcknowledgeMessage = func(channel, message ULID) Route {
		return Route{"PUT", "/channels/" + channel.EncodeFP() + "/ack/" + message.EncodeFP()}
	}
	// API.FetchMessages(channel, params) -> GET /channels/{channel}/messages
	RouteFetchMessages = func(channel ULID) Route {
		return Route{"GET", "/channels/" + channel.EncodeFP() + "/messages"}
	}
	// API.SendMessage(channel, params) -> POST /channels/{channel}/messages
	RouteSendMessage = func(channel ULID) Route {
		return Route{"POST", "/channels/" + channel.EncodeFP() + "/messages"}
	}
	// API.SearchForMessages(channel) -> POST /channels/{channel}/search
	RouteSearchForMessages = func(channel ULID) Route {
		return Route{"POST", "/channels/" + channel.EncodeFP() + "/search"}
	}
	// API.FetchMessage(channel, message) -> GET /channels/{channel}/messages/{message}
	RouteFetchMessage = func(channel, message ULID) Route {
		return Route{"GET", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP()}
	}
	// API.DeleteMessage(channel, message) -> DELETE /channels/{channel}/messages/{message}
	RouteDeleteMessage = func(channel, message ULID) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP()}
	}
	// API.EditMessage(channel, message, params) -> PATCH /channels/{channel}/messages/{message}
	RouteEditMessage = func(channel, message ULID) Route {
		return Route{"PATCH", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP()}
	}
	// API.BulkDeleteMessages(channel, ids) -> DELETE /channels/{channel}/messages/bulk
	RouteBulkDeleteMessages = func(channel ULID) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP() + "/messages/bulk"}
	}
	// API.AddReactionToMessage(channel, message, emoji) -> PUT /channels/{channel}/messages/{message}/reactions/{emoji}
	RouteAddReactionToMessage = func(channel, message ULID, emoji Emoji) Route {
		return Route{"PUT", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP() + "/reactions/" + emoji.EncodeFP()}
	}
	// API.RemoveReactionsFromMessage(channel, message, emoji) -> DELETE /channels/{channel}/messages/{message}/reactions/{emoji}
	RouteRemoveReactionsFromMessage = func(channel, message ULID, emoji Emoji) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP() + "/reactions/" + emoji.EncodeFP()}
	}
	// API.RemoveAllReactionsFromMessage(channel, message) -> DELETE /channels/{channel}/messages/{message}/reactions
	RouteRemoveAllReactionsFromMessage = func(channel, message ULID) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP() + "/messages/" + message.EncodeFP() + "/reactions"}
	}
	// API.FetchGroupMembers(channel) -> GET /channels/{channel}/members
	RouteFetchGroupMembers = func(channel ULID) Route {
		return Route{"GET", "/channels/" + channel.EncodeFP() + "/members"}
	}
	// API.CreateGroup(params) -> POST /channels/create
	RouteCreateGroup = func() Route {
		return Route{"POST", "/channels/create"}
	}
	// API.AddMemberToGroup(channel, member) -> PUT /channels/{channel}/recipients/{member}
	RouteAddMemberToGroup = func(channel, member ULID) Route {
		return Route{"PUT", "/channels/" + channel.EncodeFP() + "/recipients/" + member.EncodeFP()}
	}
	// API.RemoveMemberFromGroup(channel, member) -> DELETE /channels/{channel}/recipients/{member}
	RouteRemoveMemberFromGroup = func(channel, member ULID) Route {
		return Route{"DELETE", "/channels/" + channel.EncodeFP() + "/recipients/" + member.EncodeFP()}
	}
	// API.JoinCall(channel) -> POST /channels/{channel}/join_call
	RouteJoinCall = func(channel ULID) Route {
		return Route{"POST", "/channels/" + channel.EncodeFP() + "/join_call"}
	}
	// API.CreateWebhook(channel, name, avatar) -> POST /channels/{channel}/webhooks
	RouteCreateWebhook = func(channel ULID) Route {
		return Route{"POST", "/channels/" + channel.EncodeFP() + "/webhooks"}
	}
	// API.DeleteWebhook(webhook, "") -> DELETE /webhooks/{webhook}
	RouteDeleteWebhook = func(webhook ULID) Route {
		return Route{"DELETE", "/webhooks/" + webhook.EncodeFP()}
	}
	// API.DeleteWebhook(webhook, token) -> DELETE /webhooks/{webhook}/{token}
	RouteDeleteWebhookWithToken = func(webhook ULID, token string) Route {
		return Route{"DELETE", "/webhooks/" + webhook.EncodeFP() + "/" + url.PathEscape(token)}
	}
	// API.EditWebhook(webhook, "", params) -> PATCH /webhooks/{webhook}
	RouteEditWebhook = func(webhook ULID) Route {
		return Route{"PATCH", "/webhooks/" + webhook.EncodeFP()}
	}
	// API.EditWebhook(webhook, token, params) -> PATCH /webhooks/{webhook}/{token}
	RouteEditWebhookWithToken = func(webhook ULID, token string) Route {
		return Route{"PATCH", "/webhooks/" + webhook.EncodeFP() + "/" + url.PathEscape(token)}
	}
	// API.ExecuteWebhook(webhook, token, params) -> POST /webhooks/{webhook}/{token}
	RouteExecuteWebhook = func(webhook ULID, token string) Route {
		return Route{"POST", "/webhooks/" + webhook.EncodeFP() + "/" + url.PathEscape(token)}
	}
	// API.FetchWebhook(webhook, "") -> GET /webhooks/{webhook}
	RouteFetchWebhook = func(webhook ULID) Route {
		return Route{"GET", "/webhooks/" + webhook.EncodeFP()}
	}
	// API.FetchWebhook(webhook, token) -> GET /webhooks/{webhook}/{token}
	RouteFetchWebhookWithToken = func(webhook ULID, token string) Route {
		return Route{"GET", "/webhooks/" + webhook.EncodeFP() + "/" + url.PathEscape(token)}
	}
	// API.FetchChannelWebhooks(channel) -> GET /channels/{channel}/webhooks
	RouteFetchChannelWebhooks = func(channel ULID) Route {
		return Route{"GET", "/channels/" + channel.EncodeFP() + "/webhooks"}
	}
	// API.CreateServer(params) -> POST /servers/create
	RouteCreateServer = func() Route {
		return Route{"POST", "/servers/create"}
	}
	// API.FetchServer(server) -> GET /servers/{server}
	RouteFetchServer = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP()}
	}
	// API.DeleteServer(server) -> DELETE /servers/{server}
	RouteDeleteServer = func(server ULID) Route {
		return Route{"DELETE", "/servers/" + server.EncodeFP()}
	}
	// API.LeaveServer(server) -> DELETE /servers/{server}
	RouteLeaveServer = func(server ULID) Route {
		return Route{"DELETE", "/servers/" + server.EncodeFP()}
	}
	// API.EditServer(server, params) -> PATCH /servers/{server}
	RouteEditServer = func(server ULID) Route {
		return Route{"PATCH", "/servers/" + server.EncodeFP()}
	}
	// API.MarkServerAsRead(server) -> PUT /servers/{server}/ack
	RouteMarkServerAsRead = func(server ULID) Route {
		return Route{"PUT", "/servers/" + server.EncodeFP() + "/ack"}
	}
	// API.CreateChannel(server, params) -> POST /servers/{server}/channels
	RouteCreateChannel = func(server ULID) Route {
		return Route{"POST", "/servers/" + server.EncodeFP() + "/channels"}
	}
	// API.FetchMembers(server, params) -> GET /servers/{server}/members
	RouteFetchMembers = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/members"}
	}
	// API.FetchMember(server, member) -> GET /servers/{server}/members/{member}
	RouteFetchMember = func(server ULID, member ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/members/" + member.EncodeFP()}
	}
	// API.KickMember(server, member) -> DELETE /servers/{server}/members/{member}
	RouteKickMember = func(server ULID, member ULID) Route {
		return Route{"DELETE", "/servers/" + server.EncodeFP() + "/members/" + member.EncodeFP()}
	}
	// API.EditMember(server, member, params) -> PATCH /servers/{server}/members/{member}
	RouteEditMember = func(server ULID, member ULID) Route {
		return Route{"PATCH", "/servers/" + server.EncodeFP() + "/members/" + member.EncodeFP()}
	}
	// API.QueryMembersByName(server, query) -> GET /servers/{server}/members_experimental_query
	RouteQueryMembersByName = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/members_experimental_query"}
	}
	// API.BanUser(server, user, reason) -> PUT /servers/{server}/bans/{user}
	RouteBanUser = func(server ULID, user ULID) Route {
		return Route{"PUT", "/servers/" + server.EncodeFP() + "/bans/" + user.EncodeFP()}
	}
	// API.UnbanUser(server, user) -> DELETE /servers/{server}/bans/{user}
	RouteUnbanUser = func(server ULID, user ULID) Route {
		return Route{"DELETE", "/servers/" + server.EncodeFP() + "/bans/" + user.EncodeFP()}
	}
	// API.FetchBans(server) -> GET /servers/{server}/bans
	RouteFetchBans = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/bans"}
	}
	// API.FetchInvites(server) -> GET /servers/{server}/invites
	RouteFetchInvites = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/invites"}
	}
	// API.CreateRole(server, params) -> POST /servers/{server}/roles
	RouteCreateRole = func(server ULID) Route {
		return Route{"POST", "/servers/" + server.EncodeFP() + "/roles"}
	}
	// API.DeleteRole(server, role) -> DELETE /servers/{server}/roles/{role}
	RouteDeleteRole = func(server ULID, role ULID) Route {
		return Route{"DELETE", "/servers/" + server.EncodeFP() + "/roles/" + role.EncodeFP()}
	}
	// API.EditRole(server, role) -> PATCH /servers/{server}/roles/{role}
	RouteEditRole = func(server ULID, role ULID) Route {
		return Route{"PATCH", "/servers/" + server.EncodeFP() + "/roles/" + role.EncodeFP()}
	}
	// API.SetRoleServerPermission(server, role, allow, deny) -> PUT /servers/{server}/permissions/{role}
	RouteSetRoleServerPermission = func(server ULID, role ULID) Route {
		return Route{"PUT", "/servers/" + server.EncodeFP() + "/permissions/" + role.EncodeFP()}
	}
	// API.SetDefaultServerPermission(server,, allow, deny) -> PUT /servers/{server}/permissions/default
	RouteSetDefaultServerPermission = func(server ULID) Route {
		return Route{"PUT", "/servers/" + server.EncodeFP() + "/permissions/default"}
	}
	// API.FetchInvite(invite) -> GET /invites/{invite}
	RouteFetchInvite = func(invite string) Route {
		return Route{"GET", "/invites/" + url.PathEscape(invite)}
	}
	// API.JoinInvite(invite) -> POST /invites/{invite}
	RouteJoinInvite = func(invite string) Route {
		return Route{"POST", "/invites/" + url.PathEscape(invite)}
	}
	// API.DeleteInvite(invite) -> DELETE /invites/{invite}
	RouteDeleteInvite = func(invite string) Route {
		return Route{"DELETE", "/invites/" + url.PathEscape(invite)}
	}
	// API.FetchEmoji(emoji) -> GET /custom/emoji/{emoji}
	RouteFetchEmoji = func(emoji ULID) Route {
		return Route{"GET", "/custom/emoji/" + emoji.EncodeFP()}
	}
	// API.FetchEmoji(id) -> PUT /custom/emoji/{id}
	RouteCreateEmoji = func(id string) Route {
		return Route{"PUT", "/custom/emoji/" + url.PathEscape(id)}
	}
	// API.DeleteEmoji(emoji) -> DELETE /custom/emoji/{emoji}
	RouteDeleteEmoji = func(emoji ULID) Route {
		return Route{"DELETE", "/custom/emoji/" + emoji.EncodeFP()}
	}
	// API.FetchServerEmojis(server) -> GET /servers/{server}/emojis
	RouteFetchServerEmojis = func(server ULID) Route {
		return Route{"GET", "/servers/" + server.EncodeFP() + "/emojis"}
	}
	RouteQueryStats = func() Route {
		return Route{"GET", "/admin/stats"}
	}
	RouteGloballyFetchMessages = func() Route {
		return Route{"POST", "/admin/messages"}
	}
	// API.EditReport(report, params) -> PATCH /safety/reports/{report}
	RouteEditReport = func(report ULID) Route {
		return Route{"PATCH", "/safety/reports/" + report.EncodeFP()}
	}
	// API.FetchReport(report) -> GET /safety/reports/{report}
	RouteFetchReport = func(report ULID) Route {
		return Route{"GET", "/safety/reports/" + report.EncodeFP()}
	}
	// API.FetchReports() -> GET /safety/reports
	RouteFetchReports = func() Route {
		return Route{"GET", "/safety/reports"}
	}
	// API.ReportContent(content, additionalContext) -> POST /safety/report
	RouteReportContent = func() Route {
		return Route{"POST", "/safety/report"}
	}
	// API.FetchSnapshots(report) -> GET /safety/snapshot/{report}
	RouteFetchSnapshots = func(report ULID) Route {
		return Route{"GET", "/safety/snapshot/" + report.EncodeFP()}
	}
	// API.CreateStrike(user, reason) -> POST /safety/strikes
	RouteCreateStrike = func() Route {
		return Route{"POST", "/safety/strikes"}
	}
	// API.FetchStrikes(user) -> GET /safety/strikes/{user}
	RouteFetchStrikes = func(user ULID) Route {
		return Route{"GET", "/safety/strikes/" + user.EncodeFP()}
	}
	// API.EditStrike(strike, params) -> POST /safety/strikes/{strike}
	RouteEditStrike = func(strike ULID) Route {
		return Route{"POST", "/safety/strikes/" + strike.EncodeFP()}
	}
	// API.DeleteStrike(strike) -> DELETE /safety/strikes/{strike}
	RouteDeleteStrike = func(strike ULID) Route {
		return Route{"DELETE", "/safety/strikes/" + strike.EncodeFP()}
	}
	// API.CreateAccount(params) -> POST /auth/account/create
	RouteCreateAccount = func() Route {
		return Route{"POST", "/auth/account/create"}
	}
	// API.ResendVerification(email, captcha) -> POST /auth/account/reverify
	RouteResendVerification = func() Route {
		return Route{"POST", "/auth/account/reverify"}
	}
	// API.ConfirmAccountDeletion(token) -> PUT /auth/account/delete
	RouteConfirmAccountDeletion = func() Route {
		return Route{"PUT", "/auth/account/delete"}
	}
	// API.DeleteAccount() -> POST /auth/account/delete
	RouteDeleteAccount = func() Route {
		return Route{"POST", "/auth/account/delete"}
	}
	// API.FetchAccount() -> GET /auth/account
	RouteFetchAccount = func() Route {
		return Route{"GET", "/auth/account/"}
	}
	// API.DisableAccount() -> POST /auth/account/disable
	RouteDisableAccount = func() Route {
		return Route{"POST", "/auth/account/disable"}
	}
	// API.ChangePassword(password, currentPassword) -> PATCH /auth/account/change/password
	RouteChangePassword = func() Route {
		return Route{"PATCH", "/auth/account/change/password"}
	}
	// API.ChangeEmail(email, currentPassword) -> PATCH /auth/account/change/email
	RouteChangeEmail = func() Route {
		return Route{"PATCH", "/auth/account/change/email"}
	}
	// API.VerifyEmail(code) -> POST /auth/account/verify/{code}
	RouteVerifyEmail = func(code string) Route {
		return Route{"POST", "/auth/account/verify/" + url.PathEscape(code)}
	}
	// API.SendPasswordReset(email, captcha) -> POST /auth/account/reset_password
	RouteSendPasswordReset = func() Route {
		return Route{"POST", "/auth/account/reset_password"}
	}
	// API.PasswordReset(token, password, removeSessions) -> PATCH /auth/account/reset_password
	RoutePasswordReset = func() Route {
		return Route{"PATCH", "/auth/account/reset_password"}
	}
	// API.Login(params) -> POST /auth/session/login
	RouteLogin = func() Route {
		return Route{"POST", "/auth/session/login"}
	}
	// API.Logout() -> POST /auth/session/logout
	RouteLogout = func() Route {
		return Route{"POST", "/auth/session/logout"}
	}
	// API.FetchSessions() -> GET /auth/session/all
	RouteFetchSessions = func() Route {
		return Route{"GET", "/auth/session/all"}
	}
	// API.DeleteAllSessions(revokeSelf) -> DELETE /auth/session/all
	RouteDeleteAllSessions = func() Route {
		return Route{"DELETE", "/auth/session/all"}
	}
	// API.RevokeSession(session) -> DELETE /auth/session/{id}
	RouteRevokeSession = func(id string) Route {
		return Route{"DELETE", "/auth/session/" + url.PathEscape(id)}
	}
	// API.EditSession(session, friendlyName) -> PATCH /auth/session/{id}
	RouteEditSession = func(id string) Route {
		return Route{"PATCH", "/auth/session/" + url.PathEscape(id)}
	}
	// API.CheckOnboardingStatus() -> GET /onboard/hello
	RouteCheckOnboardingStatus = func() Route {
		return Route{"GET", "/onboard/hello"}
	}
	// API.CompleteOnboarding(username) -> POST /onboard/complete
	RouteCompleteOnboarding = func() Route {
		return Route{"POST", "/onboard/complete"}
	}
	// API.CreateMFATicket(params) -> PUT /auth/mfa/ticket
	RouteCreateMFATicket = func() Route {
		return Route{"PUT", "/auth/mfa/ticket"}
	}
	// API.FetchMFAStatus() -> GET /auth/mfa/
	RouteFetchMFAStatus = func() Route {
		return Route{"GET", "/auth/mfa/"}
	}
	// API.FetchRecoveryCodes() -> POST /auth/mfa/recovery
	RouteFetchRecoveryCodes = func() Route {
		return Route{"POST", "/auth/mfa/recovery"}
	}
	// `API.GenerateRecoveryCodes()` -> `DELETE /auth/mfa/recovery`
	RouteGenerateRecoveryCodes = func() Route {
		return Route{"PATCH", "/auth/mfa/recovery"}
	}
	// API.GetMFAMethods() -> GET /auth/mfa/methods
	RouteGetMFAMethods = func() Route {
		return Route{"GET", "/auth/mfa/methods"}
	}
	// API.EnableTOTP2FA(params) -> PUT /auth/mfa/totp
	RouteEnableTOTP2FA = func() Route {
		return Route{"PUT", "/auth/mfa/totp"}
	}
	// API.GenerateTOTPSecre() -> POST /auth/mfa/totp
	RouteGenerateTOTPSecret = func() Route {
		return Route{"POST", "/auth/mfa/totp"}
	}
	// API.DisableTOTP2FA() -> DELETE /auth/mfa/totp
	RouteDisableTOTP2FA = func() Route {
		return Route{"DELETE", "/auth/mfa/totp"}
	}
	RouteFetchSettings = func() Route {
		return Route{"POST", "/sync/settings/fetch"}
	}

	RouteSetSettings = func() Route {
		return Route{"POST", "/sync/settings/set"}
	}
	// API.FetchUnreads() -> GET /sync/unreads
	RouteFetchUnreads = func() Route {
		return Route{"GET", "/sync/unreads"}
	}
	// API.PushSubscribe() -> POST /push/subscribe
	RoutePushSubscribe = func() Route {
		return Route{"POST", "/push/subscribe"}
	}
	// API.Unsubscribe() -> POST /push/unsubscribe
	RouteUnsubscribe = func() Route {
		return Route{"POST", "/push/unsubscribe"}
	}
)

type HTTPClient interface {
	Perform(*http.Request) (*http.Response, error)
}

type DefaultHTTPClient struct {
	Client http.Client
}

func NewDefaultHTTPClient(client http.Client) DefaultHTTPClient {
	return DefaultHTTPClient{Client: client}
}

func (h DefaultHTTPClient) Perform(req *http.Request) (*http.Response, error) {
	return h.Client.Do(req)
}

type TokenType int

const (
	TokenTypeUser TokenType = iota
	TokenTypeBot
)

type Token struct {
	Type  TokenType
	Token string
}

func NewUserToken(token string) *Token {
	return &Token{Type: TokenTypeUser, Token: token}
}

func NewBotToken(token string) *Token {
	return &Token{Type: TokenTypeBot, Token: token}
}

type AutumnAPI struct {
	Token      *Token
	HTTPClient HTTPClient
	URL        *url.URL
}

type AutumnAPIConfig struct {
	HTTPClient HTTPClient
	URL        *url.URL
}

func NewAutumnAPI(token *Token, config *AutumnAPIConfig) (api *AutumnAPI, err error) {
	if config == nil {
		config = &AutumnAPIConfig{}
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = NewDefaultHTTPClient(http.Client{})
	}
	apiUrl := config.URL
	if apiUrl == nil {
		apiUrl, err = url.Parse("https://autumn.revolt.chat/")
		if err != nil {
			return
		}
	}
	api = &AutumnAPI{
		Token:      token,
		HTTPClient: httpClient,
		URL:        apiUrl,
	}
	return
}

var (
	AutumnRouteFetchConfig = func() Route {
		return Route{"GET", "/"}
	}
	AutumnRouteUpload = func(tag string) Route {
		return Route{"POST", "/" + url.PathEscape(tag)}
	}
	AutumnRouteGet = func(tag, id string) Route {
		return Route{"GET", "/" + url.PathEscape(tag) + "/" + url.PathEscape(id)}
	}
)

func (api *AutumnAPI) Request(route Route, options *RequestOptions) (*http.Response, error) {
	if options == nil {
		options = &RequestOptions{}
	}
	header := options.Header
	if header == nil {
		header = http.Header{}
	}
	if !options.ManualAccept && len(header.Get("Accept")) == 0 {
		header.Set("Accept", "application/json")
	}
	if len(header.Get("User-Agent")) == 0 {
		header.Set("User-Agent", "Regolt (https://github.com/DarpHome/regolt, "+Version+")")
	}
	if !options.Unauthenticated && api.Token != nil {
		h := "X-Session-Token"
		if api.Token.Type == TokenTypeBot {
			h = "X-Bot-Token"
		} else if api.Token.Type != TokenTypeUser {
			panic("Don't use ints, use lib consts")
		}
		header.Set(h, api.Token.Token)
	}
	body := options.Body
	if options.JSON != nil && body == nil {
		if len(header.Get("Content-Type")) == 0 {
			header.Set("Content-Type", "application/json")
		}
		b, err := json.Marshal(options.JSON)
		if err != nil {
			return nil, err
		}
		body = io.NopCloser(bytes.NewReader(b))
	}
	u := api.URL.JoinPath(strings.TrimLeft(route.Path, "/"))
	if options.QueryValues != nil {
		u.RawQuery = options.QueryValues.Encode()
	}
	response, err := api.HTTPClient.Perform(&http.Request{
		Method: route.Method,
		URL:    u,
		Body:   body,
		Header: header,
	})
	if err != nil {
		return nil, err
	}
	if err = handleResponse(response); err != nil {
		return nil, err
	}
	return response, nil
}

func (api *AutumnAPI) RequestJSON(v any, route Route, options *RequestOptions) error {
	response, err := api.Request(route, options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if v != nil {
		err = json.NewDecoder(response.Body).Decode(v)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

type UploadTag string

const (
	UploadTagAttachments UploadTag = "attachments"
	UploadTagAvatars     UploadTag = "avatars"
	UploadTagBackgrounds UploadTag = "backgrounds"
	UploadTagIcons       UploadTag = "icons"
	UploadTagBanners     UploadTag = "banners"
	UploadTagEmojis      UploadTag = "emojis"
)

type AutumnTag struct {
	MaxSize             int      `json:"max_size"`
	UseULID             bool     `json:"use_ulid"`
	Enabled             bool     `json:"enabled"`
	ServeIfFieldPresent []string `json:"serve_if_field_present,omitempty"`
	RestrictContentType string   `json:"restrict_content_type"`
}

type AutumnConfig struct {
	Version     string                `json:"autumn"`
	Tags        map[string]*AutumnTag `json:"tags"`
	JPEGQuality int                   `json:"jpeg_quality"`
}

func (api *AutumnAPI) FetchConfig() (c *AutumnConfig, err error) {
	err = api.RequestJSON(&c, AutumnRouteFetchConfig(), nil)
	return
}

// Upload file to Autumn.
// tag [required] - tag, valid tags are: ["attachments", "avatars", "backgrounds", "banners", "emojis", "icons"]
// filename [required] - filename that will displayed in client
// contentType [optional, pass empty] - content type
// contents [required] - the file contents
// Example: `id, err := autumn.Upload("attachments", "hello.txt", "", []byte("hello world"))`
func (api *AutumnAPI) Upload(tag UploadTag, filename, contentType string, contents []byte) (s string, err error) {
	buf := &bytes.Buffer{}
	mpw := multipart.NewWriter(buf)
	w, err := mpw.CreateFormFile("file", filename)
	if err != nil {
		return
	}
	_, err = w.Write(contents)
	if err != nil {
		return
	}
	h := http.Header{}
	h.Set("Content-Type", mpw.FormDataContentType())
	mpw.Close()
	t := struct {
		ID string `json:"id"`
	}{}
	err = api.RequestJSON(&t, AutumnRouteUpload(string(tag)), &RequestOptions{
		Body:   io.NopCloser(buf),
		Header: h,
	})
	s = t.ID
	return
}

// Get file from Autumn.
// tag [required] - tag, valid tags are: ["attachments", "avatars", "backgrounds", "icons", "banners", "emojis"]
// id [required] - file ID
// Example: `id, err := autumn.Get("attachments", "123")`
func (api *AutumnAPI) Get(tag, id string) ([]byte, error) {
	response, err := api.Request(AutumnRouteGet(tag, id), nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type API struct {
	Token      *Token
	HTTPClient HTTPClient
	URL        *url.URL
}

type APIConfig struct {
	HTTPClient HTTPClient
	URL        *url.URL
}

func NewAPI(token *Token, config *APIConfig) (api *API, err error) {
	if config == nil {
		config = &APIConfig{}
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = NewDefaultHTTPClient(http.Client{})
	}
	apiUrl := config.URL
	if apiUrl == nil {
		apiUrl, err = url.Parse("https://api.revolt.chat/")
		if err != nil {
			return
		}
	}
	api = &API{
		Token:      token,
		HTTPClient: httpClient,
		URL:        apiUrl,
	}
	return
}

type APIError struct {
	Response   *http.Response `json:"-"`
	Type       string         `json:"type"`
	RetryAfter float64        `json:"retry_after"`
	Err        string         `json:"error"`
	Max        int            `json:"max"`
	Permission string         `json:"permission"`
	Operation  string         `json:"operation"`
	Collection string         `json:"collection"`
	Location   string         `json:"location"`
	With       string         `json:"with"`
}

func (ae APIError) Error() string {
	errs := []string{}
	if len(ae.Err) != 0 {
		errs = append(errs, ae.Err)
	}
	if ae.Max > 0 {
		errs = append(errs, fmt.Sprintf("max: %d", ae.Max))
	}
	if len(ae.Permission) != 0 {
		errs = append(errs, fmt.Sprintf("permission: %s", ae.Permission))
	}
	if len(ae.Operation) != 0 {
		errs = append(errs, fmt.Sprintf("operation: %s", ae.Operation))
	}
	if len(ae.Collection) != 0 {
		errs = append(errs, fmt.Sprintf("collection: %s", ae.Collection))
	}
	if len(ae.Location) != 0 {
		errs = append(errs, fmt.Sprintf("in: %s", ae.Location))
	}
	if len(ae.With) != 0 {
		errs = append(errs, fmt.Sprintf("with: %s", ae.With))
	}
	if len(errs) == 0 {
		return ae.Type
	}
	return ae.Type + ": " + strings.Join(errs, ", ")
}

type RequestOptions struct {
	Body            io.ReadCloser
	JSON            any
	Header          http.Header
	Unauthenticated bool
	ManualAccept    bool
	QueryValues     url.Values
}

func handleResponse(response *http.Response) error {
	if response.StatusCode >= 400 {
		defer response.Body.Close()
		e := APIError{Response: response}
		t := struct {
			Response   *http.Response `json:"-"`
			Type       string         `json:"type"`
			RetryAfter float64        `json:"retry_after"`
			Err        any            `json:"error"`
			Max        int            `json:"max"`
			Permission string         `json:"permission"`
			Operation  string         `json:"operation"`
			Collection string         `json:"collection"`
			Location   string         `json:"location"`
			With       string         `json:"with"`
		}{}
		if err := json.NewDecoder(response.Body).Decode(&t); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		switch v := t.Err.(type) {
		case map[string]any: // Rocket error
			tempCode, _ := v["code"].(float64)
			code := int(tempCode)
			reason, _ := v["reason"].(string)
			description, _ := v["description"].(string)
			t.Type = "Rocket error"
			e.Err = fmt.Sprintf("%d %s: %s", code, reason, description)
		case string: // Revolt error
			e.Err = v
		}
		e.Type = t.Type
		e.RetryAfter = t.RetryAfter
		e.Max = t.Max
		e.Permission = t.Permission
		e.Operation = t.Operation
		e.Collection = t.Collection
		e.Location = t.Location
		e.With = t.With
		return e
	}
	return nil
}

func (api *API) Request(route Route, options *RequestOptions) (*http.Response, error) {
	if options == nil {
		options = &RequestOptions{}
	}
	header := options.Header
	if header == nil {
		header = http.Header{}
	}
	if !options.ManualAccept && len(header.Get("Accept")) == 0 {
		header.Set("Accept", "application/json")
	}
	if len(header.Get("User-Agent")) == 0 {
		header.Set("User-Agent", "Regolt (https://github.com/DarpHome/regolt, "+Version+")")
	}
	if !options.Unauthenticated && api.Token != nil {
		h := "X-Session-Token"
		if api.Token.Type == TokenTypeBot {
			h = "X-Bot-Token"
		} else if api.Token.Type != TokenTypeUser {
			panic("Don't use ints, use lib consts")
		}
		header.Set(h, api.Token.Token)
	}
	body := options.Body
	if options.JSON != nil && body == nil {
		if len(header.Get("Content-Type")) == 0 {
			header.Set("Content-Type", "application/json")
		}
		b, err := json.Marshal(options.JSON)
		if err != nil {
			return nil, err
		}
		body = io.NopCloser(bytes.NewReader(b))
	}
	u := api.URL.JoinPath(strings.TrimLeft(route.Path, "/"))
	if options.QueryValues != nil {
		u.RawQuery = options.QueryValues.Encode()
	}
	response, err := api.HTTPClient.Perform(&http.Request{
		Method: route.Method,
		URL:    u,
		Body:   body,
		Header: header,
	})
	if err != nil {
		return nil, err
	}
	if err = handleResponse(response); err != nil {
		return nil, err
	}
	return response, nil
}

func (api *API) RequestJSON(v any, route Route, options *RequestOptions) error {
	response, err := api.Request(route, options)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if v != nil {
		err = json.NewDecoder(response.Body).Decode(v)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

func (api *API) RequestNone(route Route, options *RequestOptions) error {
	o := options
	if o == nil {
		o = &RequestOptions{}
	}
	o.ManualAccept = true
	response, err := api.Request(route, o)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

type HCaptchaConfiguration struct {
	// Whether captcha is enabled
	Enabled bool `json:"enabled"`
	// Client key used for solving captcha
	Key string `json:"key"`
}

type GenericServiceConfiguration struct {
	// Whether the service is enabled
	Enabled bool `json:"enabled"`
	// URL pointing to the service
	URL string `json:"url"`
}

type VoiceServerConfiguration struct {
	// Whether voice is enabled
	Enabled bool `json:"enabled"`
	// URL pointing to the voice API
	URL string `json:"url"`
	// URL pointing to the voice WebSocket server
	WS string `json:"ws"`
}

type BuildInformation struct {
	CommitSHA       string `json:"commit_sha"`
	CommitTimestamp string `json:"commit_timestamp"`
	Semver          string `json:"semver"`
	OriginURL       string `json:"origin_url"`
	Timestamp       string `json:"timestamp"`
}

type FeatureConfiguration struct {
	/// hCaptcha configuration
	Captcha HCaptchaConfiguration `json:"captcha"`
	// Whether email verification is enabled
	Email bool `json:"email"`
	// Whether this server is invite only
	InviteOnly bool `json:"invite_only"`
	// File server service configuration
	Autumn GenericServiceConfiguration `json:"autumn"`
	// Proxy service configuration
	January GenericServiceConfiguration `json:"january"`
	// Voice server configuration
	Voso VoiceServerConfiguration `json:"voso"`
}

type Node struct {
	// Revolt API Version
	Revolt string `json:"revolt"`
	// Features enabled on this Revolt node
	Features FeatureConfiguration `json:"features"`
	// WebSocket URL
	WS string `json:"ws"`
	// URL pointing to the client serving this node
	App string `json:"app"`
	// Web Push VAPID public key
	VAPID string `json:"vapid"`
	// Build information
	Build BuildInformation `json:"build"`
}

// Fetch the server configuration for this Revolt instance.
// https://developers.revolt.chat/api/#tag/Core/operation/root_root
func (api *API) QueryNode() (n *Node, err error) {
	err = api.RequestJSON(&n, RouteQueryNode(), &RequestOptions{Unauthenticated: true})
	return
}

// Retrieve your user information.
// https://developers.revolt.chat/api/#tag/User-Information/operation/fetch_self_req
func (api *API) FetchSelf() (u *User, err error) {
	err = api.RequestJSON(&u, RouteFetchSelf(), nil)
	return
}

// Retrieve a user's information.
// https://developers.revolt.chat/api/#tag/User-Information/operation/fetch_user_req
func (api *API) FetchUser(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteFetchUser(user), nil)
	return
}

type EditProfile struct {
	// Text to set as user profile description
	Content *string `json:"content,omitempty"`
	// Attachment Id for background
	Background string `json:"background,omitempty"`
}

type EditUser struct {
	// New display name
	DisplayName *string
	// Attachment ID for avatar
	Avatar *string
	// User's active status
	Status *UserStatus
	// New user profile data. This is applied as a partial.
	Profile *EditProfile
	// Bitfield of user badges
	Badges UserBadges
	// Enum of user flags
	Flags UserFlags
	// Fields to remove from user object
	// Possible values: ["Avatar", "StatusText", "StatusPresence", "ProfileContent", "ProfileBackground", "DisplayName"]
	Remove []string
}

func (eu EditUser) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if eu.DisplayName != nil {
		r["display_name"] = eu.DisplayName
	}
	if eu.Avatar != nil {
		r["avatar"] = eu.Avatar
	}
	if eu.Status != nil {
		r["status"] = eu.Status
	}
	if eu.Profile != nil {
		r["profile"] = eu.Profile
	}
	if eu.Badges != 0 {
		r["badges"] = eu.Badges
	}
	if eu.Flags != 0 {
		r["flags"] = eu.Flags
	}
	if len(eu.Remove) > 0 {
		r["remove"] = eu.Remove
	}
	return json.Marshal(r)
}

// Edit currently authenticated user.
// https://developers.revolt.chat/api/#tag/User-Information/operation/edit_user_req
func (api *API) EditUser(user ULID, params EditUser) (u *User, err error) {
	err = api.RequestJSON(&u, RouteEditUser(user), &RequestOptions{
		JSON: params,
	})
	return
}

// Retrieve a user's flags.
// https://developers.revolt.chat/api/#tag/User-Information/operation/fetch_user_flags_fetch_user_flags
func (api *API) FetchUserFlags(user ULID) (UserFlags, error) {
	t := struct {
		Flags UserFlags `json:"flags"`
	}{}
	err := api.RequestJSON(&t, RouteFetchUserFlags(user), nil)
	if err != nil {
		return 0, err
	}
	return t.Flags, nil
}

type ChangeUsername struct {
	// New username
	Username string `json:"username"`
	// Current account password
	Password string `json:"password"`
}

// Change your username.
// https://developers.revolt.chat/api/#tag/User-Information/operation/change_username_req
func (api *API) ChangeUsername(params *ChangeUsername) (u *User, err error) {
	err = api.RequestJSON(&u, RouteChangeUsername(), &RequestOptions{JSON: params})
	return
}

// This returns a default avatar based on the given id.
// https://developers.revolt.chat/api/#tag/User-Information/operation/get_default_avatar_req
func (api *API) FetchDefaultAvatar(user ULID) ([]byte, error) {
	response, err := api.Request(RouteFetchDefaultAvatar(user), &RequestOptions{ManualAccept: true, Unauthenticated: true})
	if err != nil {
		return []byte{}, nil
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

// Retrieve a user's profile data.
// Will fail if you do not have permission to access the other user's profile.
// https://developers.revolt.chat/api/#tag/User-Information/operation/fetch_profile_req
func (api *API) FetchUserProfile(user ULID) (p *UserProfile, err error) {
	err = api.RequestJSON(&p, RouteFetchUserProfile(user), nil)
	return
}

// This fetches your direct messages, including any DM and group DM conversations.
// https://developers.revolt.chat/api/#tag/Direct-Messaging/operation/fetch_dms_req
func (api *API) FetchDirectMessageChannels() (a []*Channel, err error) {
	err = api.RequestJSON(&a, RouteFetchDirectMessageChannels(), nil)
	return
}

// Open a DM with another user.
// If the target is oneself, a saved messages channel is returned.
// https://developers.revolt.chat/api/#tag/Direct-Messaging/operation/open_dm_req
func (api *API) OpenDirectMessage(user ULID) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteOpenDirectMessage(user), nil)
	return
}

// Retrieve a list of mutual friends and servers with another user.
// https://developers.revolt.chat/api/#tag/Relationships/operation/find_mutual_req
func (api *API) FetchMutualFriendsAndServers(user ULID) (users []ULID, servers []ULID, err error) {
	t := struct {
		Users   []ULID `json:"users"`
		Servers []ULID `json:"servers"`
	}{}
	err = api.RequestJSON(&t, RouteFetchMutualFriendsAndServers(user), nil)
	if err != nil {
		users = t.Users
		servers = t.Servers
	}
	return
}

// Accept another user's friend request.
// https://developers.revolt.chat/api/#tag/Relationships/operation/add_friend_req
func (api *API) AcceptFriendRequest(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteAcceptFriendRequest(user), nil)
	return
}

// Denies another user's friend request.
// https://developers.revolt.chat/api/#tag/Relationships/operation/remove_friend_req
func (api *API) DenyFriendRequest(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteDenyFriendRequest(user), nil)
	return
}

// Denies another user's friend request or removes an existing friend.
// https://developers.revolt.chat/api/#tag/Relationships/operation/remove_friend_req
func (api *API) RemoveFriend(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteRemoveFriend(user), nil)
	return
}

// Block another user by their ID.
// https://developers.revolt.chat/api/#tag/Relationships/operation/block_user_req
func (api *API) BlockUser(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteBlockUser(user), nil)
	return
}

// Unblock another user by their ID.
// https://developers.revolt.chat/api/#tag/Relationships/operation/unblock_user_req
func (api *API) UnblockUser(user ULID) (u *User, err error) {
	err = api.RequestJSON(&u, RouteUnblockUser(user), nil)
	return
}

// Send a friend request to another user.
// username [required] - Username and discriminator combo separated by #
// https://developers.revolt.chat/api/#tag/Relationships/operation/send_friend_request_req
func (api *API) SendFriendRequest(username string) (u *User, err error) {
	err = api.RequestJSON(&u, RouteSendFriendRequest(), &RequestOptions{
		JSON: struct {
			Username string `json:"username"`
		}{username},
	})
	return
}

// Create a new Revolt bot.
// name [required] - Bot username
// https://developers.revolt.chat/api/#tag/Bots/operation/create_create_bot
func (api *API) CreateBot(name string) (b *Bot, err error) {
	err = api.RequestJSON(&b, RouteCreateBot(), &RequestOptions{
		JSON: struct {
			Name string `json:"name"`
		}{name},
	})
	return
}

// Fetch details of a public (or owned) bot by its id.
// https://developers.revolt.chat/api/#tag/Bots/operation/fetch_public_fetch_public_bot
func (api *API) FetchPublicBot(bot ULID) (b *PublicBot, err error) {
	err = api.RequestJSON(&b, RouteFetchPublicBot(bot), nil)
	return
}

type InviteBot struct {
	// Server ID
	Server ULID `json:"server,omitempty"`
	// Group ID
	Group ULID `json:"group,omitempty"`
}

// Invite a bot to a server or group by its id.
// https://developers.revolt.chat/api/#tag/Bots/operation/invite_invite_bot
func (api *API) InviteBot(bot ULID, params *InviteBot) error {
	return api.RequestNone(RouteInviteBot(bot), &RequestOptions{JSON: params})
}

// Fetch details of a bot you own by its id.
// https://developers.revolt.chat/api/#tag/Bots/operation/fetch_fetch_bot
func (api *API) FetchBot(bot ULID) (b *Bot, u *User, err error) {
	t := struct {
		Bot  *Bot  `json:"bot"`
		User *User `json:"user"`
	}{}
	err = api.RequestJSON(&t, RouteFetchBot(bot), nil)
	if err == nil {
		b = t.Bot
		u = t.User
	}
	return
}

// Fetch all of the bots that you have control over.
// https://developers.revolt.chat/api/#tag/Bots/operation/fetch_owned_fetch_owned_bots
func (api *API) FetchOwnedBots() (b []*Bot, u []*User, err error) {
	t := struct {
		Bots  []*Bot  `json:"bots"`
		Users []*User `json:"users"`
	}{}
	err = api.RequestJSON(&t, RouteFetchOwnedBots(), nil)
	if err == nil {
		b = t.Bots
		u = t.Users
	}
	return
}

// Delete a bot by its ID.
// https://developers.revolt.chat/api/#tag/Bots/operation/delete_delete_bot
func (api *API) DeleteBot(bot ULID) error {
	return api.RequestNone(RouteDeleteBot(bot), nil)
}

type EditBot struct {
	// Bot username
	Name string
	// Whether the bot can be added by anyone
	Public *bool
	// Whether analytics should be gathered for this bot
	// Must be enabled in order to show up on Revolt Discover.
	Analytics *bool
	// Interactions URL
	InteractionsURL *string
	// Fields to remove from bot object
	// Possible values: ["Token", "InteractionsURL"]
	Remove []string
}

func (eb EditBot) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if len(eb.Name) != 0 {
		r["name"] = eb.Name
	}
	if eb.Public != nil {
		r["public"] = eb.Public
	}
	if eb.Analytics != nil {
		r["analytics"] = eb.Analytics
	}
	if eb.InteractionsURL != nil {
		if len(*eb.InteractionsURL) == 0 {
			r["interactions_url"] = nil
		} else {
			r["interactions_url"] = *eb.InteractionsURL
		}
	}
	if len(eb.Remove) > 0 {
		r["remove"] = eb.Remove
	}
	return json.Marshal(r)
}

// Edit bot details by its ID.
// https://developers.revolt.chat/api/#tag/Bots/operation/edit_edit_bot
func (api *API) EditBot(bot ULID, params EditBot) (b *Bot, err error) {
	err = api.RequestJSON(&b, RouteEditBot(bot), &RequestOptions{JSON: params})
	return
}

// Fetch channel by its ID.
// https://developers.revolt.chat/api/#tag/Channel-Information/operation/channel_fetch_req
func (api *API) FetchChannel(channel ULID) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteFetchChannel(channel), nil)
	return
}

// Deletes a server channel, leaves a group or closes a group.
// https://developers.revolt.chat/api/#tag/Channel-Information/operation/channel_delete_req
func (api *API) CloseChannel(channel ULID, leaveSilently *bool) error {
	v := url.Values{}
	if leaveSilently != nil {
		if *leaveSilently {
			v.Set("leave_silently", "true")
		} else {
			v.Set("leave_silently", "false")
		}
	}
	return api.RequestNone(RouteCloseChannel(channel), &RequestOptions{QueryValues: v})
}

type EditChannel struct {
	// Channel name
	Name string
	// Channel description
	Description *string
	// Group owner
	Owner ULID
	// Icon to set. Provide an Autumn attachment ID.
	Icon *string
	// Whether this channel is age-restricted
	NSFW *bool
	// Whether this channel is archived
	Archived *bool
	// Fields to remove from channel object
	// Possible values: ["Description", "Icon", "DefaultPermissions"]
	Remove []string
}

func (ec EditChannel) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if len(ec.Name) != 0 {
		r["name"] = ec.Name
	}
	if ec.Description != nil {
		if len(*ec.Description) == 0 {
			r["description"] = nil
		} else {
			r["description"] = ec.Description
		}
	}
	if len(ec.Owner) != 0 {
		r["owner"] = ec.Owner
	}
	if ec.Icon != nil {
		if len(*ec.Icon) == 0 {
			r["icon"] = nil
		} else {
			r["icon"] = *ec.Icon
		}
	}
	if ec.NSFW != nil {
		r["nsfw"] = *ec.NSFW
	}
	if ec.Archived != nil {
		r["archived"] = *ec.Archived
	}
	if len(ec.Remove) > 0 {
		r["remove"] = ec.Remove
	}
	return json.Marshal(r)
}

// Edit a channel object by its id.
// https://developers.revolt.chat/api/#tag/Channel-Information/operation/channel_edit_req
func (api *API) EditChannel(channel ULID, params *EditChannel) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteEditChannel(channel), &RequestOptions{JSON: params})
	return
}

// Creates an invite to this channel.
// Channel must be a `TextChannel`.
// https://developers.revolt.chat/api/#tag/Channel-Invites/operation/invite_create_req
func (api *API) CreateInvite(channel ULID) (i *Invite, err error) {
	err = api.RequestJSON(&i, RouteCreateInvite(channel), nil)
	return
}

// Sets permissions for the specified role in this channel.
// Channel must be a `TextChannel` or `VoiceChannel`.
// https://developers.revolt.chat/api/#tag/Channel-Permissions/operation/permissions_set_req
func (api *API) SetRoleChannelPermission(channel, role ULID, allow, deny Permissions) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteSetRoleChannelPermission(channel, role), &RequestOptions{
		JSON: struct {
			Permissions any `json:"permissions"`
		}{Permissions: struct {
			Allow Permissions `json:"allow"`
			Deny  Permissions `json:"deny"`
		}{allow, deny}},
	})
	return
}

type SetDefaultPermission struct {
	// Permission values to set for members in a `Group`
	Permission Permissions
	// Representation of a single permission override
	Permissions *PermissionOverride
}

func (p SetDefaultPermission) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if p.Permission != 0 {
		r["permissions"] = p.Permission
	} else {
		r["permissions"] = struct {
			Allow Permissions `json:"allow"`
			Deny  Permissions `json:"deny"`
		}{p.Permissions.Allow, p.Permissions.Disallow}
	}
	return json.Marshal(r)
}

// Sets permissions for the default role in this channel.
// Channel must be a `Group`, `TextChannel` or `VoiceChannel`.
// https://developers.revolt.chat/api/#tag/Channel-Permissions/operation/permissions_set_default_req
func (api *API) SetDefaultChannelPermission(channel ULID, params *SetDefaultPermission) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteSetDefaultChannelPermission(channel), &RequestOptions{JSON: params})
	return
}

// Lets the server and all other clients know that we've seen this message in this channel.
// https://developers.revolt.chat/api/#tag/Messaging/operation/channel_ack_req
func (api *API) AcknowledgeMessage(channel, message ULID) error {
	return api.RequestNone(RouteAcknowledgeMessage(channel, message), nil)
}

type MessageSort int

const (
	MessageSortByNone MessageSort = iota
	MessageSortByRelevance
	MessageSortByLatest
	MessageSortByOldest
)

func (v MessageSort) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v MessageSort) String() string {
	switch v {
	case MessageSortByRelevance:
		return "Relevance"
	case MessageSortByLatest:
		return "Latest"
	case MessageSortByOldest:
		return "Oldest"
	}
	return ""
}

type FetchMessages struct {
	// Maximum number of messages to fetch
	// For fetching nearby messages, this is `(limit + 1)`.
	Limit int
	// Message id before which messages should be fetched
	Before ULID
	// Message id after which messages should be fetched
	After ULID
	// Message sort direction
	Sort MessageSort
	// Message id to search around
	// Specifying 'nearby' ignores 'before', 'after' and 'sort'. It will also take half of limit rounded as the limits to each side. It also fetches the message ID specified.
	Nearby ULID
	// Whether to include user (and member, if server channel) objects
	IncludeUsers *bool
}

type Messages struct {
	Messages []*Message `json:"messages"`
	Users    []*User    `json:"users"`
	Members  []*Member  `json:"members"`
}

// Fetch multiple messages.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_query_req
func (api *API) FetchMessages(channel ULID, params *FetchMessages) (m *Messages, err error) {
	v := url.Values{}
	if params != nil {
		if params.Limit != 0 {
			v.Set("limit", strconv.Itoa(params.Limit))
		}
		if len(params.Before) != 0 {
			v.Set("before", string(params.Before))
		}
		if len(params.After) != 0 {
			v.Set("after", string(params.After))
		}
		if params.Sort != MessageSortByNone {
			v.Set("sort", params.Sort.String())
		}
		if len(params.Nearby) != 0 {
			v.Set("nearby", string(params.Nearby))
		}
		if params.IncludeUsers != nil {
			if *params.IncludeUsers {
				v.Set("include_users", "true")
			} else {
				v.Set("include_users", "false")
			}
		}
	}

	if params != nil && params.IncludeUsers != nil && *params.IncludeUsers {
		err = api.RequestJSON(&m, RouteFetchMessages(channel), &RequestOptions{QueryValues: v})
	} else {
		a := []*Message{}
		err = api.RequestJSON(&a, RouteFetchMessages(channel), &RequestOptions{QueryValues: v})
		if err == nil {
			m = &Messages{Messages: a}
		}
	}
	return
}

type Reply struct {
	// Message ID
	ID ULID `json:"id"`
	// Whether this reply should mention the message's author
	Mention bool `json:"mention"`
}

type SendableEmbed struct {
	IconURL     string `json:"icon_url,omitempty"`
	URL         string `json:"url,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Media       string `json:"media,omitempty"`
	Colour      string `json:"colour,omitempty"`
}

type SendMessage struct {
	// Unique token to prevent duplicate message sending
	IdempotencyKey string `json:"-"`
	// Message content to send
	Content string `json:"content,omitempty"`
	// Attachments to include in message
	Attachments []string `json:"attachments,omitempty"`
	// Messages to reply to
	Replies []Reply `json:"replies,omitempty"`
	// Embeds to include in message
	// Text embed content contributes to the content length cap
	Embeds []SendableEmbed `json:"embeds,omitempty"`
	// Name and / or avatar override information
	Masquerade *Masquerade `json:"masquerade,omitempty"`
	// Information to guide interactions on this message
	Interactions *MessageInteractions `json:"interactions,omitempty"`
}

func (sm *SendMessage) SetContent(content string) *SendMessage {
	sm.Content = content
	return sm
}

func (sm *SendMessage) SetAttachments(attachments []string) *SendMessage {
	sm.Attachments = attachments
	return sm
}

func (sm *SendMessage) AddAttachments(attachments ...string) *SendMessage {
	return sm.SetAttachments(append(sm.Attachments, attachments...))
}

func (sm *SendMessage) SetReplies(replies []Reply) *SendMessage {
	sm.Replies = replies
	return sm
}

func (sm *SendMessage) AddReply(replies ...Reply) *SendMessage {
	return sm.SetReplies(append(sm.Replies, replies...))
}

func (sm *SendMessage) ReplyTo(messages ...ULID) *SendMessage {
	for _, m := range messages {
		sm.Replies = append(sm.Replies, Reply{ID: m})
	}
	return sm
}

func (sm *SendMessage) SetEmbeds(embeds []SendableEmbed) *SendMessage {
	sm.Embeds = embeds
	return sm
}

func (sm *SendMessage) AddEmbeds(embeds ...SendableEmbed) *SendMessage {
	return sm.SetEmbeds(append(sm.Embeds, embeds...))
}

func (sm *SendMessage) SetMasquerade(masquerade *Masquerade) *SendMessage {
	sm.Masquerade = masquerade
	return sm
}

func (sm *SendMessage) SetInteractions(interactions *MessageInteractions) *SendMessage {
	sm.Interactions = interactions
	return sm
}

// Sends a message to the given channel.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_send_message_send
func (api *API) SendMessage(channel ULID, params *SendMessage) (m *Message, err error) {
	h := http.Header{}
	if len(params.IdempotencyKey) != 0 {
		h.Set("Idempotency-Key", params.IdempotencyKey)
	}
	err = api.RequestJSON(&m, RouteSendMessage(channel), &RequestOptions{
		JSON:   params,
		Header: h,
	})
	return
}

type SearchForMessages struct {
	// Full-text search query
	// See [MongoDB documentation](https://docs.mongodb.com/manual/text-search/#-text-operator) for more information.
	Query string
	// Maximum number of messages to fetch
	Limit int
	// Message id before which messages should be fetched
	Before ULID
	// Message id after which messages should be fetched
	After ULID
	// Sort used for retrieving messages
	Sort MessageSort
	// Whether to include user (and member, if server channel) objects
	IncludeUsers *bool
}

// This route searches for messages within the given parameters.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_search_req
func (api *API) SearchForMessages(channel ULID, params *SearchForMessages) (m *Messages, err error) {
	v := url.Values{}
	v.Set("query", params.Query)
	if params.Limit != 0 {
		v.Set("limit", strconv.Itoa(params.Limit))
	}
	if len(params.Before) != 0 {
		v.Set("before", string(params.Before))
	}
	if len(params.After) != 0 {
		v.Set("after", string(params.After))
	}
	if params.Sort != MessageSortByNone {
		v.Set("sort", params.Sort.String())
	}
	if params.IncludeUsers != nil {
		if *params.IncludeUsers {
			v.Set("include_users", "true")
		} else {
			v.Set("include_users", "false")
		}
	}
	if params.IncludeUsers != nil && *params.IncludeUsers {
		err = api.RequestJSON(&m, RouteSearchForMessages(channel), &RequestOptions{QueryValues: v})
	} else {
		a := []*Message{}
		err = api.RequestJSON(&a, RouteSearchForMessages(channel), &RequestOptions{QueryValues: v})
		if err == nil {
			m = &Messages{Messages: a}
		}
	}
	return
}

// Retrieves a message by its id.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_fetch_req
func (api *API) FetchMessage(channel, message ULID) (m *Message, err error) {
	err = api.RequestJSON(&m, RouteFetchMessage(channel, message), nil)
	return
}

// Delete a message you've sent or one you have permission to delete.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_delete_req
func (api *API) DeleteMessage(channel, message ULID) error {
	return api.RequestNone(RouteDeleteMessage(channel, message), nil)
}

type EditMessage struct {
	// New message content
	Content *string `json:"content,omitempty"`
	// Embeds to include in the message
	Embeds *[]SendableEmbed `json:"embeds,omitempty"`
}

func (em *EditMessage) SetContent(content string) *EditMessage {
	em.Content = &content
	return em
}

func (em *EditMessage) UnsetContent() *EditMessage {
	em.Content = nil
	return em
}

func (em *EditMessage) SetEmbeds(embeds *[]SendableEmbed) *EditMessage {
	em.Embeds = embeds
	return em
}

func (em *EditMessage) UnsetEmbeds() *EditMessage {
	return em.SetEmbeds(nil)
}

func (em *EditMessage) AddEmbeds(embeds ...SendableEmbed) *EditMessage {
	currentEmbeds := em.Embeds
	if currentEmbeds == nil {
		currentEmbeds = &[]SendableEmbed{}
	}
	newEmbeds := append(*currentEmbeds, embeds...)
	return em.SetEmbeds(&newEmbeds)
}

// Edits a message that you've previously sent.
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_edit_req
func (api *API) EditMessage(channel, message ULID, params *EditMessage) (m *Message, err error) {
	err = api.RequestJSON(&m, RouteEditMessage(channel, message), &RequestOptions{JSON: params})
	return
}

// Delete multiple messages you've sent or one you have permission to delete.
// This will always require `ManageMessages` permission regardless of whether you own the message or not.
// Messages must have been sent within the past 1 week.
// ids [required] - Message IDs
// https://developers.revolt.chat/api/#tag/Messaging/operation/message_bulk_delete_req
func (api *API) BulkDeleteMessages(channel ULID, ids []ULID) error {
	return api.RequestNone(RouteBulkDeleteMessages(channel), &RequestOptions{
		JSON: struct {
			IDs []ULID `json:"ids"`
		}{ids},
	})
}

// React to a given message.
// https://developers.revolt.chat/api/#tag/Interactions/operation/message_react_react_message
func (api *API) AddReactionToMessage(channel, message ULID, emoji Emoji) error {
	return api.RequestNone(RouteAddReactionToMessage(channel, message, emoji), nil)
}

// Remove your own, someone else's or all of a given reaction.
// Requires `ManageMessages` if changing others' reactions.
// https://developers.revolt.chat/api/#tag/Interactions/operation/message_unreact_unreact_message
func (api *API) RemoveReactionsFromMessage(channel, message ULID, emoji Emoji) error {
	return api.RequestNone(RouteRemoveReactionsFromMessage(channel, message, emoji), nil)
}

// Remove your own, someone else's or all of a given reaction.
// Requires `ManageMessages` permission.
// https://developers.revolt.chat/api/#tag/Interactions/operation/message_clear_reactions_clear_reactions
func (api *API) RemoveAllReactionsFromMessage(channel, message ULID) error {
	return api.RequestNone(RouteRemoveAllReactionsFromMessage(channel, message), nil)
}

// Retrieves all users who are part of this group.
// https://developers.revolt.chat/api/#tag/Groups/operation/members_fetch_req
func (api *API) FetchGroupMembers(channel ULID) (a []*User, err error) {
	err = api.RequestJSON(&a, RouteFetchGroupMembers(channel), nil)
	return
}

type CreateGroup struct {
	// Group name
	Name string `json:"name"`
	// Group description
	Description string `json:"description,omitempty"`
	// Array of user IDs to add to the group
	// Must be friends with these users.
	Users []ULID `json:"users"`
	// Whether this group is age-restricted
	NSFW bool `json:"nsfw"`
}

// Create a new group channel.
// https://developers.revolt.chat/api/#tag/Groups/operation/group_create_req
func (api *API) CreateGroup(params *CreateGroup) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteCreateGroup(), &RequestOptions{
		JSON: params,
	})
	return
}

// Adds another user to the group.
// https://developers.revolt.chat/api/#tag/Groups/operation/group_add_member_req
func (api *API) AddMemberToGroup(channel, member ULID) error {
	return api.RequestNone(RouteAddMemberToGroup(channel, member), nil)
}

// Removes a user from the group.
// https://developers.revolt.chat/api/#tag/Groups/operation/group_remove_member_req
func (api *API) RemoveMemberFromGroup(channel, member ULID) error {
	return api.RequestNone(RouteRemoveMemberFromGroup(channel, member), nil)
}

// Asks the voice server for a token to join the call.
// https://developers.revolt.chat/api/#tag/Voice/operation/voice_join_req
func (api *API) JoinCall(channel ULID) (string, error) {
	t := struct {
		Token string `json:"token"`
	}{}
	if err := api.RequestJSON(&t, RouteJoinCall(channel), nil); err != nil {
		return "", err
	}
	return t.Token, nil
}

// Creates a webhook which 3rd party platforms can use to send messages.
// name [required] - The webhook name.
// avatar [optional, pass empty] - The webhook avatar. (Pass Autumn file ID)
// https://developers.revolt.chat/api/#tag/Webhooks/operation/webhook_create_req
func (api *API) CreateWebhook(channel ULID, name string, avatar string) (w *Webhook, err error) {
	err = api.RequestJSON(&w, RouteCreateWebhook(channel), &RequestOptions{
		JSON: struct {
			Name   string `json:"name"`
			Avatar string `json:"avatar,omitempty"`
		}{name, avatar},
	})
	return
}

// Deletes a webhook [with a token].
// webhook [required] - The webhook ID
// webhookToken [optional, pass empty] - The webhook private token.
func (api *API) DeleteWebhook(webhookID ULID, webhookToken string) error {
	if len(webhookToken) == 0 {
		return api.RequestNone(RouteDeleteWebhook(webhookID), nil)
	}
	return api.RequestNone(RouteDeleteWebhookWithToken(webhookID, webhookToken), &RequestOptions{Unauthenticated: true})
}

type EditWebhook struct {
	// New webhook name
	Name string `json:"name,omitempty"`
	// New avatar ID
	Avatar string `json:"avatar,omitempty"`
	// New webhook permissions
	Permissions *Permissions `json:"permissions,omitempty"`
	// Fields to remove from webhook
	// Possible values: ["Avatar"]
	Remove []string `json:"remove,omitempty"`
}

// Edits a webhook [with a token].
// webhook [required] - The webhook ID
// webhookToken [optional, pass empty] - The webhook private token.
// params [required] - What to change in webhook.
func (api *API) EditWebhook(webhookID ULID, webhookToken string, params *EditWebhook) (w *Webhook, err error) {
	if len(webhookToken) == 0 {
		err = api.RequestJSON(&w, RouteEditWebhook(webhookID), &RequestOptions{JSON: params})
	} else {
		err = api.RequestJSON(&w, RouteEditWebhookWithToken(webhookID, webhookToken), &RequestOptions{JSON: params, Unauthenticated: true})
	}
	return
}

// Fetches a webhook [with a token]. Requires valid API token if no webhook token passed.
// webhookID [required] - The webhook ID
// webhookToken [optional, pass empty] - The webhook private token.
func (api *API) FetchWebhook(webhookID ULID, webhookToken string) (w *Webhook, err error) {
	if len(webhookToken) == 0 {
		err = api.RequestJSON(&w, RouteFetchWebhook(webhookID), nil)
	} else {
		err = api.RequestJSON(&w, RouteFetchWebhookWithToken(webhookID, webhookToken), &RequestOptions{Unauthenticated: true})
	}
	return
}

// Fetches all webhooks inside the channel.
// https://developers.revolt.chat/api/#tag/Webhooks/operation/webhook_fetch_all_req
func (api *API) FetchChannelWebhooks(channel ULID) (a []*Webhook, err error) {
	err = api.RequestJSON(&a, RouteFetchChannelWebhooks(channel), nil)
	return
}

// Executes a webhook and sends a message.
// https://github.com/revoltchat/backend/blob/master/crates/delta/src/routes/webhooks/webhook_execute.rs (No OpenAPI docs)
func (api *API) ExecuteWebhook(webhook ULID, token string, params *SendMessage) (m *Message, err error) {
	err = api.RequestJSON(&m, RouteExecuteWebhook(webhook, token), &RequestOptions{JSON: params})
	return
}

type CreateServer struct {
	// Server name
	Name string `json:"name"`
	// Server description
	Description string `json:"description,omitempty"`
	// Whether this server is age-restricted
	NSFW bool `json:"nsfw"`
}

type ServerResponse struct {
	Server   Server    `json:"server"`
	Channels []Channel `json:"channels"`
}

// Create a new server.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_create_req
func (api *API) CreateServer(params *CreateServer) (sr *ServerResponse, err error) {
	err = api.RequestJSON(&sr, RouteCreateServer(), &RequestOptions{
		JSON: params,
	})
	return
}

// Fetch a server by its ID.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_fetch_req
func (api *API) FetchServer(server ULID) (s *Server, err error) {
	err = api.RequestJSON(&s, RouteFetchServer(server), nil)
	return
}

// Deletes a server if owner.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_delete_req
func (api *API) DeleteServer(server ULID) (err error) {
	err = api.RequestNone(RouteDeleteServer(server), nil)
	return
}

type LeaveServer struct {
	// Whether to not send a leave message
	LeaveSilently *bool
}

func (ls *LeaveServer) SetLeaveSilently(b bool) *LeaveServer {
	ls.LeaveSilently = &b
	return ls
}

func (ls *LeaveServer) UnsetLeaveSilently() *LeaveServer {
	ls.LeaveSilently = nil
	return ls
}

// Leaves a server if not owner.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_delete_req
func (api *API) LeaveServer(server ULID, params *LeaveServer) (err error) {
	if params == nil {
		params = &LeaveServer{}
	}
	v := url.Values{}
	if params.LeaveSilently != nil {
		if *params.LeaveSilently {
			v.Set("leave_silently", "true")
		} else {
			v.Set("leave_silently", "false")
		}
	}
	err = api.RequestNone(RouteLeaveServer(server), &RequestOptions{QueryValues: v})
	return
}

type EditSystemMessages struct {
	// ID of channel to send user join messages in
	UserJoined *ULID
	// ID of channel to send user left messages in
	UserLeft *ULID
	// ID of channel to send user kicked messages in
	UserKicked *ULID
	// ID of channel to send user banned messages in
	UserBanned *ULID
}

func (esm EditSystemMessages) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if esm.UserJoined != nil {
		if len(*esm.UserJoined) == 0 {
			r["user_joined"] = nil
		} else {
			r["user_joined"] = *esm.UserJoined
		}
	}
	if esm.UserLeft != nil {
		if len(*esm.UserLeft) == 0 {
			r["user_left"] = nil
		} else {
			r["user_left"] = *esm.UserLeft
		}
	}
	if esm.UserKicked != nil {
		if len(*esm.UserKicked) == 0 {
			r["user_kicked"] = nil
		} else {
			r["user_kicked"] = *esm.UserKicked
		}
	}
	if esm.UserBanned != nil {
		if len(*esm.UserBanned) == 0 {
			r["user_banned"] = nil
		} else {
			r["user_banned"] = *esm.UserBanned
		}
	}
	return json.Marshal(r)
}

type EditServer struct {
	// Server name
	Name string
	// Server description
	Description *string
	// Attachment ID for icon
	Icon *string
	// Attachment ID for banner
	Banner *string
	// Category structure for server
	Categories *[]Category
	// System message channel assignments
	SystemMessages *EditSystemMessages
	// Bitfield of server flags
	Flags ServerFlags
	// Whether this server is public and should show up on Revolt Discover
	Discoverable *bool
	// Whether analytics should be collected for this server
	// Must be enabled in order to show up on Revolt Discover.
	Analytics *bool
	// Fields to remove from server object
	// Possible values: ["Description", "Categories", "SystemMessages", "Icon", "Banner"]
	Remove []string
}

func (es *EditServer) SetName(name string) *EditServer {
	es.Name = name
	return es
}

func (es *EditServer) UnsetName() *EditServer {
	return es.SetName("")
}

func (es *EditServer) SetDescription(description string) *EditServer {
	es.Description = &description
	return es
}

func (es *EditServer) UnsetDescription() *EditServer {
	es.Description = nil
	return es
}

func (es *EditServer) SetIcon(icon string) *EditServer {
	es.Icon = &icon
	return es
}

func (es *EditServer) UnsetIcon() *EditServer {
	es.Icon = nil
	return es
}

func (es *EditServer) SetCategories(categories *[]Category) *EditServer {
	es.Categories = categories
	return es
}

func (es *EditServer) UnsetCategories() *EditServer {
	return es.SetCategories(nil)
}

func (es *EditServer) AddCategories(categories ...Category) *EditServer {
	currentCategories := es.Categories
	if currentCategories == nil {
		currentCategories = &[]Category{}
	}
	newCategories := append(*currentCategories, categories...)
	return es.SetCategories(&newCategories)
}

func (es *EditServer) SetSystemMessages(systemMessages EditSystemMessages) *EditServer {
	es.SystemMessages = &systemMessages
	return es
}

func (es *EditServer) UnsetSystemMessages() *EditServer {
	es.SystemMessages = nil
	return es
}

func (es *EditServer) SetFlags(flags ServerFlags) *EditServer {
	es.Flags = flags
	return es
}

func (es EditServer) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if len(es.Name) != 0 {
		r["name"] = es.Name
	}
	if es.Description != nil {
		if len(*es.Description) == 0 {
			r["description"] = nil
		} else {
			r["description"] = *es.Description
		}
	}
	if es.Icon != nil {
		if len(*es.Icon) == 0 {
			r["icon"] = nil
		} else {
			r["icon"] = *es.Icon
		}
	}
	if es.Banner != nil {
		if len(*es.Banner) == 0 {
			r["banner"] = nil
		} else {
			r["banner"] = *es.Banner
		}
	}
	if es.Categories != nil {
		r["categories"] = *es.Categories
	}
	if es.SystemMessages != nil {
		r["system_messages"] = es.SystemMessages
	}
	if es.Flags != 0 {
		r["flags"] = es.Flags
	}
	if es.Discoverable != nil {
		r["discoverable"] = *es.Discoverable
	}
	if es.Analytics != nil {
		r["analytics"] = *es.Analytics
	}
	if len(es.Remove) > 0 {
		r["remove"] = es.Remove
	}
	return json.Marshal(r)
}

// Edit a server by its ID.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_edit_req
func (api *API) EditServer(server ULID, params *EditServer) (s *Server, err error) {
	err = api.RequestJSON(&s, RouteEditServer(server), &RequestOptions{JSON: params})
	return
}

// Mark all channels in a server as read.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/server_ack_req
func (api *API) MarkServerAsRead(server ULID) (err error) {
	err = api.RequestNone(RouteMarkServerAsRead(server), nil)
	return
}

type CreateChannel struct {
	// Default: "Text"
	// Enum: "Text" "Voice"
	// Channel type
	Type ChannelType `json:"type,omitempty"`
	// Channel name
	Name string `json:"name"`
	// Channel description
	Description string `json:"description,omitempty"`
	// Whether this channel is age restricted
	NSFW bool `json:"nsfw"`
}

// Create a new Text or Voice channel.
// https://developers.revolt.chat/api/#tag/Server-Information/operation/channel_create_req
func (api *API) CreateChannel(server ULID, params *CreateChannel) (c *Channel, err error) {
	err = api.RequestJSON(&c, RouteCreateChannel(server), &RequestOptions{JSON: params})
	return
}

type FetchMembersResponse struct {
	// List of members
	Members []Member `json:"members"`
	// List of users
	Users []User `json:"users"`
}

type FetchMembers struct {
	// Whether to exclude offline users
	ExcludeOffline *bool
}

// Fetch all server members.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/member_fetch_all_req
func (api *API) FetchMembers(server ULID, params *FetchMembers) (fmr *FetchMembersResponse, err error) {
	v := url.Values{}
	if params != nil {
		if params.ExcludeOffline != nil {
			if *params.ExcludeOffline {
				v.Set("exclude_offline", "true")
			} else {
				v.Set("exclude_offline", "false")
			}
		}
	}
	err = api.RequestJSON(&fmr, RouteFetchMembers(server), &RequestOptions{QueryValues: v})
	return
}

// Retrieve a member.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/member_fetch_req
func (api *API) FetchMember(server, member ULID) (m *Member, err error) {
	err = api.RequestJSON(&m, RouteFetchMember(server, member), nil)
	return
}

// Removes a member from the server.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/member_remove_req
func (api *API) KickMember(server, member ULID) error {
	return api.RequestNone(RouteKickMember(server, member), nil)
}

type EditMember struct {
	// Member nickname
	Nickname *string
	// Attachment ID to set for avatar
	Avatar *string
	// Array of role IDs
	Roles *[]ULID
	// ISO8601 formatted timestamp
	Timeout *time.Time
	// Fields to remove from member object
	// Possible values: ["Nickname", "Avatar", "Roles", "Timeout"]
	Remove []string
}

func (em EditMember) MarshalJSON() ([]byte, error) {
	r := map[string]any{}
	if em.Nickname != nil {
		if len(*em.Nickname) == 0 {
			r["nickname"] = nil
		} else {
			r["nickname"] = *em.Nickname
		}
	}
	if em.Avatar != nil {
		if len(*em.Avatar) == 0 {
			r["avatar"] = nil
		} else {
			r["avatar"] = *em.Avatar
		}
	}
	if em.Roles != nil {
		r["roles"] = *em.Roles
	}
	if em.Timeout != nil {
		if em.Timeout.IsZero() {
			r["timeout"] = nil
		} else {
			r["timeout"] = em.Timeout.Format(iso8601Template)
		}
	}
	if len(em.Remove) > 0 {
		r["remove"] = em.Remove
	}
	return json.Marshal(r)
}

// Edit a member by their ID.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/member_edit_req
func (api *API) EditMember(server, member ULID, params *EditMember) (m *Member, err error) {
	err = api.RequestJSON(&m, RouteEditMember(server, member), &RequestOptions{
		JSON: params,
	})
	return
}

// Query members by a given name, this API is not stable and will be removed in the future.
// query [required] - String to search for
// https://developers.revolt.chat/api/#tag/Server-Members/operation/member_experimental_query_member_experimental_query
func (api *API) QueryMembersByName(server ULID, query string) (ma []*Member, mu []*User, err error) {
	t := struct {
		Members []*Member `json:"members"`
		Users   []*User   `json:"users"`
	}{}
	v := url.Values{}
	v.Set("query", query)
	v.Set("experimental_api", "true")
	err = api.RequestJSON(&t, RouteQueryMembersByName(server), &RequestOptions{QueryValues: v})
	ma = t.Members
	mu = t.Users
	return
}

type Ban struct {
	ID MemberID `json:"_id"`
	// Reason for ban creation
	Reason string `json:"reason"`
}

// Ban a user by their ID.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/ban_create_req
func (api *API) BanUser(server, user ULID, reason string) (b *Ban, err error) {
	err = api.RequestJSON(&b, RouteBanUser(server, user), &RequestOptions{
		JSON: struct {
			Reason string `json:"reason,omitempty"`
		}{reason},
	})
	return
}

// Remove a user's ban.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/ban_remove_req
func (api *API) UnbanUser(server, user ULID) error {
	return api.RequestNone(RouteUnbanUser(server, user), nil)
}

type BansResponse struct {
	// Users objects
	Users []User `json:"users"`
	// Ban objects
	Bans []Ban `json:"bans"`
}

// Fetch all bans on a server.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/ban_list_req
func (api *API) FetchBans(server ULID) (br *BansResponse, err error) {
	err = api.RequestJSON(&br, RouteFetchBans(server), nil)
	return
}

// Fetch all server invites.
// https://developers.revolt.chat/api/#tag/Server-Members/operation/invites_fetch_req
func (api *API) FetchInvites(server ULID) (i []*Invite, err error) {
	err = api.RequestJSON(&i, RouteFetchInvites(server), nil)
	return
}

type CreateRole struct {
	// Role name
	Name string `json:"name"`
	// Ranking position
	// Smaller values take priority.
	Rank *int `json:"rank,omitempty"`
}

type RoleResponse struct {
	// ID of the role
	ID ULID `json:"id"`
	// Representation of a server role
	Role Role `json:"role"`
}

// Creates a new server role.
// https://developers.revolt.chat/api/#tag/Server-Permissions/operation/roles_create_req
func (api *API) CreateRole(server ULID, params *CreateRole) (rr *RoleResponse, err error) {
	err = api.RequestJSON(&rr, RouteCreateRole(server), &RequestOptions{JSON: params})
	return
}

// Delete a server role by its ID.
// https://developers.revolt.chat/api/#tag/Server-Permissions/operation/roles_delete_req
func (api *API) DeleteRole(server, role ULID) error {
	return api.RequestNone(RouteDeleteRole(server, role), nil)
}

type EditRole struct {
	// Role name
	Name string `json:"name,omitempty"`
	// Role colour
	Colour string `json:"colour,omitempty"`
	// Whether this role should be displayed separately
	Hoist *bool `json:"hoist,omitempty"`
	// Ranking position
	// Smaller values take priority.
	Rank *int `json:"rank,omitempty"`
	// Fields to remove from role object
	// Possible values: ["Colour"]
	Remove []string `json:"remove,omitempty"`
}

// Edit a role by its ID.
// https://developers.revolt.chat/api/#tag/Server-Permissions/operation/roles_edit_req
func (api *API) EditRole(server, role ULID, params *EditRole) (r *Role, err error) {
	err = api.RequestJSON(&r, RouteEditRole(server, role), &RequestOptions{JSON: params})
	return
}

// Sets permissions for the specified role in the server.
// allow [required] - Allow bit flags
// deny [required] - Disallow bit flags
// https://developers.revolt.chat/api/#tag/Server-Permissions/operation/permissions_set_req
func (api *API) SetRoleServerPermission(server, role ULID, allow, deny Permissions) (s *Server, err error) {
	err = api.RequestJSON(&s, RouteSetRoleServerPermission(server, role), &RequestOptions{
		JSON: struct {
			Permissions any `json:"permissions"`
		}{Permissions: struct {
			Allow Permissions `json:"allow"`
			Deny  Permissions `json:"deny"`
		}{allow, deny}},
	})
	return
}

// Sets permissions for the default role in this server.
// https://developers.revolt.chat/api/#tag/Server-Permissions/operation/permissions_set_default_req
func (api *API) SetDefaultServerPermission(server ULID, params *SetDefaultPermission) (s *Server, err error) {
	err = api.RequestJSON(&s, RouteSetDefaultServerPermission(server), &RequestOptions{JSON: params})
	return
}

// Fetch an invite by its ID.
// https://developers.revolt.chat/api/#tag/Invites/operation/invite_fetch_req
func (api *API) FetchInvite(invite string) (i *InviteResponse, err error) {
	err = api.RequestJSON(&i, RouteFetchInvite(invite), nil)
	return
}

type JoinInviteResponse struct {
	// Channels in the server
	Channels []Channel `json:"channels"`
	// Representation of a server on Revolt
	Server Server `json:"server"`
}

// Join an invite by its ID.
// https://developers.revolt.chat/api/#tag/Invites/operation/invite_join_req
func (api *API) JoinInvite(invite string) (jir *JoinInviteResponse, err error) {
	err = api.RequestJSON(&jir, RouteJoinInvite(invite), nil)
	return
}

// Delete an invite by its ID.
// https://developers.revolt.chat/api/#tag/Invites/operation/invite_delete_req
func (api *API) DeleteInvite(invite string) error {
	return api.RequestNone(RouteDeleteInvite(invite), nil)
}

// Fetch an emoji by its ID.
// https://developers.revolt.chat/api/#tag/Emojis/operation/emoji_fetch_fetch_emoji
func (api *API) FetchEmoji(emoji ULID) (ce *CustomEmoji, err error) {
	err = api.RequestJSON(&ce, RouteFetchEmoji(emoji), nil)
	return
}

type CreateEmoji struct {
	// Emoji name
	Name string `json:"name"`
	// Information about what owns this emoji
	Parent CustomEmojiParent `json:"parent"`
	// Whether the emoji is mature
	NSFW bool `json:"nsfw"`
}

// Create an emoji by its Autumn upload ID.
// https://developers.revolt.chat/api/#tag/Emojis/operation/emoji_create_create_emoji
func (api *API) CreateEmoji(emoji string, params *CreateEmoji) (ce *CustomEmoji, err error) {
	err = api.RequestJSON(&ce, RouteCreateEmoji(emoji), nil)
	return
}

// Delete an emoji by its ID.
// https://developers.revolt.chat/api/#tag/Emojis/operation/emoji_delete_delete_emoji
func (api *API) DeleteEmoji(emoji ULID) error {
	return api.RequestNone(RouteDeleteEmoji(emoji), nil)
}

// Fetch all emojis on a server.
func (api *API) FetchServerEmojis(server ULID) (a []*CustomEmoji, err error) {
	err = api.RequestJSON(&a, RouteFetchServerEmojis(server), nil)
	return
}

// Fetch various technical statistics.
// https://developers.revolt.chat/api/#tag/Admin/operation/stats_stats
func (api *API) QueryStats() (st *InstanceStatistics, err error) {
	err = api.RequestJSON(&st, RouteQueryStats(), nil)
	return
}

type GloballyFetchMessages struct {
	// Message ID before which messages should be fetched
	Before ULID `json:"before,omitempty"`
	// Message ID after which messages should be fetched
	After ULID `json:"after,omitempty"`
	// Sort used for retrieving messages
	Sort MessageSort `json:"sort,omitempty"`
	// Message ID to search around.
	// Specifying 'nearby' ignores 'before', 'after' and 'sort'. It will also take half of limit rounded as the limits to each side. It also fetches the message ID specified.
	Nearby ULID `json:"nearby,omitempty"`
	// Maximum number of messages to fetch
	// For fetching nearby messages, this is `(limit + 1)`.
	Limit int `json:"limit,omitempty"`
	// Parent channel ID
	Channel ULID `json:"channel,omitempty"`
	// Message author ID
	Author ULID `json:"author,omitempty"`
	// Search query
	Query string `json:"query,omitempty"`
}

// This is a privileged route to globally fetch messages.
// https://developers.revolt.chat/api/#tag/Admin/operation/message_query_message_query
func (api *API) GloballyFetchMessages(params *GloballyFetchMessages) (m *Messages, err error) {
	var r json.RawMessage
	err = api.RequestJSON(&r, RouteGloballyFetchMessages(), &RequestOptions{JSON: params})
	if err == nil {
		if r[0] == '[' {
			m = &Messages{}
			err = json.Unmarshal(r, &m.Messages)
		} else {
			err = json.Unmarshal(r, &m)
		}
	}
	return
}

type EditReportStatus struct {
	Status          ReportStatus `json:"status"`
	RejectionReason string       `json:"rejection_reason,omitempty"`
	ClosedAt        *time.Time   `json:"closed_at,omitempty"`
}

type EditReport struct {
	// Status of the report
	Status *EditReportStatus `json:"status,omitempty"`
	// Report notes
	Notes string `json:"notes,omitempty"`
}

// Edit a report.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/edit_report_edit_report
func (api *API) EditReport(report ULID, params *EditReport) (r *Report, err error) {
	err = api.RequestJSON(&r, RouteEditReport(report), &RequestOptions{JSON: params})
	return
}

// Fetch a report by its ID.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/fetch_report_fetch_report
func (api *API) FetchReport(report ULID) (r *Report, err error) {
	err = api.RequestJSON(&r, RouteFetchReport(report), nil)
	return
}

type FetchReports struct {
	// Find reports aganist messages, servers, or users
	ContentID ULID
	// Find reports created by user
	AuthorID ULID
	// Report status to include in search
	Status ReportStatus
}

// Fetch all available reports.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/fetch_reports_fetch_reports
func (api *API) FetchReports(params *FetchReports) (a []*Report, err error) {
	v := url.Values{}
	if params != nil {
		if len(params.ContentID) != 0 {
			v.Set("content_id", string(params.ContentID))
		}
		if len(params.AuthorID) != 0 {
			v.Set("author_id", string(params.AuthorID))
		}
		if len(params.Status) != 0 {
			v.Set("status", string(params.Status))
		}
	}
	err = api.RequestJSON(&a, RouteFetchReports(), &RequestOptions{QueryValues: v})
	return
}

// Report a piece of content to the moderation team.
// content [required] - The content being reported.
// additionalContext [optional, pass empty] - Additional report description
// https://developers.revolt.chat/api/#tag/User-Safety/operation/report_content_report_content
func (api *API) ReportContent(content *ReportContent, additionalContext string) error {
	return api.RequestNone(RouteReportContent(), &RequestOptions{JSON: struct {
		Content           *ReportContent `json:"content"`
		AdditionalContext string         `json:"additional_context,omitempty"`
	}{content, additionalContext}})
}

// Fetch a snapshots for a given report.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/fetch_snapshots_fetch_snapshots
func (api *API) FetchSnapshots(report ULID) (a []*Snapshot, err error) {
	err = api.RequestJSON(&a, RouteFetchSnapshots(report), nil)
	return
}

// Create a new account strike
// user [required] - ID of reported user
// reason [required] - Attached reason
// https://developers.revolt.chat/api/#tag/User-Safety/operation/create_strike_create_strike
func (api *API) CreateStrike(user ULID, reason string) (s *Strike, err error) {
	err = api.RequestJSON(&s, RouteCreateStrike(), &RequestOptions{JSON: struct {
		UserID ULID   `json:"user_id"`
		Reason string `json:"reason"`
	}{user, reason}})
	return
}

// Fetch strikes for a user by their ID
// https://developers.revolt.chat/api/#tag/User-Safety/operation/fetch_strikes_fetch_strikes
func (api *API) FetchStrikes(user ULID) (a []*Strike, err error) {
	err = api.RequestJSON(&a, RouteFetchStrikes(user), nil)
	return
}

type EditStrike struct {
	// New attached reason
	Reason string `json:"reason,omitempty"`
}

// Edit a strike by its ID.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/edit_strike_edit_strike
func (api *API) EditStrike(strike ULID, params *EditStrike) error {
	return api.RequestNone(RouteEditStrike(strike), &RequestOptions{JSON: params})
}

// Delete a strike by its ID.
// https://developers.revolt.chat/api/#tag/User-Safety/operation/delete_strike_delete_strike
func (api *API) DeleteStrike(strike ULID) error {
	return api.RequestNone(RouteDeleteStrike(strike), nil)
}

type CreateAccount struct {
	// Valid email address
	Email string `json:"email"`
	// Password
	Password string `json:"password"`
	// Invite code
	Invite string `json:"invite,omitempty"`
	// Captcha verification code
	Captcha string `json:"captcha,omitempty"`
}

// Create a new account.
// https://developers.revolt.chat/api/#tag/Account/operation/create_account_create_account
func (api *API) CreateAccount(params *CreateAccount) error {
	return api.RequestNone(RouteCreateAccount(), &RequestOptions{JSON: params, Unauthenticated: true})
}

// Resend account creation verification email.
// email [required] - Email associated with the account
// captcha [optional, pass empty] - Captcha verification code
// https://developers.revolt.chat/api/#tag/Account/operation/resend_verification_resend_verification
func (api *API) ResendVerification(email, captcha string) error {
	return api.RequestNone(RouteResendVerification(), &RequestOptions{JSON: struct {
		Email   string `json:"email"`
		Captcha string `json:"captcha,omitempty"`
	}{email, captcha}, Unauthenticated: true})
}

// Schedule an account for deletion by confirming the received token.
// https://developers.revolt.chat/api/#tag/Account/operation/confirm_deletion_confirm_deletion
func (api *API) ConfirmAccountDeletion(token string) error {
	return api.RequestNone(RouteConfirmAccountDeletion(), &RequestOptions{JSON: struct {
		Token string `json:"token"`
	}{token}, Unauthenticated: true})
}

// Request to have an account deleted.
// https://developers.revolt.chat/api/#tag/Account/operation/delete_account_delete_account
func (api *API) DeleteAccount() error {
	return api.RequestNone(RouteDeleteAccount(), nil)
}

// Fetch account information from the current session.
// https://developers.revolt.chat/api/#tag/Account/operation/fetch_account_fetch_account
func (api *API) FetchAccount() (a *Account, err error) {
	err = api.RequestJSON(&a, RouteFetchAccount(), nil)
	return
}

// Disable an account.
// https://developers.revolt.chat/api/#tag/Account/operation/disable_account_disable_account
func (api *API) DisableAccount() error {
	return api.RequestNone(RouteDisableAccount(), nil)
}

// Change the current account password.
// password - New password
// currentPassword - Current password
// https://developers.revolt.chat/api/#tag/Account/operation/change_password_change_password
func (api *API) ChangePassword(password, currentPassword string) error {
	return api.RequestNone(RouteChangePassword(), &RequestOptions{
		JSON: struct {
			Password        string `json:"password"`
			CurrentPassword string `json:"current_password"`
		}{password, currentPassword},
	})
}

// Change the associated account email.
// email - Valid email address
// currentPassword - Current password
// https://developers.revolt.chat/api/#tag/Account/operation/change_email_change_email
func (api *API) ChangeEmail(email, currentPassword string) error {
	return api.RequestNone(RouteChangeEmail(), &RequestOptions{
		JSON: struct {
			Email           string `json:"email"`
			CurrentPassword string `json:"current_password"`
		}{email, currentPassword},
	})
}

// Verify an email address.
// ticket - May be nil
// https://developers.revolt.chat/api/#tag/Account/operation/verify_email_verify_email
func (api *API) VerifyEmail(code string) (ticket *MFATicket, err error) {
	var r *struct {
		Ticket *MFATicket `json:"ticket"`
	}
	err = api.RequestJSON(&r, RouteVerifyEmail(code), nil)
	if err == nil && r != nil {
		ticket = r.Ticket
	}
	return
}

// Send an email to reset account password.
// email [required] - Email associated with the account
// captcha [optional, pass empty string] - Captcha verification code
// https://developers.revolt.chat/api/#tag/Account/operation/send_password_reset_send_password_reset
func (api *API) SendPasswordReset(email, captcha string) error {
	return api.RequestNone(RouteSendPasswordReset(), &RequestOptions{JSON: struct {
		Email   string `json:"email"`
		Captcha string `json:"captcha,omitempty"`
	}{email, captcha}, Unauthenticated: true})
}

// Confirm password reset and change the password.
// token [required] - Reset token
// password [required] - New password
// https://developers.revolt.chat/api/#tag/Account/operation/password_reset_password_reset
func (api *API) PasswordReset(token, password string, removeSessions bool) error {
	return api.RequestNone(RoutePasswordReset(), &RequestOptions{JSON: struct {
		Token          string `json:"token"`
		Password       string `json:"password"`
		RemoveSessions bool   `json:"remove_sessions"`
	}{token, password, removeSessions}, Unauthenticated: true})
}

type MFAResponse struct {
	Password     string `json:"password,omitempty"`
	RecoveryCode string `json:"recovery_code,omitempty"`
	TOTPCode     string `json:"totp_code,omitempty"`
}

type Login struct {
	// Email
	Email string `json:"email,omitempty"`
	// Password
	Password string `json:"password,omitempty"`
	// Unvalidated or authorised MFA ticket
	// Used to resolve the correct account
	MFATicket string `json:"mfa_ticket,omitempty"`
	// MFA response
	MFAResponse *MFAResponse `json:"mfa_response,omitempty"`
	// Friendly name used for the session
	FriendlyName string `json:"friendly_name,omitempty"`
}

type LoginResponseType string

const (
	LoginResponseTypeSuccess  LoginResponseType = "Success"
	LoginResponseTypeMFA      LoginResponseType = "MFA"
	LoginResponseTypeDisabled LoginResponseType = "Disabled"
)

type MFAMethod string

const (
	MFAMethodPassword MFAMethod = "Password"
	MFAMethodRecovery MFAMethod = "Recovery"
	MFAMethodTOTP     MFAMethod = "Totp"
)

type LoginResponse struct {
	Result LoginResponseType `json:"result"`
	// Unique ID
	ID ULID `json:"_id"`
	// User ID
	UserID ULID   `json:"user_id"`
	Ticket string `json:"ticket"`
	// Session token
	Token string `json:"token"`
	// Display name
	Name           string               `json:"name"`
	Subscription   *WebPushSubscription `json:"subscription"`
	AllowedMethods []string             `json:"allowed_methods"`
}

// Login to an account.
// https://developers.revolt.chat/api/#tag/Session/operation/login_login
func (api *API) Login(params *Login) (lr *LoginResponse, err error) {
	err = api.RequestJSON(&lr, RouteLogin(), &RequestOptions{JSON: params, Unauthenticated: true})
	return
}

// Delete current session.
// https://developers.revolt.chat/api/#tag/Session/operation/logout_logout
func (api *API) Logout() error {
	return api.RequestNone(RouteLogout(), nil)
}

// Fetch all sessions associated with this account.
// https://developers.revolt.chat/api/#tag/Session/operation/fetch_all_fetch_all
func (api *API) FetchSessions() (s []*Session, err error) {
	err = api.RequestJSON(&s, RouteFetchSessions(), nil)
	return
}

// Delete all active sessions, optionally including current one.
// https://developers.revolt.chat/api/#tag/Session/operation/revoke_all_revoke_all
func (api *API) DeleteAllSessions(revokeSelf bool) error {
	v := url.Values{}
	if revokeSelf {
		v.Set("revoke_self", "true")
	} else {
		v.Set("revoke_self", "false")
	}
	return api.RequestNone(RouteDeleteAllSessions(), &RequestOptions{QueryValues: v})
}

// Delete a specific active session.
// https://developers.revolt.chat/api/#tag/Session/operation/revoke_revoke
func (api *API) RevokeSession(session string) error {
	return api.RequestNone(RouteRevokeSession(session), nil)
}

// Edit current session information.
// friendlyName [required] - Session friendly name
// https://developers.revolt.chat/api/#tag/Session/operation/edit_edit
func (api *API) EditSession(session, friendlyName string) (s *Session, err error) {
	err = api.RequestJSON(&s, RouteEditSession(session), &RequestOptions{
		JSON: struct {
			FriendlyName string `json:"friendly_name"`
		}{friendlyName},
	})
	return
}

// This will tell you whether the current account requires onboarding or whether you can continue to send requests as usual. You may skip calling this if you're restoring an existing session.
// Return value - Whether onboarding is required
// https://developers.revolt.chat/api/#tag/Onboarding/operation/hello_req
func (api *API) CheckOnboardingStatus() (onboarding bool, err error) {
	t := struct {
		Onboarding bool `json:"onboarding"`
	}{}
	err = api.RequestJSON(&t, RouteCheckOnboardingStatus(), nil)
	onboarding = t.Onboarding
	return
}

// This sets a new username, completes onboarding and allows a user to start using Revolt.
// username [required] - New username which will be used to identify the user on the platform
// https://developers.revolt.chat/api/#tag/Onboarding/operation/complete_req
func (api *API) CompleteOnboarding(username string) error {
	return api.RequestNone(RouteCompleteOnboarding(), &RequestOptions{
		JSON: struct {
			Username string `json:"username"`
		}{username},
	})
}

// Create a new MFA ticket or validate an existing one.
// https://developers.revolt.chat/api/#tag/MFA/operation/create_ticket_create_ticket
func (api *API) CreateMFATicket(params *MFAResponse) (t *MFATicket, err error) {
	err = api.RequestJSON(&t, RouteCreateMFATicket(), &RequestOptions{JSON: params})
	return
}

type MFAStatus struct {
	EmailOTP        bool `json:"email_otp"`
	TrustedHandover bool `json:"trusted_handover"`
	EmailMFA        bool `json:"email_mfa"`
	TOTPMFA         bool `json:"totp_mfa"`
	SecurityKeyMFA  bool `json:"security_key_mfa"`
	RecoveryActive  bool `json:"recovery_active"`
}

// Fetch MFA status of an account.
// https://developers.revolt.chat/api/#tag/MFA/operation/fetch_status_fetch_status
func (api *API) FetchMFAStatus() (s *MFAStatus, err error) {
	err = api.RequestJSON(&s, RouteFetchMFAStatus(), nil)
	return
}

// Fetch recovery codes for an account.
// https://developers.revolt.chat/api/#tag/MFA/operation/fetch_recovery_fetch_recovery
func (api *API) FetchRecoveryCodes() (a []string, err error) {
	err = api.RequestJSON(&a, RouteFetchRecoveryCodes(), nil)
	return
}

// Re-generate recovery codes for an account.
// https://developers.revolt.chat/api/#tag/MFA/operation/generate_recovery_generate_recovery
func (api *API) GenerateRecoveryCodes() (a []string, err error) {
	err = api.RequestJSON(&a, RouteGenerateRecoveryCodes(), nil)
	return
}

// Fetch available MFA methods.
// https://developers.revolt.chat/api/#tag/MFA/operation/get_mfa_methods_get_mfa_methods
func (api *API) GetMFAMethods() (a []MFAMethod, err error) {
	err = api.RequestJSON(&a, RouteGetMFAMethods(), nil)
	return
}

// Generate a new secret for TOTP.
// https://developers.revolt.chat/api/#tag/MFA/operation/totp_enable_totp_enable
func (api *API) EnableTOTP2FA(params *MFAResponse) error {
	return api.RequestNone(RouteEnableTOTP2FA(), &RequestOptions{JSON: params})
}

// Generate a new secret for TOTP.
// https://developers.revolt.chat/api/#tag/MFA/operation/totp_generate_secret_totp_generate_secret
func (api *API) GenerateTOTPSecret() (s string, err error) {
	t := struct {
		Secret string `json:"secret"`
	}{}
	err = api.RequestJSON(&t, RouteGenerateTOTPSecret(), nil)
	s = t.Secret
	return
}

// Disable TOTP 2FA for an account.
// https://developers.revolt.chat/api/#tag/MFA/operation/totp_disable_totp_disable
func (api *API) DisableTOTP2FA() error {
	return api.RequestNone(RouteDisableTOTP2FA(), nil)
}

// Fetch settings from server filtered by keys.
// This will return an object with the requested keys, each value is a `Setting` struct, the value is the previously uploaded data.
// keys [required] - Keys to fetch
// https://developers.revolt.chat/api/#tag/Sync/operation/get_settings_req
func (api *API) FetchSettings(keys []string) (m map[string]*Setting, err error) {
	err = api.RequestJSON(&m, RouteFetchSettings(), &RequestOptions{JSON: struct {
		Keys []string `json:"keys"`
	}{keys}})
	return
}

// Upload data to save to settings.
// timestamp [optional, pass zero] - Timestamp of settings change. Used to avoid feedback loops.
// https://developers.revolt.chat/api/#tag/Sync/operation/set_settings_req
func (api *API) SetSettings(timestamp int64, data map[string]string) error {
	v := url.Values{}
	if timestamp != 0 {
		v.Set("timestamp", strconv.Itoa(int(timestamp)))
	}
	return api.RequestNone(RouteSetSettings(), &RequestOptions{JSON: data})
}

// Fetch information about unread state on channels.
// https://developers.revolt.chat/api/#tag/Sync/operation/get_unreads_req
func (api *API) FetchUnreads() (a []*UnreadMessage, err error) {
	err = api.RequestJSON(&a, RouteFetchUnreads(), nil)
	return
}

// Create a new Web Push subscription.
// If an existing subscription exists on this session, it will be removed.
// https://developers.revolt.chat/api/#tag/Web-Push/operation/subscribe_req
func (api *API) PushSubscribe(params WebPushSubscription) error {
	return api.RequestNone(RoutePushSubscribe(), &RequestOptions{JSON: params})
}

// Remove the Web Push subscription associated with the current session.
// https://developers.revolt.chat/api/#tag/Web-Push/operation/unsubscribe_req
func (api *API) Unsubscribe() error {
	return api.RequestNone(RouteUnsubscribe(), nil)
}
