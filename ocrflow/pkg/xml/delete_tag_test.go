package xml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteTag(t *testing.T) {
	t.Run("basic deletion", func(t *testing.T) {
		origData := []byte(`<root>
	<Tag1>Value1</Tag1>
	<Tag2>Value2</Tag2>
	<Tag1>Value3</Tag1>
</root>`)
		expectedData := []byte(`<root>
	<Tag2>Value2</Tag2>
</root>`)
		modifiedData := DeleteTag(origData, "Tag1")
		if !assert.Equal(t, expectedData, modifiedData) {
			return
		}
	})
}
