package vercel

type DeploymentFilters struct {
	Limit  int
	Target string
	Branch string
}

type listDeploymentsResponse struct {
	Deployments []Deployment `json:"deployments"`
}

type Deployment struct {
	UID              string         `json:"uid"`
	Name             string         `json:"name"`
	URL              string         `json:"url"`
	State            string         `json:"state"`
	ReadyState       string         `json:"readyState"`
	ReadySubstate    string         `json:"readySubstate"`
	Target           string         `json:"target"`
	CreatedAt        int64          `json:"createdAt"`
	BuildingAt       int64          `json:"buildingAt"`
	Ready            int64          `json:"ready"`
	InspectorURL     string         `json:"inspectorUrl"`
	ErrorCode        string         `json:"errorCode"`
	ErrorMessage     string         `json:"errorMessage"`
	ChecksState      string         `json:"checksState"`
	ChecksConclusion string         `json:"checksConclusion"`
	Creator          Creator        `json:"creator"`
	Meta             map[string]any `json:"meta"`
	Project          LinkedProject  `json:"-"`
}

type Creator struct {
	UID         string `json:"uid"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	GitHubLogin string `json:"githubLogin"`
	GitLabLogin string `json:"gitlabLogin"`
}

type DeploymentDetail struct {
	Deployment
	Alias            []string        `json:"alias"`
	Source           string          `json:"source"`
	Type             string          `json:"type"`
	GitSource        map[string]any  `json:"gitSource"`
	Builds           []Build         `json:"builds"`
	ProjectSettings  ProjectSettings `json:"projectSettings"`
	BuildErrorAt     int64           `json:"buildErrorAt"`
	CanceledAt       int64           `json:"canceledAt"`
	AliasAssignedAt  int64           `json:"aliasAssignedAt"`
	AliasAssigned    any             `json:"aliasAssigned"`
	AliasError       *AliasError     `json:"aliasError"`
	ReadySubstate    string          `json:"readySubstate"`
	ChecksState      string          `json:"checksState"`
	ChecksConclusion string          `json:"checksConclusion"`
}

type Build struct {
	Src             string `json:"src"`
	Use             string `json:"use"`
	CreatedAt       int64  `json:"createdAt"`
	ReadyState      string `json:"readyState"`
	ReadyStateAt    int64  `json:"readyStateAt"`
	ErrorCode       string `json:"errorCode"`
	ErrorMessage    string `json:"errorMessage"`
	LambdaRuntime   string `json:"lambdaRuntime"`
	Runtime         string `json:"runtime"`
	MaxDuration     int    `json:"maxDuration"`
	Regions         []any  `json:"regions"`
	Entrypoint      string `json:"entrypoint"`
	OutputDirectory string `json:"outputDirectory"`
}

type ProjectSettings struct {
	Framework       string `json:"framework"`
	BuildCommand    string `json:"buildCommand"`
	DevCommand      string `json:"devCommand"`
	InstallCommand  string `json:"installCommand"`
	OutputDirectory string `json:"outputDirectory"`
	RootDirectory   string `json:"rootDirectory"`
	NodeVersion     string `json:"nodeVersion"`
}

type AliasError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
