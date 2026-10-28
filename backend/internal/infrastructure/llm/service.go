package llm

// Service is an alias to FederatedClient to preserve seamless backwards compatibility.
type Service = FederatedClient

// NewService wraps or returns the provided Client.
func NewService(client *Client) *Service {
	return client
}
