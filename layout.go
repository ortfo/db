package ortfodb

import (
	"fmt"
	ll "github.com/ewen-lbh/label-logger-go"
	"strconv"
)

// EmptyLayoutCell is a special value that represents an empty cell (used as a spacer, for example). Expressed in the user-provided layout as a nil value.
var EmptyLayoutCell = "empty"

// lcm returns the least common multiple of all the provided integers
func lcm(integers ...int) int {
	if len(integers) < 2 {
		return integers[0]
	}
	var greater int
	// choose the greater number
	if integers[0] > integers[1] {
		greater = integers[0]
	} else {
		greater = integers[1]
	}

	for {
		if (greater%integers[0] == 0) && (greater%integers[1] == 0) {
			break
		}
		greater += 1
	}
	if len(integers) == 2 {
		return greater
	}
	return lcm(append(integers[2:], greater)...)
}

// Normalize returns a normalized layout where every row has the same number of cells.
func (layout Layout) Normalize() (normalized Layout) {
	ll.Debug("normalizing layout %#v", layout)
	normalized = make(Layout, 0)

	// Determine the common width
	width := 1
	for _, row := range layout {
		width = lcm(width, len(row))
	}

	// Normalize every row
	for _, row := range layout {
		repeatFactor := width / len(row)
		normalizedRow := make([]LayoutCell, 0)
		for i := 0; i < width; i++ {
			// Spread rows evenly
			cell := row[i/repeatFactor]
			normalizedRow = append(normalizedRow, cell)
		}
		normalized = append(normalized, normalizedRow)
	}

	return
}

// Return a unique list of all the block IDs in the layout.
func (layout Layout) BlockIDs() (blockIDs []string) {
	blockIDs = make([]string, 0)
	for _, row := range layout {
		for _, cell := range row {
			// Check if the block ID is already in the list
			alreadyInList := false
			for _, blockID := range blockIDs {
				if blockID == string(cell) {
					alreadyInList = true
					break
				}
			}
			if !alreadyInList {
				blockIDs = append(blockIDs, string(cell))
			}
		}
	}
	return blockIDs
}

// ResolveLayout returns a layout, given the parsed description.
func ResolveLayout(metadata WorkMetadata, language string, blocks []ContentBlock) (Layout, error) {
	ll.Debug("Resolving layout from metadata %#v", metadata)
	layout := make(Layout, 0)
	userProvided := metadata.AdditionalMetadata["layout"]
	// Handle case where the layout is explicitly specified.
	if userProvided != "" && userProvided != nil {
		// User-provided layout uses block types and indices to refer to blocks. We need to convert those to block IDs.
		if _, ok := userProvided.([]interface{}); ok {
			for _, line := range userProvided.([]interface{}) {
				layoutLine := make([]LayoutCell, 0)
				ll.Debug("processing layout line %#v", line)
				switch line := line.(type) {
				case string:
					cell, err := ResolveBlockID(blocks, language, line)
					if err != nil {
						return layout, fmt.Errorf("while resolving block reference %q to ID: %w", line, err)
					}

					layoutLine = append(layoutLine, LayoutCell(cell))
				case nil:
					ll.Debug("encountered nil value in layout single line, treating as empty cell")
					layoutLine = append(layoutLine, LayoutCell(EmptyLayoutCell))
				case []interface{}:
					ll.Debug("processing layout line %#v", line)
					for _, cell := range line {
						if val, ok := cell.(string); ok {
							cell, err := ResolveBlockID(blocks, language, val)
							if err != nil {
								return layout, fmt.Errorf("while resolving block reference %q to ID: %w", val, err)
							}

							layoutLine = append(layoutLine, LayoutCell(cell))
						} else if cell == nil {
							ll.Debug("encountered nil value in layout line, treating as empty cell")
							layoutLine = append(layoutLine, LayoutCell(EmptyLayoutCell))
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
	ll.Debug("Layout resolved to %#v", layout)
	return layout.Normalize(), nil
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
