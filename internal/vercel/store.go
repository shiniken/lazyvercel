package vercel

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Store struct {
	client              *Client
	projects            []Project
	initialProjectIndex int
	filters             DeploymentFilters
	refresh             time.Duration

	mu          sync.RWMutex
	deployments map[string][]Deployment
	summaries   map[string]Deployment
	details     map[string]DeploymentDetail
	lastRefresh time.Time
}

func NewStore(ctx context.Context, opts Options) (*Store, error) {
	linkedPins, err := DiscoverProjects(opts.Dirs)
	if err != nil {
		return nil, err
	}
	token, err := ResolveToken(opts.Token)
	if err != nil {
		return nil, err
	}

	client := NewClient(token)
	accounts, err := client.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}

	catalog, err := loadProjectCatalog(ctx, client, accounts)
	if err != nil {
		return nil, err
	}

	cwdLinked, hasCWD, err := DiscoverProjectIfLinked(".")
	if err != nil {
		return nil, err
	}
	projects, initialIndex := mergeProjects(catalog, linkedPins, cwdLinked, hasCWD)
	if len(projects) == 0 {
		return nil, fmt.Errorf("no Vercel projects found for this account")
	}

	store := &Store{
		client:              client,
		projects:            projects,
		initialProjectIndex: initialIndex,
		filters:             DeploymentFilters{Limit: opts.Limit, Target: opts.Target, Branch: opts.Branch},
		refresh:             opts.Refresh,
		deployments:         make(map[string][]Deployment),
		summaries:           make(map[string]Deployment),
		details:             make(map[string]DeploymentDetail),
	}

	if err := store.Refresh(ctx, projects[initialIndex]); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Projects() []Project {
	return append([]Project(nil), s.projects...)
}

func (s *Store) InitialProjectIndex() int {
	return s.initialProjectIndex
}

func (s *Store) Filters() DeploymentFilters {
	return s.filters
}

func (s *Store) RefreshInterval() time.Duration {
	return s.refresh
}

func (s *Store) LastRefresh() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRefresh
}

func (s *Store) Refresh(ctx context.Context, project Project) error {
	deployments, err := s.client.ListDeployments(ctx, project, s.filters)
	if err != nil {
		return fmt.Errorf("%s: %w", project.Name, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.deployments[project.ID] = deployments
	if len(deployments) > 0 {
		s.summaries[project.ID] = deployments[0]
	}
	s.details = make(map[string]DeploymentDetail)
	s.lastRefresh = time.Now()
	return nil
}

func (s *Store) Deployments(project Project) []Deployment {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Deployment(nil), s.deployments[project.ID]...)
}

func (s *Store) Summary(project Project) (Deployment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deployments := s.deployments[project.ID]
	if len(deployments) > 0 {
		return deployments[0], true
	}
	deployment, ok := s.summaries[project.ID]
	return deployment, ok
}

func (s *Store) RefreshSummaries(ctx context.Context, projects []Project) error {
	for _, project := range projects {
		if project.ID == "" {
			continue
		}
		filters := s.filters
		filters.Limit = 1
		deployments, err := s.client.ListDeployments(ctx, project, filters)
		if err != nil {
			return fmt.Errorf("%s: %w", project.Name, err)
		}
		s.mu.Lock()
		if len(deployments) > 0 {
			s.summaries[project.ID] = deployments[0]
		}
		s.mu.Unlock()
	}
	return nil
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

func (s *Store) BuildLogs(ctx context.Context, deployment Deployment) ([]BuildLogLine, error) {
	return s.client.GetBuildLogs(ctx, deployment.Project, deployment, 200)
}

func loadProjectCatalog(ctx context.Context, client *Client, accounts []Account) ([]Project, error) {
	var projects []Project
	for _, account := range accounts {
		accountProjects, err := client.ListProjects(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", account.Slug, err)
		}
		projects = append(projects, accountProjects...)
	}
	return projects, nil
}

func mergeProjects(catalog []Project, linkedPins []LinkedProject, cwdLinked LinkedProject, hasCWD bool) ([]Project, int) {
	projectsByID := make(map[string]*Project, len(catalog)+len(linkedPins)+1)
	for index := range catalog {
		project := catalog[index]
		projectsByID[project.ID] = &project
	}

	for _, linked := range linkedPins {
		project := mergeLinkedProject(projectsByID, linked)
		project.Pinned = true
		if project.Name == "" {
			project.Name = linked.ProjectName
		}
	}

	initialID := ""
	if hasCWD {
		project := mergeLinkedProject(projectsByID, cwdLinked)
		project.LinkedCWD = true
		if project.Name == "" {
			project.Name = cwdLinked.ProjectName
		}
		initialID = project.ID
	}

	projects := make([]Project, 0, len(projectsByID))
	for _, project := range projectsByID {
		projects = append(projects, *project)
	}

	sort.SliceStable(projects, func(i, j int) bool {
		left := projects[i]
		right := projects[j]
		if left.LinkedCWD != right.LinkedCWD {
			return left.LinkedCWD
		}
		if left.Pinned != right.Pinned {
			return left.Pinned
		}
		if left.UpdatedAt != right.UpdatedAt {
			return left.UpdatedAt > right.UpdatedAt
		}
		if left.AccountSlug != right.AccountSlug {
			return left.AccountSlug < right.AccountSlug
		}
		return left.Name < right.Name
	})

	initialIndex := 0
	for index, project := range projects {
		if project.ID == initialID {
			initialIndex = index
			break
		}
	}
	return projects, initialIndex
}

func mergeLinkedProject(projectsByID map[string]*Project, linked LinkedProject) *Project {
	project, ok := projectsByID[linked.ProjectID]
	if !ok {
		fallback := ProjectFromLinked(linked)
		project = &fallback
		projectsByID[linked.ProjectID] = project
	}
	project.LinkedDir = linked.Dir
	if project.AccountID == "" {
		project.AccountID = linked.OrgID
	}
	if project.Name == "" {
		project.Name = linked.ProjectName
	}
	return project
}
