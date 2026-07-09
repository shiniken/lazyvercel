package vercel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBaseURL = "https://api.vercel.com"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) ListAccounts(ctx context.Context) ([]Account, error) {
	values := url.Values{}
	values.Set("limit", "100")

	var result listTeamsResponse
	if err := c.getJSON(ctx, "/v2/teams", values, &result); err != nil {
		return nil, err
	}

	accounts := make([]Account, 0, len(result.Teams))
	for _, team := range result.Teams {
		name := team.Name
		if name == "" {
			name = team.Slug
		}
		accounts = append(accounts, Account{
			ID:   team.ID,
			Slug: team.Slug,
			Name: name,
		})
	}
	return accounts, nil
}

func (c *Client) ListProjects(ctx context.Context, account Account) ([]Project, error) {
	var projects []Project
	seen := map[string]bool{}
	var until int64

	for page := 0; page < 20; page++ {
		values := url.Values{}
		values.Set("teamId", account.ID)
		values.Set("limit", "100")
		if until > 0 {
			values.Set("until", fmt.Sprintf("%d", until-1))
		}

		var result listProjectsResponse
		if err := c.getJSON(ctx, "/v10/projects", values, &result); err != nil {
			return nil, err
		}

		added := 0
		for _, item := range result.Projects {
			if item.ID == "" || seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			added++
			projects = append(projects, Project{
				ID:          item.ID,
				Name:        item.Name,
				AccountID:   account.ID,
				AccountSlug: account.Slug,
				AccountName: account.Name,
				Framework:   item.Framework,
				UpdatedAt:   item.UpdatedAt,
				Link:        item.Link,
			})
		}
		if result.Pagination.Next == 0 || added == 0 || len(result.Projects) == 0 {
			break
		}
		until = result.Pagination.Next
	}

	return projects, nil
}

func (c *Client) ListDeployments(ctx context.Context, project Project, filters DeploymentFilters) ([]Deployment, error) {
	values := url.Values{}
	values.Set("projectId", project.ID)
	values.Set("teamId", project.AccountID)
	values.Set("limit", fmt.Sprintf("%d", filters.Limit))
	if filters.Target != "" {
		values.Set("target", filters.Target)
	}
	if filters.Branch != "" {
		values.Set("branch", filters.Branch)
	}

	var result listDeploymentsResponse
	if err := c.getJSON(ctx, "/v7/deployments", values, &result); err != nil {
		return nil, err
	}

	for index := range result.Deployments {
		result.Deployments[index].Project = project
	}
	return result.Deployments, nil
}

func (c *Client) GetDeployment(ctx context.Context, project Project, idOrURL string) (DeploymentDetail, error) {
	values := url.Values{}
	values.Set("teamId", project.AccountID)

	var result DeploymentDetail
	if err := c.getJSON(ctx, "/v13/deployments/"+url.PathEscape(idOrURL), values, &result); err != nil {
		return DeploymentDetail{}, err
	}
	result.Project = project
	return result, nil
}

func (c *Client) GetBuildLogs(ctx context.Context, project Project, deployment Deployment, limit int) ([]BuildLogLine, error) {
	key := deployment.UID
	if key == "" {
		key = deployment.URL
	}
	if key == "" {
		return nil, fmt.Errorf("deployment has neither uid nor url")
	}
	if limit <= 0 {
		limit = 200
	}

	values := url.Values{}
	values.Set("teamId", project.AccountID)
	values.Set("builds", "1")
	values.Set("limit", fmt.Sprintf("%d", limit))
	values.Set("direction", "backward")

	var events []deploymentEvent
	if err := c.getJSON(ctx, "/v3/deployments/"+url.PathEscape(key)+"/events", values, &events); err != nil {
		return nil, err
	}

	lines := make([]BuildLogLine, 0, len(events))
	for _, event := range events {
		line := event.BuildLogLine()
		if line.Text == "" && line.Step == "" && line.Entrypoint == "" && line.StatusCode == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, output any) error {
	endpoint, err := url.Parse(apiBaseURL + path)
	if err != nil {
		return err
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "lazyvercel")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 4*1024*1024))
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Vercel API %s failed with %s: %s", path, res.Status, summarizeBody(body))
	}

	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("decode Vercel API %s: %w", path, err)
	}
	return nil
}

func summarizeBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return "empty response"
	}
	if len(text) > 500 {
		return text[:500] + "..."
	}
	return text
}
