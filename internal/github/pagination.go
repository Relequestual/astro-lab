package github

// PageInfo holds pagination state from GraphQL responses
type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}
