# lazyvercel

`lazyvercel` is a terminal UI for watching Vercel deployment state from local project directories.

It starts read-only: discover linked projects, list recent deployments, inspect a selected deployment, and open the deployment or Vercel inspector from the keyboard.

## Install locally

```bash
go install ./cmd/lazyvercel
```

Once published:

```bash
go install github.com/shiniken/lazyvercel/cmd/lazyvercel@latest
```

## Usage

Link each project directory with Vercel first:

```bash
cd ~/code/my-app
vercel link
```

Then run. If you have already run `vercel login`, `lazyvercel` will reuse the Vercel CLI auth token. You can also pass a token explicitly:

```bash
lazyvercel --dir ~/code/my-app --dir ~/code/admin
VERCEL_TOKEN=... lazyvercel --dir ~/code/my-app
```

If no `--dir` is provided, `lazyvercel` reads the current directory.

Useful filters:

```bash
lazyvercel --target production
lazyvercel --branch main
lazyvercel --limit 50
```

For scriptable output without launching the TUI:

```bash
lazyvercel --dir ~/code/my-app --plain
```

## Keys

- `tab` / `shift+tab`: switch panes
- `j` / `k` or arrows: move selection
- `r`: refresh
- `o`: open the selected deployment URL
- `i`: open the selected deployment in Vercel
- `q`: quit

## Roadmap

- Build log pane
- Runtime log pane
- Config file for workspace groups
- Git-aware default branch filtering
- Optional guarded actions: cancel, redeploy, promote
