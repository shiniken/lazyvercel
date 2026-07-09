package vercel

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Options struct {
	Dirs        []string
	Token       string
	Limit       int
	Target      string
	Branch      string
	Plain       bool
	ShowVersion bool
}

func ParseOptions(args []string) (Options, error) {
	var dirs multiFlag
	opts := Options{
		Limit:  20,
		Target: "",
	}

	fs := flag.NewFlagSet("lazyvercel", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Var(&dirs, "dir", "linked Vercel project directory; can be repeated")
	fs.StringVar(&opts.Token, "token", "", "Vercel access token; defaults to VERCEL_TOKEN, then Vercel CLI auth")
	fs.IntVar(&opts.Limit, "limit", opts.Limit, "deployments to fetch per project")
	fs.StringVar(&opts.Target, "target", opts.Target, "filter deployments by target, such as production or preview")
	fs.StringVar(&opts.Branch, "branch", opts.Branch, "filter deployments by git branch")
	fs.BoolVar(&opts.Plain, "plain", false, "print deployments once and exit")
	fs.BoolVar(&opts.ShowVersion, "version", false, "print version and exit")

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: lazyvercel [flags]\n\n")
		fmt.Fprintf(fs.Output(), "Flags:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nExamples:\n")
		fmt.Fprintf(fs.Output(), "  lazyvercel --dir ~/code/app --dir ~/code/admin\n")
		fmt.Fprintf(fs.Output(), "  VERCEL_TOKEN=... lazyvercel --target production\n")
	}

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	opts.Dirs = dirs
	if len(opts.Dirs) == 0 {
		opts.Dirs = []string{"."}
	}

	if opts.Limit < 1 {
		return Options{}, fmt.Errorf("--limit must be greater than zero")
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}
	opts.Target = strings.TrimSpace(opts.Target)
	opts.Branch = strings.TrimSpace(opts.Branch)

	return opts, nil
}

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("directory cannot be empty")
	}
	*m = append(*m, value)
	return nil
}
