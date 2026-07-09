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

func (c *Client) ListDeployments(ctx context.Context, project LinkedProject, filters DeploymentFilters) ([]Deployment, error) {
	values := url.Values{}
	values.Set("projectId", project.ProjectID)
	values.Set("teamId", project.OrgID)
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

func (c *Client) GetDeployment(ctx context.Context, project LinkedProject, idOrURL string) (DeploymentDetail, error) {
	values := url.Values{}
	values.Set("teamId", project.OrgID)

	var result DeploymentDetail
	if err := c.getJSON(ctx, "/v13/deployments/"+url.PathEscape(idOrURL), values, &result); err != nil {
		return DeploymentDetail{}, err
	}
	result.Project = project
	return result, nil
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
