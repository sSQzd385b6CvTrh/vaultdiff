package diff

import "sort"

// ChangeType represents the type of change between two secret versions.
type ChangeType string

const (
	Added    ChangeType = "added"
	Removed  ChangeType = "removed"
	Modified ChangeType = "modified"
	Unchanged ChangeType = "unchanged"
)

// Change represents a single key-level difference between two secret maps.
type Change struct {
	Key      string
	OldValue string
	NewValue string
	Type     ChangeType
}

// Result holds the full diff result between two secret versions.
type Result struct {
	Path    string
	Changes []Change
}

// Secrets compares two maps of secret key-value pairs and returns a Result.
func Secrets(path string, oldSecrets, newSecrets map[string]interface{}) Result {
	result := Result{Path: path}

	seen := make(map[string]bool)

	for key, newVal := range newSecrets {
		seen[key] = true
		newStr := toString(newVal)
		if oldVal, exists := oldSecrets[key]; exists {
			oldStr := toString(oldVal)
			if oldStr != newStr {
				result.Changes = append(result.Changes, Change{
					Key:      key,
					OldValue: oldStr,
					NewValue: newStr,
					Type:     Modified,
				})
			} else {
				result.Changes = append(result.Changes, Change{
					Key:  key,
					Type: Unchanged,
				})
			}
		} else {
			result.Changes = append(result.Changes, Change{
				Key:      key,
				NewValue: newStr,
				Type:     Added,
			})
		}
	}

	for key, oldVal := range oldSecrets {
		if !seen[key] {
			result.Changes = append(result.Changes, Change{
				Key:      key,
				OldValue: toString(oldVal),
				Type:     Removed,
			})
		}
	}

	sort.Slice(result.Changes, func(i, j int) bool {
		return result.Changes[i].Key < result.Changes[j].Key
	})

	return result
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
