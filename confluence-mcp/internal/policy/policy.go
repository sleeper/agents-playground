package policy

import "slices"

// Store holds authorization policies keyed by subject identifiers.
type Store struct {
	defaultSpaces []string
	subjects      map[string][]string
}

// NewStore builds a policy store from default and subject-specific rules.
func NewStore(defaultSpaces []string, policies map[string][]string) *Store {
	return &Store{
		defaultSpaces: dedupe(defaultSpaces),
		subjects:      policies,
	}
}

// AllowedSpaces returns the set of spaces a subject may access, merging defaults.
func (s *Store) AllowedSpaces(subject string) []string {
	merged := append([]string{}, s.defaultSpaces...)
	if spaces, ok := s.subjects[subject]; ok {
		merged = append(merged, spaces...)
	}
	return dedupe(merged)
}

func dedupe(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var result []string
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	slices.Sort(result)
	return result
}
