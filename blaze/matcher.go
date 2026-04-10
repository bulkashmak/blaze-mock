package blaze

import (
	"bytes"
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
