package common

// Tag represents a classification tag for foods and recipes
type Tag string

// Valid tag values
const (
	TagLockIn Tag = "lockin"
	TagOut    Tag = "out"
)

// AllTags contains all valid tag values
var AllTags = []Tag{
	TagLockIn,
	TagOut,
}

// IsValidTag checks if a tag string is valid
func IsValidTag(tag string) bool {
	for _, validTag := range AllTags {
		if string(validTag) == tag {
			return true
		}
	}
	return false
}

// ValidateTags checks if all tags in a slice are valid
func ValidateTags(tags []string) bool {
	for _, tag := range tags {
		if !IsValidTag(tag) {
			return false
		}
	}
	return true
}

// GetValidTagStrings returns all valid tags as strings
func GetValidTagStrings() []string {
	result := make([]string, len(AllTags))
	for i, tag := range AllTags {
		result[i] = string(tag)
	}
	return result
}
