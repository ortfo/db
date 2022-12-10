package ortfodb

import (
	"fmt"
	"strconv"
)

// ResolveLayout returns a layout, given the parsed description.
func ResolveLayout(description ParsedWork, language string) (Layout, error) {
	layout := make([][]LayoutCell, 0)
	userProvided := description.Metadata.AdditionalMetadata["layout"]
	// Handle case where the layout is explicitly specified.
	if userProvided != "" && userProvided != nil {
		// User-provided layout uses block types and indices to refer to blocks. We need to convert those to block IDs.
		if _, ok := userProvided.([]interface{}); ok {
			for _, line := range userProvided.([]interface{}) {
				layoutLine := make([]LayoutCell, 0)
				switch line.(type) {
				case string:
					cell, err := ResolveBlockID(description, language, line.(string))
					if err != nil {
						return layout, fmt.Errorf("while resolving block reference %q to ID: %w", line.(string), err)
					}

					layoutLine = append(layoutLine, LayoutCell(cell))
				case []interface{}:
					for _, cell := range line.([]interface{}) {
						if val, ok := cell.(string); ok {
							cell, err := ResolveBlockID(description, language, val)
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
	}
	return layout, nil
}

// ResolveBlockID returns the ID of a block, given its ref (user-facing content block references comprising of a content block type shorthand and an index). This index is 1-based.
func ResolveBlockID(description ParsedWork, language string, blockRef string) (string, error) {
	typ, indexStr := blockRef[0:1], blockRef[1:]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return "", fmt.Errorf("invalid content block reference: %w", err)
	}

	switch typ {
	case "p":
		return description.Paragraphs[language][index-1].ID, nil
	case "m":
		return description.MediaEmbedDeclarations[language][index-1].ID, nil
	case "l":
		return description.Links[language][index-1].ID, nil
	}

	return "", fmt.Errorf("invalid content block reference: %s is not one of p, m, l", typ)
}
