package alto

func ApplyFullCertainty(a *Alto) {
	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		for si := range page.PrintSpace.TextBlocks {
			block := &page.PrintSpace.TextBlocks[si]
			for li := range block.Lines {
				line := &block.Lines[li]
				for ci := range line.Strings {
					str := &line.Strings[ci]
					str.WC = 1.0
				}
			}
		}
	}
}
