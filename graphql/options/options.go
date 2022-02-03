package options

// ClientConfig is the config used for creating a new
// graphql client.
type ClientConfig struct {
	MutationURL        string
	QueryURL           string
	SubscriptionURL    string
	DefaultHTTPHeaders map[string]string
}
