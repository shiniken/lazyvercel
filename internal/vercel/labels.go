package vercel

func (d Deployment) StateLabel() string {
	if d.State != "" {
		return d.State
	}
	if d.ReadyState != "" {
		return d.ReadyState
	}
	return "UNKNOWN"
}
