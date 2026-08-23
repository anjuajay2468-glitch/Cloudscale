package auth

import "testing"

func TestValidateCorrectAPIKey(t *testing.T) {
	authenticator := New("cloudscale-secret-key")

	if !authenticator.Validate("cloudscale-secret-key") {
		t.Fatal("expected correct API key to be accepted")
	}
}

func TestValidateIncorrectAPIKey(t *testing.T) {
	authenticator := New("cloudscale-secret-key")

	if authenticator.Validate("wrong-key") {
		t.Fatal("expected incorrect API key to be rejected")
	}
}

func TestValidateEmptyAPIKey(t *testing.T) {
	authenticator := New("")

	if authenticator.Enabled() {
		t.Fatal("expected authentication to be disabled")
	}

	if authenticator.Validate("anything") {
		t.Fatal("authentication should reject when no API key is configured")
	}
}

func TestReaderAuthorization(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"reader": {
			Key:  []byte("reader-key"),
			Role: RoleReader,
		},
	})

	role, ok := authenticator.Authenticate("reader-key")

	if !ok {
		t.Fatal("expected reader key to authenticate")
	}

	if role != RoleReader {
		t.Fatalf("expected reader role, got %s", role)
	}

	if !authenticator.Authorize(role, "GET") {
		t.Fatal("reader should be allowed to GET")
	}

	if authenticator.Authorize(role, "PUT") {
		t.Fatal("reader should not be allowed to PUT")
	}

	if authenticator.Authorize(role, "DELETE") {
		t.Fatal("reader should not be allowed to DELETE")
	}
}

func TestAdminAuthorization(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"admin": {
			Key:  []byte("admin-key"),
			Role: RoleAdmin,
		},
	})

	role, ok := authenticator.Authenticate("admin-key")

	if !ok {
		t.Fatal("expected admin key to authenticate")
	}

	if role != RoleAdmin {
		t.Fatalf("expected admin role, got %s", role)
	}

	for _, method := range []string{
		"GET",
		"PUT",
		"DELETE",
	} {
		if !authenticator.Authorize(role, method) {
			t.Fatalf(
				"admin should be allowed to %s",
				method,
			)
		}
	}
}

func TestUnknownKey(t *testing.T) {
	authenticator := NewWithKeys(map[string]APIKey{
		"reader": {
			Key:  []byte("reader-key"),
			Role: RoleReader,
		},
	})

	_, ok := authenticator.Authenticate("wrong-key")

	if ok {
		t.Fatal("expected unknown key to be rejected")
	}
}
