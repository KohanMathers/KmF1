package feeds

import (
	"encoding/json"
	"fmt"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

func ParseArchiveStatus(data []byte) (ArchiveStatus, error) {
	var status ArchiveStatus
	if err := json.Unmarshal(wire.TrimBOM(data), &status); err != nil {
		return ArchiveStatus{}, fmt.Errorf("feeds: ArchiveStatus: unmarshal: %w", err)
	}
	return status, nil
}
