# Configuration

Dev Cockpit reads `~/.devcockpit/config.yaml` at startup. The file is
optional: every setting has a default, and anything you leave out keeps it.

If `~/.devcockpit` is not writable, for example under a sandboxed shell, Dev
Cockpit falls back to a temporary location and carries on rather than
failing.

## A minimal file

You only need the keys you want to change.

```yaml
theme: dark
update_interval: 1000

ui:
  mouse_enabled: true

modules:
  dashboard:
    graph_height: 10
```

## General

```yaml
theme: dark
update_interval: 1000
enable_telemetry: false
log_level: info
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `theme` | string | `dark` | Overall theme |
| `update_interval` | int | `1000` | Metric refresh in **milliseconds** |
| `enable_telemetry` | bool | `false` | Off by default and stays off unless you turn it on |
| `log_level` | string | `info` | `debug`, `info`, `warn` or `error` |

`update_interval` is the one to lower if you want a livelier dashboard, and
to raise if Dev Cockpit is using more CPU than you would like. See
[Troubleshooting](/troubleshooting#high-cpu-usage).

## UI

```yaml
ui:
  color_scheme: cyberpunk
  animation_speed: 60
  show_fps: false
  mouse_enabled: true
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `color_scheme` | string | `cyberpunk` | Palette used across modules |
| `animation_speed` | int | `60` | Target frames per second |
| `show_fps` | bool | `false` | Show a frame counter, useful when debugging rendering |
| `mouse_enabled` | bool | `true` | Mouse reporting inside the TUI |

Turn `mouse_enabled` off if you would rather your terminal handle selection
and copying itself.

## Modules

### Dashboard

```yaml
modules:
  dashboard:
    refresh_rate: 1
    show_cpu_details: true
    show_mem_details: true
    show_disk_details: true
    graph_height: 10
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `refresh_rate` | int | `1` | Dashboard refresh in **seconds** |
| `show_cpu_details` | bool | `true` | Expanded CPU panel |
| `show_mem_details` | bool | `true` | Expanded memory panel |
| `show_disk_details` | bool | `true` | Expanded disk panel |
| `graph_height` | int | `10` | Height of history graphs, in rows |

Note the unit difference: `refresh_rate` here is in seconds, while the
top-level `update_interval` is in milliseconds.

### Docker

```yaml
modules:
  docker:
    socket_path: /var/run/docker.sock
    show_all_containers: false
    auto_refresh: true
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `socket_path` | string | `/var/run/docker.sock` | Point this at OrbStack or a rootless socket if yours differs |
| `show_all_containers` | bool | `false` | Include stopped containers |
| `auto_refresh` | bool | `true` | Poll for changes |

If Docker is not detected, the socket path is the first thing to check. See
[Troubleshooting](/troubleshooting#docker-not-detected).

### Network

```yaml
modules:
  network:
    packet_capture: false
    port_scan_timeout: 2
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `default_interface` | string | detected | Detected per platform at startup |
| `packet_capture` | bool | `false` | Off by default; capture needs elevated rights |
| `port_scan_timeout` | int | `2` | Seconds before a port probe gives up |

### Security

```yaml
modules:
  security:
    scan_interval: 300
    check_firewall: true
    check_filevault: true
    check_sip: true
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `scan_interval` | int | `300` | Seconds between re-checks |
| `check_firewall` | bool | `true` | Firewall status |
| `check_filevault` | bool | `true` | FileVault, macOS |
| `check_sip` | bool | `true` | System Integrity Protection, macOS |

These control whether a check runs, not what it reports. Dev Cockpit never
changes a security setting.

### Logs

```yaml
modules:
  logs:
    max_lines: 500
    custom_log_paths:
      - /var/log/nginx/error.log
      - ~/projects/api/logs/app.log
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `max_lines` | int | `500` | Lines held per source |
| `custom_log_paths` | list | empty | Extra files to aggregate |

`custom_log_paths` is how you bring your own application logs alongside the
system ones.

## System

```yaml
system:
  command_timeout: 30
  max_retries: 3
  sudo_command: sudo
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `command_timeout` | int | `30` | Seconds before an external command is killed |
| `max_retries` | int | `3` | Retries for a failed command |
| `sudo_command` | string | `sudo` | Change to `doas` or similar if you use one |

Raise `command_timeout` if a slow disk makes cleanup scans time out. See
[Troubleshooting](/troubleshooting#operation-timeout).

## Storage

```yaml
storage:
  max_history_days: 30
  compress_old_data: true
```

| Key | Type | Default | Notes |
| --- | --- | --- | --- |
| `data_dir` | string | `~/.devcockpit/data` | Metrics database location |
| `max_history_days` | int | `30` | History older than this is pruned at startup |
| `compress_old_data` | bool | `true` | Compress history as it ages |

Metrics history is stored locally in SQLite and never leaves your machine.

## Full reference

Everything in one file, at its default:

```yaml
theme: dark
update_interval: 1000
enable_telemetry: false
log_level: info

ui:
  color_scheme: cyberpunk
  animation_speed: 60
  show_fps: false
  mouse_enabled: true

modules:
  dashboard:
    refresh_rate: 1
    show_cpu_details: true
    show_mem_details: true
    show_disk_details: true
    graph_height: 10
  docker:
    socket_path: /var/run/docker.sock
    show_all_containers: false
    auto_refresh: true
  network:
    packet_capture: false
    port_scan_timeout: 2
  security:
    scan_interval: 300
    check_firewall: true
    check_filevault: true
    check_sip: true
  logs:
    max_lines: 500
    custom_log_paths: []

system:
  command_timeout: 30
  max_retries: 3
  sudo_command: sudo

storage:
  max_history_days: 30
  compress_old_data: true
```

## See also

- [Modules](/modules) for what each setting affects
- [CLI Reference](/cli) for file locations
- [Troubleshooting](/troubleshooting)
