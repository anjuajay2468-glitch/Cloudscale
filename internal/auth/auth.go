package auth

import "crypto/subtle"

type Role string

const (
	RoleReader Role = "reader"
	RoleAdmin  Role = "admin"
)

type APIKey struct {
	Key  []byte
	Role Role
}

type Authenticator struct {
	keys map[string]APIKey
}

func New(apiKey string) *Authenticator {
	authenticator := &Authenticator{
		keys: make(map[string]APIKey),
	}

	if apiKey != "" {
		authenticator.keys["default"] = APIKey{
			Key:  []byte(apiKey),
			Role: RoleAdmin,
		}
	}

	return authenticator
}

func NewWithKeys(keys map[string]APIKey) *Authenticator {
	authenticator := &Authenticator{
		keys: make(map[string]APIKey, len(keys)),
	}

	for name, key := range keys {
		keyCopy := make([]byte, len(key.Key))
		copy(keyCopy, key.Key)

		authenticator.keys[name] = APIKey{
			Key:  keyCopy,
			Role: key.Role,
		}
	}

	return authenticator
}

func (a *Authenticator) Enabled() bool {
	return len(a.keys) > 0
}

func (a *Authenticator) Validate(provided string) bool {
	_, ok := a.Authenticate(provided)
	return ok
}

func (a *Authenticator) Authenticate(provided string) (Role, bool) {
	if !a.Enabled() {
		return "", false
	}

	providedBytes := []byte(provided)

	for _, key := range a.keys {
		if len(providedBytes) != len(key.Key) {
			continue
		}

		if subtle.ConstantTimeCompare(
			providedBytes,
			key.Key,
		) == 1 {
			return key.Role, true
		}
	}

	return "", false
}

func (a *Authenticator) Authorize(role Role, method string) bool {
	switch role {
	case RoleAdmin:
		return method == "GET" ||
			method == "PUT" ||
			method == "DELETE"

	case RoleReader:
		return method == "GET"

	default:
		return false
	}
}
