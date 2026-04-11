package blaze

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
)

// StringMatcher matches a string value.
type StringMatcher interface {
	Match(value string) bool
}

// EqualTo returns a StringMatcher that matches exactly.
func EqualTo(v string) StringMatcher {
	return &equalToMatcher{value: v}
}

type equalToMatcher struct{ value string }

func (m *equalToMatcher) Match(value string) bool { return value == m.value }

// Prefix returns a StringMatcher that matches a prefix.
func Prefix(v string) StringMatcher {
	return &prefixMatcher{value: v}
}

type prefixMatcher struct{ value string }

func (m *prefixMatcher) Match(value string) bool { return strings.HasPrefix(value, m.value) }

// Suffix returns a StringMatcher that matches a suffix.
func Suffix(v string) StringMatcher {
	return &suffixMatcher{value: v}
}

type suffixMatcher struct{ value string }

func (m *suffixMatcher) Match(value string) bool { return strings.HasSuffix(value, m.value) }

// Contains returns a StringMatcher that checks if the value contains the substring.
func Contains(v string) StringMatcher {
	return &containsMatcher{value: v}
}

type containsMatcher struct{ value string }

func (m *containsMatcher) Match(value string) bool { return strings.Contains(value, m.value) }

// MatchesRegex returns a StringMatcher that matches against a regex pattern.
func MatchesRegex(pattern string) StringMatcher {
	return &regexMatcher{re: regexp.MustCompile(pattern)}
}

type regexMatcher struct{ re *regexp.Regexp }

func (m *regexMatcher) Match(value string) bool { return m.re.MatchString(value) }

// BodyMatcher matches a request body.
type BodyMatcher interface {
	MatchBody(body []byte) bool
}

// EqualToBody returns a BodyMatcher that matches the body exactly.
func EqualToBody(expected []byte) BodyMatcher {
	return &equalToBodyMatcher{expected: expected}
}

type equalToBodyMatcher struct{ expected []byte }

func (m *equalToBodyMatcher) MatchBody(body []byte) bool { return bytes.Equal(body, m.expected) }

// ContainsString returns a BodyMatcher that checks if the body contains a substring.
func ContainsString(substr string) BodyMatcher {
	return &containsStringMatcher{substr: substr}
}

type containsStringMatcher struct{ substr string }

func (m *containsStringMatcher) MatchBody(body []byte) bool {
	return bytes.Contains(body, []byte(m.substr))
}

// AllOf returns a BodyMatcher that passes only when all provided matchers pass.
func AllOf(matchers ...BodyMatcher) BodyMatcher {
	return &allOfMatcher{matchers: matchers}
}

type allOfMatcher struct{ matchers []BodyMatcher }

func (m *allOfMatcher) MatchBody(body []byte) bool {
	for _, matcher := range m.matchers {
		if !matcher.MatchBody(body) {
			return false
		}
	}
	return true
}

// EqualToJSON returns a BodyMatcher that compares JSON structurally (ignoring key order).
// By default, extra fields in the actual body cause a mismatch. Use IgnoreExtraFields()
// to allow additional fields.
func EqualToJSON(jsonStr string) *equalToJSONMatcher {
	return &equalToJSONMatcher{expected: jsonStr}
}

type equalToJSONMatcher struct {
	expected    string
	ignoreExtra bool
}

// IgnoreExtraFields allows the actual body to contain fields not present in the expected JSON.
func (m *equalToJSONMatcher) IgnoreExtraFields() *equalToJSONMatcher {
	m.ignoreExtra = true
	return m
}

func (m *equalToJSONMatcher) MatchBody(body []byte) bool {
	var expected, actual any
	if err := json.Unmarshal([]byte(m.expected), &expected); err != nil {
		return false
	}
	if err := json.Unmarshal(body, &actual); err != nil {
		return false
	}
	if m.ignoreExtra {
		return jsonContains(expected, actual)
	}
	return reflect.DeepEqual(expected, actual)
}

// jsonContains checks that all fields in expected exist in actual with matching values.
// Extra fields in actual are allowed.
func jsonContains(expected, actual any) bool {
	switch e := expected.(type) {
	case map[string]any:
		a, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for k, ev := range e {
			av, exists := a[k]
			if !exists {
				return false
			}
			if !jsonContains(ev, av) {
				return false
			}
		}
		return true
	case []any:
		a, ok := actual.([]any)
		if !ok || len(a) != len(e) {
			return false
		}
		for i := range e {
			if !jsonContains(e[i], a[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(expected, actual)
	}
}

// MatchesJSONPath returns a BodyMatcher that extracts a value at the given JSONPath
// and checks it against a StringMatcher.
// Path syntax: "$.field", "$.nested.field", "$.array.0.field".
func MatchesJSONPath(path string, matcher StringMatcher) BodyMatcher {
	return &jsonPathMatcher{path: path, matcher: matcher}
}

type jsonPathMatcher struct {
	path    string
	matcher StringMatcher
}

func (m *jsonPathMatcher) MatchBody(body []byte) bool {
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return false
	}
	value := jsonPathLookup(data, m.path)
	return m.matcher.Match(value)
}
