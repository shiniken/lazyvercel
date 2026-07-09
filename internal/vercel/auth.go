package vercel

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type cliAuthFile struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

func ResolveToken(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if token := os.Getenv("VERCEL_TOKEN"); token != "" {
		return token, nil
	}

	auth, path, err := readCLIAuth()
	if err != nil {
		return "", fmt.Errorf("missing Vercel token; set VERCEL_TOKEN, pass --token, or run `vercel login`")
	}
	if auth.Token == "" {
		return "", fmt.Errorf("%s does not contain a Vercel token; run `vercel login`", path)
	}
	if auth.ExpiresAt > 0 && time.Now().Unix() >= auth.ExpiresAt {
		return "", fmt.Errorf("Vercel CLI token in %s is expired; run `vercel login`", path)
	}
	return auth.Token, nil
}

func readCLIAuth() (cliAuthFile, string, error) {
	for _, path := range candidateAuthPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var auth cliAuthFile
		if err := json.Unmarshal(data, &auth); err != nil {
			return cliAuthFile{}, path, err
		}
		return auth, path, nil
	}
	return cliAuthFile{}, "", os.ErrNotExist
}

func candidateAuthPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	paths := []string{
		filepath.Join(home, ".vercel", "auth.json"),
	}

	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, filepath.Join(home, "Library", "Application Support", "com.vercel.cli", "auth.json"))
	case "linux":
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			paths = append(paths, filepath.Join(configHome, "com.vercel.cli", "auth.json"))
		}
		paths = append(paths, filepath.Join(home, ".config", "com.vercel.cli", "auth.json"))
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			paths = append(paths, filepath.Join(appData, "com.vercel.cli", "auth.json"))
		}
	}

	return paths
}
