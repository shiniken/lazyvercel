# lazyvercel

`lazyvercel` is a read-only terminal UI for watching Vercel projects, deployments, and build logs from your shell.

It loads your Vercel projects across teams, selects the project tied to the current directory when there is one, and keeps deployment status close without opening the Vercel dashboard.

## Install locally

```bash
go install ./cmd/lazyvercel
```

Once published:

```bash
go install github.com/shiniken/lazyvercel/cmd/lazyvercel@latest
```

## Usage

Authenticate with Vercel first:

```bash
vercel login
```

Optionally link local project directories with Vercel:

```bash
cd ~/code/my-app
vercel link
```

Then run. If you have already run `vercel login`, `lazyvercel` will reuse the Vercel CLI auth token. You can also pass a token explicitly:

```bash
lazyvercel
VERCEL_TOKEN=... lazyvercel --dir ~/code/my-app
```

By default, `lazyvercel` loads the projects visible to your Vercel account. If the current directory is linked to a Vercel project, that project is selected first and marked `cwd`. If the current directory is not linked, the most recently updated project is selected.

The project list shows lightweight latest-deployment context for visible rows. Deployment lists and summaries are fetched lazily so the app does not poll every project on every refresh.

Use `--dir` to pin linked local projects near the top of the project list:

```bash
lazyvercel --dir ~/code/my-app --dir ~/code/admin
```

Useful filters:

```bash
lazyvercel --target production
lazyvercel --branch main
lazyvercel --limit 50
lazyvercel --refresh 10s
lazyvercel --refresh 0
```

The TUI refreshes automatically every 30 seconds by default. When any deployment is queued, initializing, or building, it temporarily polls every 5 seconds so active deploys feel live. Use `--refresh 0` to disable automatic refresh and rely on `r`.

For scriptable output without launching the TUI:

```bash
lazyvercel --dir ~/code/my-app --plain
```

## Keys

- `tab` / `shift+tab`: switch panes
- `j` / `k` or arrows: move selection
- `r`: refresh
- `l`: show build logs for the selected deployment
- `d`: return to deployment detail
- `o`: open the selected deployment URL
- `i`: open the selected deployment in Vercel
- `q`: quit

## Logs

The `l` key fetches build events from Vercel for the selected deployment. This is the useful path for active deployments and failed builds. When auto-refresh is enabled and a build is active, the logs pane refreshes along with the deployment list.

Runtime request logs are a separate Vercel API and are planned as a later pane.

## Roadmap

- Runtime log pane
- Log follow mode for active builds
- Config file for workspace groups
- Git-aware default branch filtering
- Optional guarded actions: cancel, redeploy, promote
