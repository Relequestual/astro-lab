package github

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Relequestual/astro-lab/internal/models"
)

// ViewerLogin queries the authenticated user's login
func (c *Client) ViewerLogin(ctx context.Context) (string, error) {
	resp, err := c.DoWithRetry(ctx, GraphQLRequest{
		Query: `query { viewer { login } }`,
	})
	if err != nil {
		return "", err
	}

	var data struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("parsing viewer login: %w", err)
	}
	return data.Viewer.Login, nil
}

// FetchLists fetches all user lists with pagination
func (c *Client) FetchLists(ctx context.Context) ([]models.StarList, error) {
	var allLists []models.StarList
	var cursor *string

	for {
		vars := map[string]interface{}{
			"first": 100,
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		resp, err := c.DoWithRetry(ctx, GraphQLRequest{
			Query: `query Lists($first: Int!, $after: String) {
                viewer {
                    lists(first: $first, after: $after) {
                        nodes {
                            id
                            name
                            slug
                            description
                            isPrivate
                            updatedAt
                            lastAddedAt
                            items(first: 0) { totalCount }
                        }
                        pageInfo {
                            endCursor
                            hasNextPage
                        }
                    }
                }
            }`,
			Variables: vars,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching lists: %w", err)
		}

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
						LastAddedAt time.Time `json:"lastAddedAt"`
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
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return nil, fmt.Errorf("parsing lists: %w", err)
		}

		for _, n := range data.Viewer.Lists.Nodes {
			allLists = append(allLists, models.StarList{
				ID:          n.ID,
				Name:        n.Name,
				Slug:        n.Slug,
				Description: n.Description,
				IsPrivate:   n.IsPrivate,
				ItemCount:   n.Items.TotalCount,
				UpdatedAt:   n.UpdatedAt,
				LastAddedAt: n.LastAddedAt,
			})
		}

		if !data.Viewer.Lists.PageInfo.HasNextPage {
			break
		}
		cursor = &data.Viewer.Lists.PageInfo.EndCursor
	}

	return allLists, nil
}

// FetchStarredRepos fetches starred repositories with pagination
// If since is non-zero, stops when reaching repos starred before that time
func (c *Client) FetchStarredRepos(ctx context.Context, since time.Time) ([]models.Repository, error) {
	var allRepos []models.Repository
	var cursor *string

	for {
		vars := map[string]interface{}{
			"first": 100,
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		resp, err := c.DoWithRetry(ctx, GraphQLRequest{
			Query: `query Stars($first: Int!, $after: String) {
                viewer {
                    starredRepositories(first: $first, after: $after, orderBy: { field: STARRED_AT, direction: DESC }) {
                        edges {
                            starredAt
                            node {
                                id
                                nameWithOwner
                                description
                                url
                            }
                        }
                        pageInfo {
                            endCursor
                            hasNextPage
                        }
                    }
                }
            }`,
			Variables: vars,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching stars: %w", err)
		}

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
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return nil, fmt.Errorf("parsing stars: %w", err)
		}

		hitCutoff := false
		for _, e := range data.Viewer.StarredRepositories.Edges {
			if !since.IsZero() && e.StarredAt.Before(since) {
				hitCutoff = true
				break
			}
			allRepos = append(allRepos, models.Repository{
				ID:            e.Node.ID,
				NameWithOwner: e.Node.NameWithOwner,
				StarredAt:     e.StarredAt,
				Description:   e.Node.Description,
				URL:           e.Node.URL,
			})
		}

		if hitCutoff || !data.Viewer.StarredRepositories.PageInfo.HasNextPage {
			break
		}
		cursor = &data.Viewer.StarredRepositories.PageInfo.EndCursor
	}

	return allRepos, nil
}

// FetchListItems fetches all items in a specific list with pagination
func (c *Client) FetchListItems(ctx context.Context, listID string) ([]models.Repository, error) {
	var allItems []models.Repository
	var cursor *string

	for {
		vars := map[string]interface{}{
			"listId": listID,
			"first":  100,
		}
		if cursor != nil {
			vars["after"] = *cursor
		}

		resp, err := c.DoWithRetry(ctx, GraphQLRequest{
			Query: `query ListItems($listId: ID!, $first: Int!, $after: String) {
                node(id: $listId) {
                    ... on UserList {
                        items(first: $first, after: $after) {
                            nodes {
                                ... on Repository {
                                    id
                                    nameWithOwner
                                    description
                                    url
                                }
                            }
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                        }
                    }
                }
            }`,
			Variables: vars,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching list items: %w", err)
		}

		var data struct {
			Node struct {
				Items struct {
					Nodes []struct {
						ID            string `json:"id"`
						NameWithOwner string `json:"nameWithOwner"`
						Description   string `json:"description"`
						URL           string `json:"url"`
					} `json:"nodes"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
					} `json:"pageInfo"`
				} `json:"items"`
			} `json:"node"`
		}
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			return nil, fmt.Errorf("parsing list items: %w", err)
		}

		for _, n := range data.Node.Items.Nodes {
			if n.ID == "" {
				continue // skip non-repository items
			}
			allItems = append(allItems, models.Repository{
				ID:            n.ID,
				NameWithOwner: n.NameWithOwner,
				Description:   n.Description,
				URL:           n.URL,
			})
		}

		if !data.Node.Items.PageInfo.HasNextPage {
			break
		}
		cursor = &data.Node.Items.PageInfo.EndCursor
	}

	return allItems, nil
}

// UpdateUserListsForItem sets the list memberships for a repository
func (c *Client) UpdateUserListsForItem(ctx context.Context, itemID string, listIDs []string) error {
	ids := make([]interface{}, len(listIDs))
	for i, id := range listIDs {
		ids[i] = id
	}

	_, err := c.DoWithRetry(ctx, GraphQLRequest{
		Query: `mutation UpdateListsForItem($input: UpdateUserListsForItemInput!) {
            updateUserListsForItem(input: $input) {
                item {
                    __typename
                    ... on Repository {
                        id
                        nameWithOwner
                    }
                }
            }
        }`,
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"itemId":  itemID,
				"listIds": ids,
			},
		},
	})
	return err
}
