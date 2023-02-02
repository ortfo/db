package ortfodb

import (
	"fmt"
	"strconv"
)

// ResolveLayout returns a layout, given the parsed description.
func ResolveLayout(metadata WorkMetadata, language string, blocks []ContentBlock) (Layout, error) {
	layout := make([][]LayoutCell, 0)
	userProvided := metadata.AdditionalMetadata["layout"]
	// Handle case where the layout is explicitly specified.
	if userProvided != "" && userProvided != nil {
		// User-provided layout uses block types and indices to refer to blocks. We need to convert those to block IDs.
		if _, ok := userProvided.([]interface{}); ok {
			for _, line := range userProvided.([]interface{}) {
				layoutLine := make([]LayoutCell, 0)
				switch line.(type) {
				case string:
					cell, err := ResolveBlockID(blocks, language, line.(string))
					if err != nil {
						return layout, fmt.Errorf("while resolving block reference %q to ID: %w", line.(string), err)
					}

					layoutLine = append(layoutLine, LayoutCell(cell))
				case []interface{}:
					for _, cell := range line.([]interface{}) {
						if val, ok := cell.(string); ok {
							cell, err := ResolveBlockID(blocks, language, val)
							if err != nil {
								return layout, fmt.Errorf("while resolving block reference %q to ID: %w", val, err)
							}

							layoutLine = append(layoutLine, LayoutCell(cell))
						}
					}
				}
				layout = append(layout, layoutLine)
			}
		}
	} else {
		// If no layout is specified, we use the default layout.
		for _, block := range blocks {
			layout = append(layout, []LayoutCell{LayoutCell(block.ID)})
		}
	}
	return layout, nil
}

// ResolveBlockID returns the ID of a block, given its ref (user-facing content block references comprising of a content block type shorthand and an index). This index is 1-based.
func ResolveBlockID(blocks []ContentBlock, language string, blockRef string) (string, error) {
	typ, indexStr := blockRef[0:1], blockRef[1:]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", fmt.Errorf("invalid content block reference: %w", err)
	}

	currentIndexByType := map[ContentBlockType]int{"paragraph": 0, "media": 0, "link": 0}

	if typ != "p" && typ != "m" && typ != "l" {
		return "", fmt.Errorf("invalid content block reference: %s is not one of p, m, l", typ)
	}

	for _, block := range blocks {
		currentIndexByType[block.Type]++
		if currentIndexByType[block.Type] == index && string(block.Type)[0:1] == typ {
			return block.ID, nil
		}
	}

	return "", fmt.Errorf("invalid content block reference: %s%d does not exist", typ, index)

}
