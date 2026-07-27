package wire

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func Decode(payload string, compressed bool) ([]byte, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return nil, fmt.Errorf("wire: empty payload")
	}

	if trimmed[0] == '{' || trimmed[0] == '[' {
		return []byte(trimmed), nil
	}

	var inner string
	if err := json.Unmarshal([]byte(trimmed), &inner); err != nil {
		return nil, fmt.Errorf("wire: payload is neither a JSON object/array nor a JSON string: %w", err)
	}

	if !compressed {
		return []byte(inner), nil
	}

	return inflate(inner)
}

func inflate(s string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("wire: base64 decode: %w", err)
	}

	r := flate.NewReader(bytes.NewReader(compressed))
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("wire: inflate: %w", err)
	}

	return out, nil
}
