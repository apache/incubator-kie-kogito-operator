package api

// WebHookType literal type to distinguish between different types of webHooks.
type WebHookType string

const (
	// GitHubWebHook GitHub webHook.
	GitHubWebHook WebHookType = "GitHub"
	// GenericWebHook Generic webHook.
	GenericWebHook WebHookType = "Generic"
)

// WebHookSecret Secret to use for a given webHook.
// +k8s:openapi-gen=true
type WebHookSecret struct {
	// WebHook type, either GitHub or Generic.
	// +kubebuilder:validation:Enum=GitHub;Generic
	Type WebHookType `json:"type,omitempty"`
	// Secret value for webHook
	Secret string `json:"secret,omitempty"`
}
