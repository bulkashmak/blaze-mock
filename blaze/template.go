package blaze

import "strings"

// renderTemplate replaces {{.key}} placeholders in tmpl with values from the map.
func renderTemplate(tmpl string, values map[string]string) string {
	result := tmpl
	for key, val := range values {
		result = strings.ReplaceAll(result, "{{."+key+"}}", val)
	}
	return result
}
