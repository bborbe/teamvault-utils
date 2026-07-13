---
description: Guided first-time setup of teamvault-cli (install, config, login hint, verify)
allowed-tools: Bash(command -v:*), Bash(uname:*), Bash(brew:*), Bash(zsh:*), Bash(mkdir:*), Bash(teamvault-cli:*), Read, Write, AskUserQuestion
---

# Set up teamvault-cli

Walk a new user through first-time setup of `teamvault-cli` against a TeamVault instance, end to end. Optimized for the two things people get wrong: the **`user` is usually your directory/login username** (not an email), and on macOS **Homebrew's bin must be on the PATH for non-interactive shells** (direnv / `.envrc` run there).

**Never handle the password.** `teamvault-cli login` prompts for it interactively; that value must never enter this session or the transcript. This command configures + instructs + verifies — it does NOT run `login`.

Run the steps in order. Stop and report if a step fails; don't paper over it.

## Step 1 — Homebrew (macOS)

Only on macOS (`uname` = `Darwin`). Check:

```bash
command -v brew
```

If missing, tell the user to install it themselves (it needs `sudo` and is interactive — do not run it for them):

> Homebrew isn't installed. In a plain terminal, run:
> `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
> then re-run `/teamvault-cli:setup`.

On Linux, skip Homebrew — use the release binary or `go install` (see Step 2).

## Step 2 — Install teamvault-cli

macOS (recommended):

```bash
brew install seibert-data/tap/teamvault-cli   # update later: brew upgrade teamvault-cli
```

**Linux / Go:** see `docs/getting-started.md` §1 for the platform-specific commands (prebuilt release binary, or `go install github.com/Seibert-Data/teamvault-cli/v5@latest`). Then skip to Step 4 — Homebrew's PATH steps below are macOS-only.

## Step 3 — PATH check (interactive AND non-interactive)

This is the step that silently breaks `.envrc`/direnv later. Homebrew's bin (`/opt/homebrew/bin` on Apple Silicon, `/usr/local/bin` on Intel) must be on the PATH for **both** an interactive login shell and a non-interactive shell.

```bash
brew_bin="$(brew --prefix 2>/dev/null)/bin"
echo "brew bin: $brew_bin"
echo -n "interactive login shell: "; zsh -lic 'command -v teamvault-cli' 2>/dev/null || echo "NOT FOUND"
echo -n "non-interactive shell:   "; zsh -c 'command -v teamvault-cli' 2>/dev/null || echo "NOT FOUND"
```

- **Both found** → PATH is fine, continue.
- **Interactive found, non-interactive NOT** → the brew shellenv is only in an interactive rc file. direnv/`.envrc` (non-interactive) won't see `teamvault-cli`. Tell the user to add it where non-interactive shells read it:
  > Add this line to `~/.zshenv` (sourced by every shell, including non-interactive) — or ensure your `~/.zprofile` `brew shellenv` line is present and your terminal starts a login shell:
  > `eval "$(/opt/homebrew/bin/brew shellenv)"`  *(use `/usr/local/bin` on Intel)*
- **Neither found** → the install didn't land on PATH at all; have the user add the `brew shellenv` line above to `~/.zprofile` and open a new terminal, then re-run this command.

## Step 4 — Configure

Determine the two values:

- **URL** — the TeamVault base URL (e.g. `https://teamvault.example.com`), **no trailing slash** — the CLI now strips one defensively, but write it clean. Ask the user for their instance URL.
- **`user`** — the TeamVault username, typically your **directory/LDAP login name, NOT an email address** (some orgs derive it from your email's local part). Ask via `AskUserQuestion` if unknown; do not guess.

Create the XDG-default dir (`mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/teamvault-cli"`), then write `config.json` there (substitute the LDAP username; never put a password in the file):

```json
{
    "url": "https://teamvault.example.com",
    "user": "<your-username>"
}
```

If `~/.teamvault.json` already exists (legacy default), point that out — the XDG file takes precedence.

## Step 5 — Login (user runs this, NOT Claude)

`teamvault-cli login` verifies the password and stores it in the macOS Keychain. It is **interactive** and prompts for the password — so the user must run it themselves in a plain terminal. Tell them, verbatim:

> In a **plain terminal** (not here in Claude), run:
> `teamvault-cli login`
> Enter your TeamVault password when prompted. It's stored in your macOS Keychain; you won't need to type it again.

Do not run `login` yourself and do not ask the user to paste the password here. Wait for them to confirm they've done it.

*(On Linux/Windows there's no Keychain backend yet — the user supplies the password via `TEAMVAULT_PASS` or the config file instead. Note this only if they're not on macOS.)*

## Step 6 — Verify

After the user confirms login, test end to end with a **non-secret** call — ask for a TeamVault key they can read (the alphanumeric ID from a secret's URL), then check exit status without printing any value:

```bash
teamvault-cli username --teamvault-key <KEY> >/dev/null && echo "OK — teamvault-cli is set up" || echo "FAILED — see error above"
```

- **OK** → setup complete. Mention the `.envrc` pattern for project use: `export FOO=$(teamvault-cli password --teamvault-key <KEY>)`.
- **403 / authentication failed** → the password isn't in the Keychain for this instance. Have the user re-run `teamvault-cli login` in a plain terminal (the error message says so since v5.5.2).
- **key not found / 404** → wrong `user` (not the LDAP name) or wrong key. Recheck Step 4's username.

Never print, echo, or log a resolved secret value to confirm success — the exit status is the confirmation.
