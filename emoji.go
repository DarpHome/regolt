package regolt

import (
	"encoding/json"
	"net/url"
	"strings"
)

type Emoji struct {
	ID    ULID
	Emoji string
}

func NewUnicodeEmoji(emoji string) Emoji {
	return Emoji{Emoji: emoji}
}

func (e Emoji) EncodeFP() string {
	if e.ID != "" {
		return e.ID.EncodeFP()
	}
	return url.PathEscape(e.Emoji)
}

func (e *Emoji) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if strings.Contains(s[:1], "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		// custom emoji
		e.ID = ULID(s)
	} else {
		// unicode emoji
		e.Emoji = s
	}
	return nil
}
