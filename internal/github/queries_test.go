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
                            "url": "https://github.com/charmbracelet/bubbletea",
                            "primaryLanguage": {"name": "Go"},
                            "stargazerCount": 25000,
                            "forkCount": 1200
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
						ID             string `json:"id"`
						NameWithOwner  string `json:"nameWithOwner"`
						Description    string `json:"description"`
						URL            string `json:"url"`
						PrimaryLanguage struct {
							Name string `json:"name"`
						} `json:"primaryLanguage"`
						StargazerCount int `json:"stargazerCount"`
						ForkCount      int `json:"forkCount"`
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
	if edges[0].Node.PrimaryLanguage.Name != "Go" {
		t.Errorf("primaryLanguage.name: got %q want %q", edges[0].Node.PrimaryLanguage.Name, "Go")
	}
	if edges[0].Node.StargazerCount != 25000 {
		t.Errorf("stargazerCount: got %d want %d", edges[0].Node.StargazerCount, 25000)
	}
	if edges[0].Node.ForkCount != 1200 {
		t.Errorf("forkCount: got %d want %d", edges[0].Node.ForkCount, 1200)
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
                        "url": "https://github.com/golang/go",
                        "primaryLanguage": {"name": "Go"},
                        "stargazerCount": 120000,
                        "forkCount": 17000
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
					ID             string `json:"id"`
					NameWithOwner  string `json:"nameWithOwner"`
					Description    string `json:"description"`
					URL            string `json:"url"`
					PrimaryLanguage struct {
						Name string `json:"name"`
					} `json:"primaryLanguage"`
					StargazerCount int `json:"stargazerCount"`
					ForkCount      int `json:"forkCount"`
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
	node := data.Node.Items.Nodes[0]
	if node.ID != "R_456" {
		t.Errorf("id: got %q", node.ID)
	}
	if node.PrimaryLanguage.Name != "Go" {
		t.Errorf("primaryLanguage.name: got %q want %q", node.PrimaryLanguage.Name, "Go")
	}
	if node.StargazerCount != 120000 {
		t.Errorf("stargazerCount: got %d want %d", node.StargazerCount, 120000)
	}
	if node.ForkCount != 17000 {
		t.Errorf("forkCount: got %d want %d", node.ForkCount, 17000)
	}
}

func TestParseCreateListResponse(t *testing.T) {
	fixture := `{
		"createUserList": {
			"list": {
				"id": "UL_new",
				"name": "My New List",
				"slug": "my-new-list",
				"description": "A test list",
				"isPrivate": true,
				"updatedAt": "2024-06-01T12:00:00Z"
			}
		}
	}`

	var data struct {
		CreateUserList struct {
			List struct {
				ID          string    `json:"id"`
				Name        string    `json:"name"`
				Slug        string    `json:"slug"`
				Description string    `json:"description"`
				IsPrivate   bool      `json:"isPrivate"`
				UpdatedAt   time.Time `json:"updatedAt"`
			} `json:"list"`
		} `json:"createUserList"`
	}
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	l := data.CreateUserList.List
	if l.ID != "UL_new" {
		t.Errorf("id: got %q want %q", l.ID, "UL_new")
	}
	if l.Name != "My New List" {
		t.Errorf("name: got %q want %q", l.Name, "My New List")
	}
	if l.Slug != "my-new-list" {
		t.Errorf("slug: got %q want %q", l.Slug, "my-new-list")
	}
	if l.Description != "A test list" {
		t.Errorf("description: got %q want %q", l.Description, "A test list")
	}
	if !l.IsPrivate {
		t.Error("expected isPrivate to be true")
	}
	expected := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if !l.UpdatedAt.Equal(expected) {
		t.Errorf("updatedAt: got %v want %v", l.UpdatedAt, expected)
	}
}

func TestParseUpdateListResponse(t *testing.T) {
	fixture := `{
		"updateUserList": {
			"list": {
				"id": "UL_1",
				"name": "Renamed List"
			}
		}
	}`

	var data struct {
		UpdateUserList struct {
			List struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"list"`
		} `json:"updateUserList"`
	}
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if data.UpdateUserList.List.ID != "UL_1" {
		t.Errorf("id: got %q want %q", data.UpdateUserList.List.ID, "UL_1")
	}
	if data.UpdateUserList.List.Name != "Renamed List" {
		t.Errorf("name: got %q want %q", data.UpdateUserList.List.Name, "Renamed List")
	}
}

func TestParseDeleteListResponse(t *testing.T) {
	fixture := `{
		"deleteUserList": {
			"user": {
				"login": "testuser"
			}
		}
	}`

	var data struct {
		DeleteUserList struct {
			User struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"deleteUserList"`
	}
	if err := json.Unmarshal([]byte(fixture), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if data.DeleteUserList.User.Login != "testuser" {
		t.Errorf("login: got %q want %q", data.DeleteUserList.User.Login, "testuser")
	}
}
