package wire

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

func MergeInto(dst map[string]any, patch map[string]any) {
	for key, patchVal := range patch {
		dstVal, exists := dst[key]
		if !exists {
			dst[key] = patchVal
			continue
		}

		pv, patchIsMap := patchVal.(map[string]any)
		if !patchIsMap {

			dst[key] = patchVal
			continue
		}

		switch dv := dstVal.(type) {
		case map[string]any:
			MergeInto(dv, pv)
		case []any:
			dst[key] = mergeArrayPatch(dv, pv)
		default:
			dst[key] = pv
		}
	}
}

func mergeArrayPatch(dst []any, patch map[string]any) []any {

	indices := make([]int, 0, len(patch))
	byIndex := make(map[int]any, len(patch))
	for k, v := range patch {
		idx, err := strconv.Atoi(k)
		if err != nil {

			continue
		}
		indices = append(indices, idx)
		byIndex[idx] = v
	}
	sort.Ints(indices)

	for _, idx := range indices {
		v := byIndex[idx]
		switch {
		case idx < len(dst):
			if dstElem, ok := dst[idx].(map[string]any); ok {
				if patchElem, ok := v.(map[string]any); ok {
					MergeInto(dstElem, patchElem)
					continue
				}
			}
			dst[idx] = v
		case idx == len(dst):
			dst = append(dst, v)
		default:

			for len(dst) < idx {
				dst = append(dst, nil)
			}
			dst = append(dst, v)
		}
	}

	return dst
}

func Reconstruct(records []Record, compressed bool) (map[string]any, error) {
	state := make(map[string]any)

	for i, rec := range records {
		raw, err := Decode(rec.Payload, compressed)
		if err != nil {
			return nil, fmt.Errorf("wire: record %d: %w", i, err)
		}

		var patch any
		if err := json.Unmarshal(raw, &patch); err != nil {
			return nil, fmt.Errorf("wire: record %d: unmarshal: %w", i, err)
		}

		patchMap, ok := patch.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("wire: record %d: top-level payload is not a JSON object (got %T)", i, patch)
		}

		MergeInto(state, patchMap)
	}

	return state, nil
}
