package blaze

import "testing"

func TestEqualTo(t *testing.T) {
	m := EqualTo("hello")
	if !m.Match("hello") {
		t.Error("expected match")
	}
	if m.Match("world") {
		t.Error("expected no match")
	}
}

func TestPrefix(t *testing.T) {
	m := Prefix("/api")
	if !m.Match("/api/users") {
		t.Error("expected match")
	}
	if m.Match("/web/api") {
		t.Error("expected no match")
	}
}

func TestSuffix(t *testing.T) {
	m := Suffix(".json")
	if !m.Match("data.json") {
		t.Error("expected match")
	}
	if m.Match("data.xml") {
		t.Error("expected no match")
	}
}

func TestContains(t *testing.T) {
	m := Contains("needle")
	if !m.Match("haystackneedlehaystack") {
		t.Error("expected match")
	}
	if m.Match("haystack") {
		t.Error("expected no match")
	}
}

func TestMatchesRegex(t *testing.T) {
	m := MatchesRegex(`^\d{3}$`)
	if !m.Match("123") {
		t.Error("expected match")
	}
	if m.Match("1234") {
		t.Error("expected no match")
	}
}

func TestEqualToBody(t *testing.T) {
	m := EqualToBody([]byte("exact"))
	if !m.MatchBody([]byte("exact")) {
		t.Error("expected match")
	}
	if m.MatchBody([]byte("other")) {
		t.Error("expected no match")
	}
}

func TestContainsString(t *testing.T) {
	m := ContainsString("amount")
	if !m.MatchBody([]byte(`{"amount": 100}`)) {
		t.Error("expected match")
	}
	if m.MatchBody([]byte(`{"price": 100}`)) {
		t.Error("expected no match")
	}
}
