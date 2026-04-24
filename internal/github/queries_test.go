package github

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseViewerLogin(t *testing.T) {
	fixture := `{"viewer": {"login": "testuser"}}`

	var data struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if data.Viewer.Login != "testuser" {
		t.Errorf("expected testuser, got %s", data.Viewer.Login)
	}
}

func TestParseListsResponse(t *testing.T) {
	fixture := `{
        "viewer": {
            "lists": {
                "nodes": [
                    {
                        "id": "UL_1",
                        "name": "Go Projects",
                        "slug": "go-projects",
                        "description": "Great Go repos",
                        "isPrivate": false,
                        "updatedAt": "2024-01-15T10:00:00Z",
                        "lastAddedAt": "2024-01-15T10:00:00Z",
                        "items": {"totalCount": 42}
                    }
                ],
                "pageInfo": {
                    "endCursor": "abc123",
                    "hasNextPage": false
                }
            }
        }
    }`

	var data struct {
		Viewer struct {
			Lists struct {
				Nodes []struct {
					ID          string    `json:"id"`
					Name        string    `json:"name"`
					Slug        string    `json:"slug"`
					Description string    `json:"description"`
					IsPrivate   bool      `json:"isPrivate"`
					UpdatedAt   time.Time `json:"updatedAt"`
					Items       struct {
						TotalCount int `json:"totalCount"`
					} `json:"items"`
				} `json:"nodes"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"lists"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(data.Viewer.Lists.Nodes) != 1 {
		t.Fatalf("expected 1 list, got %d", len(data.Viewer.Lists.Nodes))
	}

	node := data.Viewer.Lists.Nodes[0]
	if node.ID != "UL_1" {
		t.Errorf("id: got %q want %q", node.ID, "UL_1")
	}
	if node.Name != "Go Projects" {
		t.Errorf("name: got %q want %q", node.Name, "Go Projects")
	}
	if node.Items.TotalCount != 42 {
		t.Errorf("totalCount: got %d want %d", node.Items.TotalCount, 42)
	}
	if data.Viewer.Lists.PageInfo.HasNextPage {
		t.Error("expected hasNextPage to be false")
	}
}

func TestParseStarredReposResponse(t *testing.T) {
	fixture := `{
        "viewer": {
            "starredRepositories": {
                "edges": [
                    {
                        "starredAt": "2024-03-15T12:00:00Z",
                        "node": {
                            "id": "R_123",
                            "nameWithOwner": "charmbracelet/bubbletea",
                            "description": "A TUI framework",
                            "url": "https://github.com/charmbracelet/bubbletea"
                        }
                    }
                ],
                "pageInfo": {
                    "endCursor": "cursor123",
                    "hasNextPage": true
                }
            }
        }
    }`

	var data struct {
		Viewer struct {
			StarredRepositories struct {
				Edges []struct {
					StarredAt time.Time `json:"starredAt"`
					Node      struct {
						ID            string `json:"id"`
						NameWithOwner string `json:"nameWithOwner"`
						Description   string `json:"description"`
						URL           string `json:"url"`
					} `json:"node"`
				} `json:"edges"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"starredRepositories"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	edges := data.Viewer.StarredRepositories.Edges
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}

	if edges[0].Node.NameWithOwner != "charmbracelet/bubbletea" {
		t.Errorf("nameWithOwner: got %q", edges[0].Node.NameWithOwner)
	}
	if !data.Viewer.StarredRepositories.PageInfo.HasNextPage {
		t.Error("expected hasNextPage to be true")
	}
}

func TestParseListItemsResponse(t *testing.T) {
	fixture := `{
        "node": {
            "items": {
                "nodes": [
                    {
                        "id": "R_456",
                        "nameWithOwner": "golang/go",
                        "description": "The Go programming language",
                        "url": "https://github.com/golang/go"
                    }
                ],
                "pageInfo": {
                    "endCursor": "itemcursor",
                    "hasNextPage": false
                }
            }
        }
    }`

	var data struct {
		Node struct {
			Items struct {
				Nodes []struct {
					ID            string `json:"id"`
					NameWithOwner string `json:"nameWithOwner"`
				} `json:"nodes"`
				PageInfo struct {
					HasNextPage bool `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"items"`
		} `json:"node"`
	}

	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(data.Node.Items.Nodes) != 1 {
		t.Fatalf("expected 1 item, got %d", len(data.Node.Items.Nodes))
	}
	if data.Node.Items.Nodes[0].ID != "R_456" {
		t.Errorf("id: got %q", data.Node.Items.Nodes[0].ID)
	}
}
