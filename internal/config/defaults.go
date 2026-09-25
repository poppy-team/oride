package config

// DefaultKeyBindings returns the shipped chord → action id map.
//
// A copy, so a caller that mutates what it was handed cannot change the
// defaults for the rest of the process.
func DefaultKeyBindings() map[string]string {
	out := make(map[string]string, len(defaultKeyBindings))
	for chord, actionID := range defaultKeyBindings {
		out[chord] = actionID
	}
	return out
}
