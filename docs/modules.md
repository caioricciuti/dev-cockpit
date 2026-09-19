# Modules

Dev Cockpit is thirteen modules behind one window. This page covers how
navigation works, then what each module does.

## The one thing to understand first

The interface has **two levels**, and most early confusion comes from not
realising which one you are in.

**Browse level.** You are moving between modules. `Tab` and the arrow keys
move from one to the next. Keys like `q` and `?` belong to the application.

**Focused level.** You have pressed `Enter` on a module, and from then on
**the module receives every key**. `j` scrolls its list rather than doing
anything global. `Esc` hands control back to the browse level.

If a key seems to do nothing, you are almost certainly at the wrong level.
Press `Esc` until you are back to browsing.

One exception worth knowing: when a module has a confirmation prompt or input
field open, `Esc` closes that first and leaves you focused. Press it again to
unfocus.

## Global keys

Available at the browse level.

| Key | Action |
| --- | --- |
| `Tab`, `→` | Next module |
| `Shift+Tab`, `←` | Previous module |
| `Home` | First module (Dashboard) |
| `End` | Last module (Support) |
| `Enter` | Focus the current module |
| `Esc` | Leave a focused module |
| `?` | Toggle the help overlay |
| `l` | Toggle the log overlay |
| `q` | Quit |
| `Ctrl+C` | Quit from anywhere, including inside a focused module |

`Ctrl+C` is the only key that always works. Everything else is handed to the
module while one is focused.

## Common keys inside modules

Most modules follow the same conventions.

| Key | Action |
| --- | --- |
| `j`, `↓` | Move down |
| `k`, `↑` | Move up |
| `g` / `G` | Jump to top / bottom |
| `Enter` | Select or execute |
| `r` | Refresh |
| `/` | Search, where the module supports it |
| `Esc` | Close a prompt, or unfocus |
| `1`–`5` | Switch between a module's own sub-views |

Number keys switch sub-views **within** a module. They do not jump between
modules.

## The modules

Listed in the order they appear. `Home` takes you to the first, `End` to the
last.

### 1. Dashboard

Real-time CPU, GPU, memory, disk and network monitoring, with history graphs
drawn from the local metrics database.

Sub-views are on `1`–`5`. Graph height and which detail panels are shown are
configurable, see [Configuration](/configuration#dashboard).

### 2. Processes

An interactive process manager.

| Key | Action |
| --- | --- |
| `/` | Search by name |
| `s` | Cycle the sort field |
| `S` | Reverse the sort direction |
| `x` | Kill the selected process |
| `g` / `G` | Jump to top / bottom |

`x` always asks for confirmation before sending a signal. Nothing is killed
on a single keypress.

### 3. Services

Homebrew services and detected listening ports in one list, so you can see
whether Postgres or Redis is actually up without leaving the app. `r`
refreshes.

### 4. Quick Actions

One-keypress maintenance tasks: flushing caches, restarting common services
and similar routine jobs. Select with `Enter`. Anything that needs elevated
rights will prompt for it.

### 5. Cleanup

Reclaim disk space from caches, logs, trash and downloads.

| Key | Action |
| --- | --- |
| `Enter` | Toggle a target for cleaning |
| `a` | Select all |
| `n` | Select none |
| `r` | Rescan sizes |

Targets are scanned and shown with their sizes before anything is deleted,
and you choose them individually. If you only want the trash emptied, there
is a CLI equivalent: `devcockpit cleanup empty-trash`.

### 6. Packages

Homebrew, npm and other detected package managers in one place. Lists
outdated packages and lets you update them. `o` filters to outdated, `u`
updates.

Detection depends on the tools being on your `PATH`. If a manager you use is
missing, see
[Troubleshooting](/troubleshooting#package-manager-issues).

### 7. System

Hardware details, OS information and system diagnostics. Sub-views on
`1`–`4`, and `Tab` cycles them.

### 8. Docker

Container management: running and stopped containers, images, and cleanup of
reclaimable space. Works with Docker Desktop, OrbStack and anything else
exposing a standard Docker socket. The socket path is configurable, see
[Configuration](/configuration#docker).

### 9. Network

Interface analysis and connectivity diagnostics, in five sub-views.

| Key | View |
| --- | --- |
| `1` | Overview, interfaces and gateway |
| `2` | Ports, what is listening |
| `3` | Diagnostics, ping, traceroute and DNS |
| `4` | Quality, throughput and latency |
| `5` | Tools, whois |

`Tab` and `Shift+Tab` cycle the views. The Quality view depends on the
`networkQuality` tool and is skipped automatically when unavailable.

In the Diagnostics and Tools views, `Enter` opens an input field for a target
host. Targets accept host names, IPv4 and IPv6 only.

### 10. Security

Firewall, FileVault and System Integrity Protection status on macOS, with the
equivalent checks on Linux. Read-only: Dev Cockpit reports on these settings
and never changes them. `r` re-runs the checks.

### 11. Diagnostics

The health score and the checks behind it, each expandable with `Enter` for
the measurements and suggested fixes. Same data as `devcockpit diag`.

### 12. Logs

A multi-source log aggregator. Sub-views on `1`–`4` select the source, `/`
searches, `f` follows new lines and `c` clears the view. Additional log files
can be added, see [Configuration](/configuration#logs).

### 13. Support

Project sponsorship and support links. No system access.

## See also

- [CLI Reference](/cli) for the command-line equivalents
- [Configuration](/configuration) for per-module settings
- [Troubleshooting](/troubleshooting) when a module misbehaves
