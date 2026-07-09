package vercel

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Store struct {
	client   *Client
	projects []LinkedProject
	filters  DeploymentFilters

	mu          sync.RWMutex
	deployments map[string][]Deployment
	details     map[string]DeploymentDetail
	lastRefresh time.Time
}

func NewStore(ctx context.Context, opts Options) (*Store, error) {
	projects, err := DiscoverProjects(opts.Dirs)
	if err != nil {
		return nil, err
	}
	token, err := ResolveToken(opts.Token)
	if err != nil {
		return nil, err
	}

	store := &Store{
		client:      NewClient(token),
		projects:    projects,
		filters:     DeploymentFilters{Limit: opts.Limit, Target: opts.Target, Branch: opts.Branch},
		deployments: make(map[string][]Deployment),
		details:     make(map[string]DeploymentDetail),
	}

	if err := store.Refresh(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Projects() []LinkedProject {
	return append([]LinkedProject(nil), s.projects...)
}

func (s *Store) Filters() DeploymentFilters {
	return s.filters
}

func (s *Store) LastRefresh() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh
}

func (s *Store) Refresh(ctx context.Context) error {
	next := make(map[string][]Deployment)
	for _, project := range s.projects {
		deployments, err := s.client.ListDeployments(ctx, project, s.filters)
		if err != nil {
			return fmt.Errorf("%s: %w", project.ProjectName, err)
		}
		next[project.ProjectID] = deployments
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.deployments = next
	s.lastRefresh = time.Now()
	return nil
}

func (s *Store) Deployments(project LinkedProject) []Deployment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Deployment(nil), s.deployments[project.ProjectID]...)
}

func (s *Store) Detail(ctx context.Context, deployment Deployment) (DeploymentDetail, error) {
	key := deployment.UID
	if key == "" {
		key = deployment.URL
	}
	if key == "" {
		return DeploymentDetail{}, fmt.Errorf("deployment has neither uid nor url")
	}

	s.mu.RLock()
	cached, ok := s.details[key]
	s.mu.RUnlock()
	if ok {
		return cached, nil
	}

	detail, err := s.client.GetDeployment(ctx, deployment.Project, key)
	if err != nil {
		return DeploymentDetail{}, err
	}

	s.mu.Lock()
	s.details[key] = detail
	s.mu.Unlock()
	return detail, nil
}
