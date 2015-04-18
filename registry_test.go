package justinstall

import (
	"fmt"
	"net/http"
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

	registry := SmartLoadRegistry(false)

	checkLink := func(rawurl string) bool {
		response, err := http.Get(rawurl)
		if err != nil {
			return false
		}
		defer response.Body.Close()

		return response.StatusCode == http.StatusOK
	}

	checkArch := func(name string, entry *registryEntry, architecture string, rawUrl string) {
		if rawUrl == "" {
			return
		}

		url := entry.expandString(rawUrl)

		if !checkLink(url) {
			errors = append(errors, fmt.Sprintf("%v (%v): %v", name, architecture, url))
		}
	}

	for _, name := range registry.SortedPackageNames() {
		entry := registry.Packages[name]

		checkArch(name, &entry, "x86", entry.Installer.X86)
		checkArch(name, &entry, "x86_64", entry.Installer.X86_64)
	}

	assert.Empty(t, errors)
}
