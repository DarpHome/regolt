package regolt

import "net/url"

type ULID string

func (id ULID) EncodeFP() string {
	return url.PathEscape(string(id))
}

// !             |-- Revolt API version
// !             |
// !             | |-- Regolt major version
// !             | |
// !             | | |-- Regolt minor version
const Version = "1.1.0"
