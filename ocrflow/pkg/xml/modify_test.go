package xml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModifyTag(t *testing.T) {
	t.Run("basic modification", func(t *testing.T) {
		origData := []byte(`<root>
	<Tag1>Value1</Tag1>
	<Tag2>Value2</Tag2>
	<Tag1>Value3</Tag1>
</root>`)
		expectedData := []byte(`<root>
	<Tag1>ModifiedValue1</Tag1>
	<Tag2>Value2</Tag2>
	<Tag1>ModifiedValue3</Tag1>
</root>`)
		modifiedData := ModifyTag(origData, "Tag1", func(v string) string {
			return "Modified" + string(v)
		})

		if !assert.Equal(t, string(expectedData), string(modifiedData)) {
			return
		}
	})

	t.Run("tag not present", func(t *testing.T) {
		origData := []byte(`<root>
	<Tag2>Value2</Tag2>
</root>`)
		expectedData := origData
		modifiedData := ModifyTag(origData, "Tag1", func(v string) string {
			return "Modified" + string(v)
		})
		if !assert.Equal(t, string(expectedData), string(modifiedData)) {
			return
		}
	})
}
