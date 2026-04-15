package blaze

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestStringMatcherJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		matcher StringMatcher
		json    matcherJSON
	}{
		{"EqualTo", EqualTo("hello"), matcherJSON{EqualTo: ptr("hello")}},
		{"Prefix", Prefix("hel"), matcherJSON{Prefix: ptr("hel")}},
		{"Suffix", Suffix("llo"), matcherJSON{Suffix: ptr("llo")}},
		{"Contains", Contains("ell"), matcherJSON{Contains: ptr("ell")}},
		{"MatchesRegex", MatchesRegex("^h.*o$"), matcherJSON{Matches: ptr("^h.*o$")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Forward: matcher → JSON
			got := stringMatcherToJSON(tt.matcher)
			gotBytes, _ := json.Marshal(got)
			wantBytes, _ := json.Marshal(tt.json)
			if string(gotBytes) != string(wantBytes) {
				t.Errorf("stringMatcherToJSON: got %s, want %s", gotBytes, wantBytes)
			}

			// Reverse: JSON → matcher
			restored, err := jsonToStringMatcher(tt.json)
			if err != nil {
				t.Fatalf("jsonToStringMatcher: %v", err)
			}
			if !restored.Match("hello") == tt.matcher.Match("hello") {
				t.Errorf("round-trip mismatch on 'hello'")
			}
		})
	}
}

func TestBodyMatcherJSONRoundTrip(t *testing.T) {
	t.Run("EqualToBody", func(t *testing.T) {
		m := EqualToBody([]byte("exact"))
		dto := bodyMatcherToJSON(m)
		if dto.EqualTo == nil || *dto.EqualTo != "exact" {
			t.Fatalf("expected equalTo='exact', got %+v", dto)
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte("exact")) {
			t.Error("restored matcher should match 'exact'")
		}
		if restored.MatchBody([]byte("other")) {
			t.Error("restored matcher should not match 'other'")
		}
	})

	t.Run("ContainsString", func(t *testing.T) {
		m := ContainsString("needle")
		dto := bodyMatcherToJSON(m)
		if dto.Contains == nil || *dto.Contains != "needle" {
			t.Fatalf("expected contains='needle', got %+v", dto)
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte("has needle inside")) {
			t.Error("restored matcher should match")
		}
	})

	t.Run("EqualToJSON", func(t *testing.T) {
		m := EqualToJSON(`{"a":1}`)
		dto := bodyMatcherToJSON(m)
		if dto.EqualToJSON == nil || *dto.EqualToJSON != `{"a":1}` {
			t.Fatalf("expected equalToJson, got %+v", dto)
		}
		if dto.IgnoreExtra {
			t.Error("ignoreExtra should be false")
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte(`{"a":1}`)) {
			t.Error("restored matcher should match")
		}
	})

	t.Run("EqualToJSON_IgnoreExtra", func(t *testing.T) {
		m := EqualToJSON(`{"a":1}`).IgnoreExtraFields()
		dto := bodyMatcherToJSON(m)
		if !dto.IgnoreExtra {
			t.Error("ignoreExtra should be true")
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte(`{"a":1,"b":2}`)) {
			t.Error("restored matcher should match with extra fields")
		}
	})

	t.Run("MatchesJSONPath", func(t *testing.T) {
		m := MatchesJSONPath("$.name", EqualTo("alice"))
		dto := bodyMatcherToJSON(m)
		if dto.MatchesJSONPath == nil {
			t.Fatal("expected matchesJsonPath")
		}
		if dto.MatchesJSONPath.Expression != "$.name" {
			t.Errorf("expected expression '$.name', got %s", dto.MatchesJSONPath.Expression)
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte(`{"name":"alice"}`)) {
			t.Error("restored matcher should match")
		}
	})

	t.Run("AllOf", func(t *testing.T) {
		m := AllOf(ContainsString("a"), ContainsString("b"))
		dto := bodyMatcherToJSON(m)
		if len(dto.AllOf) != 2 {
			t.Fatalf("expected 2 allOf items, got %d", len(dto.AllOf))
		}
		restored, err := jsonToBodyMatcher(*dto)
		if err != nil {
			t.Fatal(err)
		}
		if !restored.MatchBody([]byte("ab")) {
			t.Error("restored should match 'ab'")
		}
		if restored.MatchBody([]byte("a")) {
			t.Error("restored should not match 'a' alone")
		}
	})
}

func TestStubDTOCodeStub(t *testing.T) {
	stub := Stub{
		ID: "code-1",
		Request: RequestMatcher{
			Method: "POST",
			Path:   compilePath("/api/test"),
		},
		Response: ResponseDef{
			BodyFunc: func(r *http.Request) (int, map[string]string, []byte, error) {
				return 200, nil, nil, nil
			},
		},
	}

	dto := stubToDTO(stub)
	if dto.Type != "code" {
		t.Errorf("expected type='code', got %q", dto.Type)
	}
	if dto.Response.Description != "dynamic (Go func)" {
		t.Errorf("expected description='dynamic (Go func)', got %q", dto.Response.Description)
	}
	if dto.Response.Status != 0 {
		t.Errorf("code stub should not expose status, got %d", dto.Response.Status)
	}
}

func TestInvalidMatcherJSON(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		_, err := jsonToStringMatcher(matcherJSON{})
		if err == nil {
			t.Error("expected error for empty matcher")
		}
	})

	t.Run("multiple fields", func(t *testing.T) {
		_, err := jsonToStringMatcher(matcherJSON{EqualTo: ptr("a"), Prefix: ptr("b")})
		if err == nil {
			t.Error("expected error for multiple fields")
		}
	})

	t.Run("empty body matcher", func(t *testing.T) {
		_, err := jsonToBodyMatcher(bodyMatcherJSON{})
		if err == nil {
			t.Error("expected error for empty body matcher")
		}
	})
}

func TestDtoToStubValidation(t *testing.T) {
	tests := []struct {
		name string
		dto  stubDTO
	}{
		{"missing method", stubDTO{Request: requestDTO{Path: "/a"}, Response: responseDTO{Status: 200}}},
		{"missing path", stubDTO{Request: requestDTO{Method: "GET"}, Response: responseDTO{Status: 200}}},
		{"missing status", stubDTO{Request: requestDTO{Method: "GET", Path: "/a"}, Response: responseDTO{}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dtoToStub(tt.dto)
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func ptr(s string) *string { return &s }
