package justinstall

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

// expandString expands any environment variable in the given string, with additional variables
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

		if split[0] == "" && split[1] == "" {
			continue
		}

		split[0] = strings.ToUpper(split[0]) // Normalize variable names to upper case
		split[0] = strings.Replace(split[0], "(X86)", "_X86", -1)

		ret[split[0]] = split[1]
	}

	return ret
}
