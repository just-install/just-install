package justinstall

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

// expandString expands any environment variable in the given string, with additonal variables
// coming from the given context.
func expandString(s string, context map[string]string) string {
	data := environMap()

	// Merge the given context
	for k, v := range context {
		data[k] = v
	}

	var buf bytes.Buffer

	template.Must(template.New("expand").Parse(s)).Execute(&buf, data)

	return buf.String()
}

// environMap returns the current environment variables as a map.
func environMap() map[string]string {
	ret := make(map[string]string)
	env := os.Environ()

	for _, v := range env {
		split := strings.SplitN(v, "=", 2)
		ret[split[0]] = split[1]
	}

	return ret
}
