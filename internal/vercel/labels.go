package vercel

import "strings"

func (d Deployment) StateLabel() string {
	if d.State != "" {
		return d.State
	}
	if d.ReadyState != "" {
		return d.ReadyState
	}
	return "UNKNOWN"
}

func (e deploymentEvent) BuildLogLine() BuildLogLine {
	created := e.Created.Int64()
	if created == 0 {
		created = e.Date.Int64()
	}
	if created == 0 {
		created = e.Payload.Created.Int64()
	}
	if created == 0 {
		created = e.Payload.Date.Int64()
	}

	info := e.Info
	if (info == eventInfo{}) {
		info = e.Payload.Info
	}

	step := strings.TrimSpace(info.Step)
	if step == "" {
		step = strings.TrimSpace(info.Name)
	}
	if step == "" {
		step = strings.TrimSpace(info.ReadyState)
	}

	text := strings.TrimSpace(e.Text)
	if text == "" {
		text = strings.TrimSpace(e.Payload.Text)
	}

	statusCode := e.StatusCode.String()
	if statusCode == "" {
		statusCode = e.Payload.StatusCode.String()
	}

	return BuildLogLine{
		CreatedAt:  created,
		Type:       e.Type,
		Step:       step,
		Entrypoint: strings.TrimSpace(firstNonEmpty(info.Entrypoint, info.Path, info.Type)),
		Text:       text,
		StatusCode: statusCode,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
