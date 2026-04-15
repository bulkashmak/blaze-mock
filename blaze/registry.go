package blaze

import (
	"net/http"
	"strings"
	"sync"
)

// StubRegistry is a thread-safe in-memory store for stubs.
type StubRegistry struct {
	mu    sync.RWMutex
	stubs []Stub
}

func (r *StubRegistry) Add(s Stub) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = append(r.stubs, s)
	return s.ID
}

func (r *StubRegistry) Update(id string, s Stub) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.stubs {
		if r.stubs[i].ID == id {
			s.ID = id
			r.stubs[i] = s
			return id
		}
	}
	return ""
}

func (r *StubRegistry) Remove(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, s := range r.stubs {
		if s.ID == id {
			r.stubs = append(r.stubs[:i], r.stubs[i+1:]...)
			return true
		}
	}
	return false
}

func (r *StubRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stubs = nil
}

func (r *StubRegistry) Get(id string) *Stub {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.stubs {
		if r.stubs[i].ID == id {
			return &r.stubs[i]
		}
	}
	return nil
}

func (r *StubRegistry) List() []Stub {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Stub, len(r.stubs))
	copy(out, r.stubs)
	return out
}

// Match returns the first stub that matches the request, along with extracted path parameters.
func (r *StubRegistry) Match(req *http.Request, body []byte) (*Stub, map[string]string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for i := range r.stubs {
		s := &r.stubs[i]
		params := matchStub(s, req, body)
		if params != nil {
			return s, params
		}
	}
	return nil, nil
}

// matchStub checks if a stub matches the request. Returns path params (empty map if matched but no params), or nil if no match.
func matchStub(s *Stub, req *http.Request, body []byte) map[string]string {
	rm := &s.Request

	if !strings.EqualFold(rm.Method, req.Method) {
		return nil
	}

	pm, ok := rm.Path.(*pathMatcher)
	if !ok {
		if !rm.Path.Match(req.URL.Path) {
			return nil
		}
	} else {
		params := pm.extractParams(req.URL.Path)
		if params == nil {
			return nil
		}
		// Check remaining matchers before returning
		if !matchHeaders(rm.Headers, req) {
			return nil
		}
		if !matchQueryParams(rm.QueryParams, req) {
			return nil
		}
		if rm.Body != nil && !rm.Body.MatchBody(body) {
			return nil
		}
		return params
	}

	if !matchHeaders(rm.Headers, req) {
		return nil
	}
	if !matchQueryParams(rm.QueryParams, req) {
		return nil
	}
	if rm.Body != nil && !rm.Body.MatchBody(body) {
		return nil
	}

	return make(map[string]string)
}

func matchHeaders(matchers map[string]StringMatcher, req *http.Request) bool {
	for name, matcher := range matchers {
		value := req.Header.Get(name)
		if !matcher.Match(value) {
			return false
		}
	}
	return true
}

func matchQueryParams(matchers map[string]StringMatcher, req *http.Request) bool {
	query := req.URL.Query()
	for name, matcher := range matchers {
		value := query.Get(name)
		if !matcher.Match(value) {
			return false
		}
	}
	return true
}
