package templates

import "strings"

var unicodeToASCII = map[rune]string{
	'\u2011': "-", // non-breaking hyphen
	'\u2012': "-", // figure dash
	'\u2013': "-", // en dash
	'\u2014': "-", // em dash
	'\u2212': "-", // minus sign
	'\u00A0': " ", // non-breaking space
	'\u200B': "",  // zero-width space
	'\u200C': "",  // zero-width non-joiner
	'\u200D': "",  // zero-width joiner
	'\uFEFF': "",  // BOM / zero-width no-break space
}

// NormalizeString replaces common Unicode lookalike characters with
// their ASCII equivalents. This prevents issues when values like restart
// policies are read from YAML and passed to Docker, which expects ASCII text.
func NormalizeString(s string) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		if replacement, ok := unicodeToASCII[r]; ok {
			// Return the first rune of the replacement string
			if len(replacement) > 0 {
				return rune(replacement[0])
			}
			return -1 // remove the character
		}
		return r
	}, s)
}

// Normalize applies NormalizeString to all template string fields that could
// contain Unicode lookalike characters.
func (t *Template) Normalize() {
	t.ID = NormalizeString(t.ID)
	t.Name = NormalizeString(t.Name)
	t.Description = NormalizeString(t.Description)
	t.Image = NormalizeString(t.Image)
	t.Restart = NormalizeString(t.Restart)
	for k, v := range t.Env {
		delete(t.Env, k)
		t.Env[NormalizeString(k)] = NormalizeString(v)
	}
}

// ValidRestartPolicies returns the set of restart policies that Docker
// accepts.
func ValidRestartPolicies() []string {
	return []string{"no", "always", "on-failure", "unless-stopped"}
}

// IsValidRestartPolicy checks whether the given restart policy is one
// that Docker accepts.
func IsValidRestartPolicy(policy string) bool {
	for _, p := range ValidRestartPolicies() {
		if policy == p {
			return true
		}
	}
	return false
}

type Template struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Image       string            `yaml:"image" json:"image"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Ports       []Port            `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes     []Volume          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Restart     string            `yaml:"restart,omitempty" json:"restart,omitempty"`
}

type Port struct {
	Host      int `yaml:"host" json:"host"`
	Container int `yaml:"container" json:"container"`
}

type Volume struct {
	Name      string `yaml:"name" json:"name"`
	Container string `yaml:"container" json:"container"`
}
