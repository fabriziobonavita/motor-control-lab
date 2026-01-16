package artifacts

import (
	"encoding/json"
	"os"
)

// WriteJSON writes v as pretty-printed JSON.
func WriteJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
