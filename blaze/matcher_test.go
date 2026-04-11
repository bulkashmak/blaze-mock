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

func TestAllOf(t *testing.T) {
	body := []byte(`{"status": "active", "amount": 100, "customer": {"name": "alice"}}`)

	t.Run("passes when all matchers pass", func(t *testing.T) {
		m := AllOf(
			MatchesJSONPath("$.status", EqualTo("active")),
			MatchesJSONPath("$.customer.name", EqualTo("alice")),
		)
		if !m.MatchBody(body) {
			t.Error("expected match")
		}
	})

	t.Run("fails when one matcher fails", func(t *testing.T) {
		m := AllOf(
			MatchesJSONPath("$.status", EqualTo("active")),
			MatchesJSONPath("$.customer.name", EqualTo("bob")),
		)
		if m.MatchBody(body) {
			t.Error("expected no match")
		}
	})

	t.Run("passes with single matcher", func(t *testing.T) {
		m := AllOf(MatchesJSONPath("$.status", EqualTo("active")))
		if !m.MatchBody(body) {
			t.Error("expected match")
		}
	})

	t.Run("passes with zero matchers", func(t *testing.T) {
		m := AllOf()
		if !m.MatchBody(body) {
			t.Error("expected match with no matchers")
		}
	})

	t.Run("combines different matcher types", func(t *testing.T) {
		m := AllOf(
			MatchesJSONPath("$.status", EqualTo("active")),
			ContainsString(`"amount"`),
			EqualToJSON(`{"status": "active"}`).IgnoreExtraFields(),
		)
		if !m.MatchBody(body) {
			t.Error("expected match combining JSONPath, ContainsString, and EqualToJSON")
		}
	})
}

func TestEqualToJSON(t *testing.T) {
	t.Run("matches identical JSON", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice", "age": 30}`)
		if !m.MatchBody([]byte(`{"name": "alice", "age": 30}`)) {
			t.Error("expected match")
		}
	})

	t.Run("matches with different key order", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice", "age": 30}`)
		if !m.MatchBody([]byte(`{"age": 30, "name": "alice"}`)) {
			t.Error("expected match regardless of key order")
		}
	})

	t.Run("rejects extra fields by default", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice"}`)
		if m.MatchBody([]byte(`{"name": "alice", "age": 30}`)) {
			t.Error("expected no match when actual has extra fields")
		}
	})

	t.Run("rejects missing fields", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice", "age": 30}`)
		if m.MatchBody([]byte(`{"name": "alice"}`)) {
			t.Error("expected no match when actual is missing fields")
		}
	})

	t.Run("rejects different values", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice"}`)
		if m.MatchBody([]byte(`{"name": "bob"}`)) {
			t.Error("expected no match for different values")
		}
	})

	t.Run("matches nested objects", func(t *testing.T) {
		m := EqualToJSON(`{"user": {"name": "alice", "role": "admin"}}`)
		if !m.MatchBody([]byte(`{"user": {"role": "admin", "name": "alice"}}`)) {
			t.Error("expected match for nested objects with different key order")
		}
	})

	t.Run("matches arrays", func(t *testing.T) {
		m := EqualToJSON(`{"tags": ["a", "b"]}`)
		if !m.MatchBody([]byte(`{"tags": ["a", "b"]}`)) {
			t.Error("expected match for equal arrays")
		}
		if m.MatchBody([]byte(`{"tags": ["b", "a"]}`)) {
			t.Error("expected no match for different array order")
		}
	})

	t.Run("rejects invalid JSON body", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice"}`)
		if m.MatchBody([]byte(`not json`)) {
			t.Error("expected no match for invalid JSON")
		}
	})
}

func TestEqualToJSON_IgnoreExtraFields(t *testing.T) {
	t.Run("allows extra fields", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice"}`).IgnoreExtraFields()
		if !m.MatchBody([]byte(`{"name": "alice", "age": 30, "role": "admin"}`)) {
			t.Error("expected match when ignoring extra fields")
		}
	})

	t.Run("still requires all expected fields", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice", "age": 30}`).IgnoreExtraFields()
		if m.MatchBody([]byte(`{"name": "alice"}`)) {
			t.Error("expected no match when expected fields are missing")
		}
	})

	t.Run("works with nested objects", func(t *testing.T) {
		m := EqualToJSON(`{"user": {"name": "alice"}}`).IgnoreExtraFields()
		if !m.MatchBody([]byte(`{"user": {"name": "alice", "age": 30}, "extra": true}`)) {
			t.Error("expected match when nested objects have extra fields")
		}
	})

	t.Run("still rejects wrong values", func(t *testing.T) {
		m := EqualToJSON(`{"name": "alice"}`).IgnoreExtraFields()
		if m.MatchBody([]byte(`{"name": "bob", "age": 30}`)) {
			t.Error("expected no match for different values")
		}
	})
}

func TestMatchesJSONPath(t *testing.T) {
	body := []byte(`{"user": {"name": "alice", "age": 30}, "status": "active"}`)

	t.Run("matches top-level field", func(t *testing.T) {
		m := MatchesJSONPath("$.status", EqualTo("active"))
		if !m.MatchBody(body) {
			t.Error("expected match")
		}
	})

	t.Run("matches nested field", func(t *testing.T) {
		m := MatchesJSONPath("$.user.name", EqualTo("alice"))
		if !m.MatchBody(body) {
			t.Error("expected match")
		}
	})

	t.Run("matches with Contains", func(t *testing.T) {
		m := MatchesJSONPath("$.user.name", Contains("lic"))
		if !m.MatchBody(body) {
			t.Error("expected match")
		}
	})

	t.Run("matches numeric field as string", func(t *testing.T) {
		m := MatchesJSONPath("$.user.age", EqualTo("30"))
		if !m.MatchBody(body) {
			t.Error("expected match for numeric value compared as string")
		}
	})

	t.Run("rejects wrong value", func(t *testing.T) {
		m := MatchesJSONPath("$.user.name", EqualTo("bob"))
		if m.MatchBody(body) {
			t.Error("expected no match")
		}
	})

	t.Run("rejects missing path", func(t *testing.T) {
		m := MatchesJSONPath("$.nonexistent", EqualTo("anything"))
		if m.MatchBody(body) {
			t.Error("expected no match for missing path")
		}
	})

	t.Run("rejects invalid JSON", func(t *testing.T) {
		m := MatchesJSONPath("$.field", EqualTo("value"))
		if m.MatchBody([]byte(`not json`)) {
			t.Error("expected no match for invalid JSON")
		}
	})

	t.Run("matches array element", func(t *testing.T) {
		m := MatchesJSONPath("$.items.0.id", EqualTo("first"))
		body := []byte(`{"items": [{"id": "first"}, {"id": "second"}]}`)
		if !m.MatchBody(body) {
			t.Error("expected match for array element")
		}
	})

	t.Run("matches with regex", func(t *testing.T) {
		m := MatchesJSONPath("$.user.name", MatchesRegex(`^ali`))
		if !m.MatchBody(body) {
			t.Error("expected match with regex matcher")
		}
	})
}
