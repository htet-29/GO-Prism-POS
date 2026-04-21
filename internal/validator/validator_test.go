package validator

import "testing"

func TestValidator_New(t *testing.T) {
	v := New()
	if v.Errors == nil {
		t.Fatal("expected Errors map to be initialized, got nil")
	}

	if len(v.Errors) != 0 {
		t.Errorf("expected empty Errors map, got length:  %d", len(v.Errors))
	}
}

func TestValidator_Valid(t *testing.T) {
	v := New()

	if !v.Valid() {
		t.Error("expected new validator to be valid")
	}

	v.AddError("email", "invalid email address")

	if v.Valid() {
		t.Error("expected validator with errors to be invalid")
	}
}

func TestValidator_AddError(t *testing.T) {
	v := New()

	v.AddError("field", "first error message")
	if msg, exists := v.Errors["field"]; !exists || msg != "first error message" {
		t.Errorf("expected 'first error message', got '%s'", msg)
	}

	// Adding an error to the same key should not overwrite the first message
	v.AddError("field", "second error message")
	if msg := v.Errors["field"]; msg != "first error message" {
		t.Errorf("expected 'first error message' to remain, but got '%s'", msg)
	}
}

func TestValidator_Check(t *testing.T) {
	v := New()

	// Check with ok == true should not add an error
	v.Check(true, "field1", "error 1")
	if !v.Valid() {
		t.Error("expected validator to remain valid when Check receives true")
	}

	// Check with ok == false should add an error
	v.Check(false, "field2", "error 2")
	if v.Valid() {
		t.Error("expected validator to be invalid when Check receives false")
	}
	if msg := v.Errors["field2"]; msg != "error 2" {
		t.Errorf("expected 'error 2', got '%s'", msg)
	}
}

func TestPermittedValue(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		if !PermittedValue("apple", "apple", "banana", "orange") {
			t.Error("expected 'apple' to be permitted")
		}
		if PermittedValue("grape", "apple", "banana", "orange") {
			t.Error("expected 'grape' to NOT be permitted")
		}
	})

	t.Run("ints", func(t *testing.T) {
		if !PermittedValue(2, 1, 2, 3) {
			t.Error("expected 2 to be permitted")
		}
		if PermittedValue(4, 1, 2, 3) {
			t.Error("expected 4 to NOT be permitted")
		}
	})
}

func TestMatches_EmailRX(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid standard email", "alice@example.com", true},
		{"valid with subdomain", "alice@mail.example.com", true},
		{"valid with plus symbol", "bob+tag@example.com", true},
		{"missing at symbol", "aliceexample.com", false},
		{"missing domain", "alice@", false},
		{"invalid characters", "alice@exa mple.com", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(tt.email, EmailRX); got != tt.valid {
				t.Errorf("Matches(%q, EmailRX) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		if !Unique([]string{"a", "b", "c"}) {
			t.Error("expected slice with unique items to return true")
		}
		if Unique([]string{"a", "b", "a"}) {
			t.Error("expected slice with duplicate items to return false")
		}
	})

	t.Run("ints", func(t *testing.T) {
		if !Unique([]int{1, 2, 3, 4, 5}) {
			t.Error("expected slice with unique ints to return true")
		}
		if Unique([]int{1, 2, 3, 2, 5}) {
			t.Error("expected slice with duplicate ints to return false")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		if !Unique([]string{}) {
			t.Error("expected empty slice to return true")
		}
	})
}
