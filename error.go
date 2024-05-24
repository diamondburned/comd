package comd

import (
	"encoding/json"
	"errors"
	"io"
)

// IsJSONError reports whether the error can be marshaled to JSON.
func IsJSONError(err error) bool {
	var je interface{ IsJSONError() bool }
	return errors.As(err, &je) && je.IsJSONError()
}

// WriteJSONError writes the error to the writer as JSON.
func WriteJSONError(w io.Writer, err error) error {
	if IsJSONError(err) {
		return json.NewEncoder(w).Encode(map[string]any{"error": err})
	} else {
		return json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
	}
}
