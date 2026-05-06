package alto

import "strings"

// ApplyRenameCategoriesALTO renames categories according to the provided mapping.
// If the target label already exists, matching blocks are re-pointed to the target tag ID.
// Otherwise the original tag labels are renamed in place.
func ApplyRenameCategoriesALTO(doc *Alto, renames map[string]string) error {
	if doc == nil || len(renames) == 0 {
		return nil
	}

	replacements := make(map[string]string)
	relabels := make(map[string]string)

	for from, to := range renames {
		from = strings.TrimSpace(from)
		to = strings.TrimSpace(to)
		if from == "" || to == "" || from == to {
			continue
		}

		fromIDs := labelToIDSet(doc, from)
		if len(fromIDs) == 0 {
			continue
		}

		targetID, err := resolveTagID(doc, to)
		if err == nil {
			for fromID := range fromIDs {
				replacements[fromID] = targetID
			}
			continue
		}

		relabels[from] = to
	}

	if len(relabels) > 0 {
		for i := range doc.Tags.OtherTags {
			if next, ok := relabels[strings.TrimSpace(doc.Tags.OtherTags[i].Label)]; ok {
				doc.Tags.OtherTags[i].Label = next
			}
		}
	}

	if len(replacements) == 0 {
		return nil
	}

	for pi := range doc.Layout.Page {
		page := &doc.Layout.Page[pi]
		for bi := range page.PrintSpace.TextBlocks {
			block := &page.PrintSpace.TextBlocks[bi]
			tagIDs := strings.Fields(block.TagRefs)
			if len(tagIDs) == 0 {
				continue
			}

			seen := make(map[string]struct{}, len(tagIDs))
			rewritten := make([]string, 0, len(tagIDs))
			for _, tagID := range tagIDs {
				if replacement, ok := replacements[tagID]; ok {
					tagID = replacement
				}
				if _, exists := seen[tagID]; exists {
					continue
				}
				seen[tagID] = struct{}{}
				rewritten = append(rewritten, tagID)
			}

			block.TagRefs = strings.Join(rewritten, " ")
		}
	}

	return nil
}
