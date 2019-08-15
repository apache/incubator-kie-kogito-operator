package inventory

// Namespace will fetch the inner API with a default client
func Namespace() NamespaceInterface {
	return newNamespace(&Client{})
}

// NamespaceC will use a defined client to fetch the inner API
func NamespaceC(c *Client) NamespaceInterface {
	return newNamespace(c)
}
