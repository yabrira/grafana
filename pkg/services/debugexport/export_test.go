package debugexport

import "testing"

func TestHashPasswordDeterministic(t *testing.T) {
	first := HashPassword("admin")
	second := HashPassword("admin")
	if first != second {
		t.Fatalf("expected stable hash, got %s and %s", first, second)
	}
}

func TestPrimaryLogin(t *testing.T) {
	users := []*UserRow{{Login: "admin"}}
	if got := PrimaryLogin(users); got != "admin" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultAdminTokenPresent(t *testing.T) {
	if DefaultAdminToken() == "" {
		t.Fatal("expected bootstrap token")
	}
}
