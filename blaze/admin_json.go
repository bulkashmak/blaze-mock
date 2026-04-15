package blaze

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// stubDTO is the JSON representation of a stub.
type stubDTO struct {
	ID       string      `json:"id,omitempty"`
	Type     string      `json:"type,omitempty"`
	Request  requestDTO  `json:"request"`
	Response responseDTO `json:"response"`
}

type requestDTO struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Headers     map[string]matcherJSON `json:"headers,omitempty"`
	QueryParams map[string]matcherJSON `json:"queryParams,omitempty"`
	Body        *bodyMatcherJSON       `json:"body,omitempty"`
}

type responseDTO struct {
	Status      int               `json:"status"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	BodyFile    string            `json:"bodyFile,omitempty"`
	Description string            `json:"description,omitempty"`
}

type matcherJSON struct {
	EqualTo  *string `json:"equalTo,omitempty"`
	Prefix   *string `json:"prefix,omitempty"`
	Suffix   *string `json:"suffix,omitempty"`
	Contains *string `json:"contains,omitempty"`
	Matches  *string `json:"matches,omitempty"`
}

type bodyMatcherJSON struct {
	EqualTo         *string          `json:"equalTo,omitempty"`
	Contains        *string          `json:"contains,omitempty"`
	EqualToJSON     *string          `json:"equalToJson,omitempty"`
	IgnoreExtra     bool             `json:"ignoreExtraFields,omitempty"`
	MatchesJSONPath *jsonPathJSON    `json:"matchesJsonPath,omitempty"`
	AllOf           []bodyMatcherJSON `json:"allOf,omitempty"`
}

type jsonPathJSON struct {
	Expression string      `json:"expression"`
	Matcher    matcherJSON `json:"matcher"`
}

// stubToDTO converts an internal Stub to its JSON representation.
func stubToDTO(s Stub) stubDTO {
	dto := stubDTO{
		ID:      s.ID,
		Request: requestMatcherToDTO(s.Request),
	}

	if s.Response.BodyFunc != nil {
		dto.Type = "code"
		dto.Response = responseDTO{
			Description: "dynamic (Go func)",
		}
	} else {
		dto.Response = responsDefToDTO(s.Response)
	}

	return dto
}

func requestMatcherToDTO(rm RequestMatcher) requestDTO {
	dto := requestDTO{
		Method: rm.Method,
	}

	if pm, ok := rm.Path.(*pathMatcher); ok {
		dto.Path = pm.pattern
	} else {
		dto.Path = fmt.Sprintf("%s", rm.Path)
	}

	if len(rm.Headers) > 0 {
		dto.Headers = make(map[string]matcherJSON, len(rm.Headers))
		for k, v := range rm.Headers {
			dto.Headers[k] = stringMatcherToJSON(v)
		}
	}

	if len(rm.QueryParams) > 0 {
		dto.QueryParams = make(map[string]matcherJSON, len(rm.QueryParams))
		for k, v := range rm.QueryParams {
			dto.QueryParams[k] = stringMatcherToJSON(v)
		}
	}

	if rm.Body != nil {
		dto.Body = bodyMatcherToJSON(rm.Body)
	}

	return dto
}

func responsDefToDTO(rd ResponseDef) responseDTO {
	return responseDTO{
		Status:   rd.Status,
		Headers:  rd.Headers,
		Body:     string(rd.Body),
		BodyFile: rd.BodyFile,
	}
}

func stringMatcherToJSON(m StringMatcher) matcherJSON {
	switch v := m.(type) {
	case *equalToMatcher:
		return matcherJSON{EqualTo: &v.value}
	case *prefixMatcher:
		return matcherJSON{Prefix: &v.value}
	case *suffixMatcher:
		return matcherJSON{Suffix: &v.value}
	case *containsMatcher:
		return matcherJSON{Contains: &v.value}
	case *regexMatcher:
		s := v.re.String()
		return matcherJSON{Matches: &s}
	case *pathMatcher:
		return matcherJSON{EqualTo: &v.pattern}
	default:
		s := fmt.Sprintf("%v", m)
		return matcherJSON{EqualTo: &s}
	}
}

func bodyMatcherToJSON(m BodyMatcher) *bodyMatcherJSON {
	switch v := m.(type) {
	case *equalToBodyMatcher:
		s := string(v.expected)
		return &bodyMatcherJSON{EqualTo: &s}
	case *containsStringMatcher:
		return &bodyMatcherJSON{Contains: &v.substr}
	case *equalToJSONMatcher:
		return &bodyMatcherJSON{
			EqualToJSON: &v.expected,
			IgnoreExtra: v.ignoreExtra,
		}
	case *jsonPathMatcher:
		return &bodyMatcherJSON{
			MatchesJSONPath: &jsonPathJSON{
				Expression: v.path,
				Matcher:    stringMatcherToJSON(v.matcher),
			},
		}
	case *allOfMatcher:
		items := make([]bodyMatcherJSON, len(v.matchers))
		for i, child := range v.matchers {
			items[i] = *bodyMatcherToJSON(child)
		}
		return &bodyMatcherJSON{AllOf: items}
	default:
		return nil
	}
}

// dtoToStub converts a JSON stub DTO to an internal Stub.
func dtoToStub(dto stubDTO) (Stub, error) {
	if dto.Request.Method == "" {
		return Stub{}, errors.New("request.method is required")
	}
	if dto.Request.Path == "" {
		return Stub{}, errors.New("request.path is required")
	}
	if dto.Response.Status == 0 {
		return Stub{}, errors.New("response.status is required")
	}

	id := dto.ID
	if id == "" {
		id = uuid.New().String()
	}

	rm := RequestMatcher{
		Method: dto.Request.Method,
		Path:   compilePath(dto.Request.Path),
	}

	if len(dto.Request.Headers) > 0 {
		rm.Headers = make(map[string]StringMatcher, len(dto.Request.Headers))
		for k, v := range dto.Request.Headers {
			m, err := jsonToStringMatcher(v)
			if err != nil {
				return Stub{}, fmt.Errorf("headers[%s]: %w", k, err)
			}
			rm.Headers[k] = m
		}
	}

	if len(dto.Request.QueryParams) > 0 {
		rm.QueryParams = make(map[string]StringMatcher, len(dto.Request.QueryParams))
		for k, v := range dto.Request.QueryParams {
			m, err := jsonToStringMatcher(v)
			if err != nil {
				return Stub{}, fmt.Errorf("queryParams[%s]: %w", k, err)
			}
			rm.QueryParams[k] = m
		}
	}

	if dto.Request.Body != nil {
		bm, err := jsonToBodyMatcher(*dto.Request.Body)
		if err != nil {
			return Stub{}, fmt.Errorf("body: %w", err)
		}
		rm.Body = bm
	}

	resp := ResponseDef{
		Status:   dto.Response.Status,
		Headers:  dto.Response.Headers,
		Body:     []byte(dto.Response.Body),
		BodyFile: dto.Response.BodyFile,
	}

	return Stub{ID: id, Request: rm, Response: resp}, nil
}

func jsonToStringMatcher(j matcherJSON) (StringMatcher, error) {
	count := 0
	if j.EqualTo != nil {
		count++
	}
	if j.Prefix != nil {
		count++
	}
	if j.Suffix != nil {
		count++
	}
	if j.Contains != nil {
		count++
	}
	if j.Matches != nil {
		count++
	}
	if count == 0 {
		return nil, errors.New("matcher must have exactly one field set")
	}
	if count > 1 {
		return nil, errors.New("matcher must have exactly one field set")
	}

	switch {
	case j.EqualTo != nil:
		return EqualTo(*j.EqualTo), nil
	case j.Prefix != nil:
		return Prefix(*j.Prefix), nil
	case j.Suffix != nil:
		return Suffix(*j.Suffix), nil
	case j.Contains != nil:
		return Contains(*j.Contains), nil
	case j.Matches != nil:
		return MatchesRegex(*j.Matches), nil
	default:
		return nil, errors.New("matcher must have exactly one field set")
	}
}

func jsonToBodyMatcher(j bodyMatcherJSON) (BodyMatcher, error) {
	if len(j.AllOf) > 0 {
		matchers := make([]BodyMatcher, len(j.AllOf))
		for i, child := range j.AllOf {
			m, err := jsonToBodyMatcher(child)
			if err != nil {
				return nil, fmt.Errorf("allOf[%d]: %w", i, err)
			}
			matchers[i] = m
		}
		return AllOf(matchers...), nil
	}

	if j.MatchesJSONPath != nil {
		sm, err := jsonToStringMatcher(j.MatchesJSONPath.Matcher)
		if err != nil {
			return nil, fmt.Errorf("matchesJsonPath.matcher: %w", err)
		}
		return MatchesJSONPath(j.MatchesJSONPath.Expression, sm), nil
	}

	if j.EqualToJSON != nil {
		m := EqualToJSON(*j.EqualToJSON)
		if j.IgnoreExtra {
			m.IgnoreExtraFields()
		}
		return m, nil
	}

	if j.EqualTo != nil {
		return EqualToBody([]byte(*j.EqualTo)), nil
	}

	if j.Contains != nil {
		return ContainsString(*j.Contains), nil
	}

	return nil, errors.New("body matcher must have at least one field set")
}
