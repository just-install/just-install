package justinstall

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

func TestValidRegistry(t *testing.T) {
	schemaLoader := gojsonschema.NewReferenceLoader("file://./just-install-schema.json")
	documentLoader := gojsonschema.NewReferenceLoader("file://./just-install.json")

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)

	assert.Nil(t, err)
	assert.Empty(t, result.Errors())
	assert.True(t, result.Valid())
}

func TestRegistryReachableLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reachability test in short mode")
	}

	var errors []string

	registry := smartLoadRegistry(false)

	checkLink := func(rawurl string) bool {
		t.Logf("Checking %v", rawurl)

		response, err := http.Get(rawurl)
		if err != nil {
			return false
		}
		defer response.Body.Close()

		return response.StatusCode == http.StatusOK
	}

	checkArch := func(name string, version string, architecture string, rawUrl string) {
		if rawUrl == "" {
			return
		}

		url := strings.Replace(rawUrl, "${version}", version, -1)

		if !checkLink(url) {
			errors = append(errors, fmt.Sprintf("%v (%v): %v", name, architecture, url))
		}
	}

	for _, name := range registry.SortedPackageNames() {
		entry := registry.Packages[name]

		checkArch(name, entry.Version, "x86", entry.Installer.X86)
		checkArch(name, entry.Version, "x86_64", entry.Installer.X86_64)
	}

	assert.Empty(t, errors)
}
