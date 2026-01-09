package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Update all existing boards to include need_input column
		boards, err := app.FindRecordsByFilter("boards", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, board := range boards {
			columns := board.Get("columns")
			if columns == nil {
				continue
			}

			// Parse existing columns
			var colArray []string
			switch v := columns.(type) {
			case []string:
				colArray = v
			case []any:
				for _, c := range v {
					if s, ok := c.(string); ok {
						colArray = append(colArray, s)
					}
				}
			case string:
				if err := json.Unmarshal([]byte(v), &colArray); err != nil {
					continue
				}
			default:
				continue
			}

			// Check if need_input already exists
			hasNeedInput := false
			for _, col := range colArray {
				if col == "need_input" {
					hasNeedInput = true
					break
				}
			}

			if hasNeedInput {
				continue // Already has need_input
			}

			// Insert need_input after in_progress, before review
			var newColumns []string
			for _, col := range colArray {
				newColumns = append(newColumns, col)
				if col == "in_progress" {
					newColumns = append(newColumns, "need_input")
				}
			}

			// Update board
			board.Set("columns", newColumns)
			if err := app.Save(board); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		// Rollback: remove need_input from all boards' columns
		boards, err := app.FindRecordsByFilter("boards", "", "", 0, 0)
		if err != nil {
			return err
		}

		for _, board := range boards {
			columns := board.Get("columns")
			if columns == nil {
				continue
			}

			var colArray []string
			switch v := columns.(type) {
			case []string:
				colArray = v
			case []any:
				for _, c := range v {
					if s, ok := c.(string); ok {
						colArray = append(colArray, s)
					}
				}
			case string:
				if err := json.Unmarshal([]byte(v), &colArray); err != nil {
					continue
				}
			default:
				continue
			}

			// Remove need_input
			var newColumns []string
			for _, col := range colArray {
				if col != "need_input" {
					newColumns = append(newColumns, col)
				}
			}

			board.Set("columns", newColumns)
			if err := app.Save(board); err != nil {
				return err
			}
		}

		return nil
	})
}
