# CLI Reference

Dev Cockpit is a TUI first, but every read-only view has a command-line
equivalent. These are built for piping into other tools, dropping into a
shell script, or getting an answer without leaving what you were doing.

Output is plain text when redirected or piped, and coloured when written to a
terminal, so `devcockpit status > report.txt` gives you clean text with no
escape codes.

## Commands at a glance

| Command | What it does |
| --- | --- |
| `devcockpit` | Launch the interactive TUI |
| `devcockpit status` | Quick system status and health score |
| `devcockpit diag` | Full diagnostics report |
| `devcockpit ps` | Process list, top 15 by CPU |
| `devcockpit services` | Homebrew service status |
| `devcockpit security` | Security feature status |
| `devcockpit cleanup list` | Show cache sizes without deleting |
| `devcockpit cleanup empty-trash` | Empty the trash |
| `devcockpit update` | Update to the latest version |
| `devcockpit uninstall` | Remove Dev Cockpit from the system |

## Global flags

| Flag | Effect |
| --- | --- |
| `--help`, `-h` | Show help |
| `--version`, `-v` | Show version information |
| `--debug` | Launch with debug logging streamed to the console |
| `--logs` | Print the debug log file location |

## status

A one-screen summary: CPU, memory and disk usage, followed by a health score
and the individual checks behind it.

```bash
devcockpit status
```

```
Dev Cockpit v2.1.0 — Quick Status
───────────────────────────────
  CPU                8.9%
  Memory             64.0% (10.2 GB / 16.0 GB)
  Disk               23.9% (350.5 GB free)

  Health Score       ███████████████████░ 95/100 [A]

  ✓ Disk               OK    23.9% used (350.5 GB free)
  ✓ Storage            OK    3 caches, 1.0 GB total
  ✓ Performance        OK    2 heavy process(es), within normal range
  ⚠ Network            WARN  High latency — DNS 1042 ms, GW 45 ms
  ✓ Services           OK    0 running, 0 stopped (of 0)
  ✓ Security           OK    All security features enabled
```

The health score is a weighted roll-up of the six checks. Anything below an
`OK` is worth a look with `devcockpit diag`.

## diag

The long form of `status`. Every check is expanded with the measurements
behind it and concrete suggestions when something is wrong.

```bash
devcockpit diag
```

Reach for this when `status` shows a `WARN` or `CRIT` and you want to know
why rather than just that.

## ps

Process list, sorted and truncated.

```bash
devcockpit ps                      # top 15 by CPU
devcockpit ps --sort mem --top 5   # top 5 by memory
devcockpit ps --sort pid --top 50  # 50 processes by PID
```

| Flag | Values | Default |
| --- | --- | --- |
| `--sort` | `cpu`, `mem`, `pid` | `cpu` |
| `--top` | any positive integer | `15` |

Because output is plain when piped, this composes normally:

```bash
devcockpit ps --sort mem --top 50 | grep node
```

## services

Status of your Homebrew services. Useful for answering "is Postgres actually
running?" without `brew services list` and its startup cost.

```bash
devcockpit services
```

## security

Reports on the platform's security features. On macOS that means the
firewall, FileVault and System Integrity Protection. On Linux it reports the
equivalent available checks.

```bash
devcockpit security
```

This is read-only. Dev Cockpit never changes a security setting from the CLI.

## cleanup

`cleanup` is deliberately split so that inspecting and deleting are different
commands.

```bash
devcockpit cleanup list          # show cache sizes, deletes nothing
devcockpit cleanup empty-trash   # empty the trash
```

`cleanup list` is always safe to run. For anything more selective than
emptying the trash, use the Cleanup module in the TUI, where you can see and
choose each target before it is removed.

## update

```bash
devcockpit update            # check, confirm, then install
devcockpit update --check    # report only, change nothing
devcockpit update --force    # skip the confirmation prompt
```

Updates are verified before they are installed. Dev Cockpit downloads the
published SHA-256 checksum alongside the binary and refuses to install if it
does not match, if the checksum is missing, or if it cannot be read. The old
binary is backed up first and restored automatically if the replacement fails
to run.

## uninstall

```bash
devcockpit uninstall           # confirm, then remove
devcockpit uninstall --force   # no confirmation
```

Removes the binary. Your configuration directory at `~/.devcockpit` is left
in place, so reinstalling keeps your settings. Delete it by hand if you want
a clean slate:

```bash
rm -rf ~/.devcockpit
```

## Files

| Path | Contents |
| --- | --- |
| `~/.devcockpit/config.yaml` | Configuration, see [Configuration](/configuration) |
| `~/.devcockpit/debug.log` | Debug log, written when `--debug` is used |
| `~/.devcockpit/data/` | Metrics history database |

## See also

- [Modules](/modules) for what each TUI screen does
- [Configuration](/configuration) for every setting and its default
- [Troubleshooting](/troubleshooting) when something misbehaves
