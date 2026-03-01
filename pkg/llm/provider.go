package llm

// FieldType indicates what kind of UI element to render for a setup field.
type FieldType string

const (
	FieldInput  FieldType = "input"
	FieldSelect FieldType = "select"
)

// FieldOption is a label/value pair used by FieldSelect fields.
type FieldOption struct {
	Label string
	Value string
}

// SetupField describes a single piece of configuration a provider needs from the user.
// The setup wizard renders these generically using huh so providers never import UI code.
type SetupField struct {
	Key         string
	Type        FieldType     // defaults to FieldInput if zero
	Title       string
	Description string
	Placeholder string
	Secret      bool          // echo as password
	Default     string        // applied when the user leaves the field blank
	Options     []FieldOption // only used when Type == FieldSelect
	EnvFallback string        // environment variable to check when viper value is empty
}

// Provider is implemented by every LLM backend.
// Adding a new provider means implementing this interface and registering it
// in register_providers.go — no other files need to change.
type Provider interface {
	// ID returns the stable identifier stored in config.yaml (e.g. "openrouter").
	ID() string

	// DisplayName returns the human-readable name shown in the setup wizard.
	DisplayName() string

	// DefaultModel returns the model name used when the user leaves the field blank.
	DefaultModel() string

	// SetupFields describes the fields to collect during first-run setup.
	// The wizard renders them in order using huh.
	SetupFields() []SetupField

	// BuildClient constructs a ready-to-use LLMClient from the collected values
	// (either from the setup wizard or read from viper at runtime).
	// values keys match SetupField.Key; model may be empty (use DefaultModel).
	BuildClient(values map[string]string, model string) (LLMClient, error)
}
