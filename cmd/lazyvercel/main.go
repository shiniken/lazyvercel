package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/shiniken/lazyvercel/internal/ui"
	"github.com/shiniken/lazyvercel/internal/vercel"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "lazyvercel: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := vercel.ParseOptions(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if opts.ShowVersion {
		fmt.Printf("lazyvercel %s\n", displayVersion())
		return nil
	}

	store, err := vercel.NewStore(context.Background(), opts)
	if err != nil {
		return err
	}
	if opts.Plain {
		printSummary(store)
		return nil
	}

	program := tea.NewProgram(ui.NewModel(store), tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func displayVersion() string {
	if version != "" && version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func printSummary(store *vercel.Store) {
	for _, project := range store.Projects() {
		deployments := store.Deployments(project)
		if len(deployments) == 0 {
			continue
		}
		context := firstNonEmpty(project.AccountSlug, project.AccountName, project.AccountID)
		if project.LinkedDir != "" {
			context = project.LinkedDir
		}
		fmt.Printf("%s  %s\n", project.Name, context)
		for _, deployment := range deployments {
			title := firstNonEmpty(
				fmt.Sprint(deployment.Meta["githubCommitMessage"]),
				deployment.URL,
				deployment.UID,
			)
			fmt.Printf("  %-12s %-10s %-10s %-18s %s\n",
				deployment.StateLabel(),
				firstNonEmpty(deployment.Target, "preview"),
				age(deployment.CreatedAt),
				shortSHA(fmt.Sprint(deployment.Meta["githubCommitSha"])),
				truncate(oneLine(title), 110),
			)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && value != "<nil>" {
			return value
		}
	}
	return "-"
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) > 12 {
		return sha[:12]
	}
	if sha == "" || sha == "<nil>" {
		return "-"
	}
	return sha
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func truncate(value string, width int) string {
	if width <= 3 || len(value) <= width {
		return value
	}
	return value[:width-3] + "..."
}

func age(ms int64) string {
	if ms <= 0 {
		return "-"
	}
	d := time.Since(time.UnixMilli(ms))
	if d < time.Minute {
		return "now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
