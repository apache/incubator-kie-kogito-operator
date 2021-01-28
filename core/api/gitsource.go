package api

// GitSource Git coordinates to locate the source code to build.
// +k8s:openapi-gen=true
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="Kogito Git Source"
type GitSource struct {
	// Git URI for the s2i source.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git URI"
	URI string `json:"uri"`
	// Branch to use in the Git repository.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Reference"
	Reference string `json:"reference,omitempty"`
	// Context/subdirectory where the code is located, relative to the repo root.
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Git Context"
	ContextDir string `json:"contextDir,omitempty"`
}
