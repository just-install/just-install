package justinstall

import (
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
