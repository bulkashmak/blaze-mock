package blaze

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type pathParamsKey struct{}

// pathMatcher matches URL paths, supporting {param} placeholders.
type pathMatcher struct {
	re         *regexp.Regexp
	paramNames []string
}

// compilePath converts a path pattern like "/api/payments/{id}" into a regex-based matcher.
func compilePath(pattern string) *pathMatcher {
	var paramNames []string
	regexPattern := "^"
	remaining := pattern

	for {
		openIdx := strings.Index(remaining, "{")
		if openIdx == -1 {
			regexPattern += regexp.QuoteMeta(remaining)
			break
		}
		closeIdx := strings.Index(remaining[openIdx:], "}")
		if closeIdx == -1 {
			regexPattern += regexp.QuoteMeta(remaining)
			break
		}
		closeIdx += openIdx

		regexPattern += regexp.QuoteMeta(remaining[:openIdx])
		paramName := remaining[openIdx+1 : closeIdx]
		paramNames = append(paramNames, paramName)
		regexPattern += "([^/]+)"
		remaining = remaining[closeIdx+1:]
	}

	regexPattern += "$"

	return &pathMatcher{
		re:         regexp.MustCompile(regexPattern),
		paramNames: paramNames,
	}
}

func (m *pathMatcher) Match(value string) bool {
	return m.re.MatchString(value)
}

// extractParams returns path parameter values from the given path, or nil if no match.
func (m *pathMatcher) extractParams(path string) map[string]string {
	matches := m.re.FindStringSubmatch(path)
	if matches == nil {
		return nil
	}
	params := make(map[string]string, len(m.paramNames))
	for i, name := range m.paramNames {
		params[name] = matches[i+1]
	}
	return params
}

// withPathParams attaches path parameters to the request context.
func withPathParams(r *http.Request, params map[string]string) *http.Request {
	if len(params) == 0 {
		return r
	}
	ctx := context.WithValue(r.Context(), pathParamsKey{}, params)
	return r.WithContext(ctx)
}

// PathParam extracts a named path parameter from the request.
func PathParam(r *http.Request, name string) string {
	params, ok := r.Context().Value(pathParamsKey{}).(map[string]string)
	if !ok {
		return ""
	}
	return params[name]
}
