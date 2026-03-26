package diagnostics

import (
	"fmt"
)

// CheckNetwork tests DNS resolution and gateway ping
func CheckNetwork() CheckResult {
	r := CheckResult{Category: "Network"}

	// DNS check
	dnsOK, dnsMs := checkDNS()
	if dnsOK {
		r.Details = append(r.Details, fmt.Sprintf("DNS resolution: %d ms", dnsMs))
	} else {
		r.Details = append(r.Details, "DNS resolution: failed")
	}

	// Gateway ping
	gwOK, gwMs := checkGateway()
	if gwOK {
		r.Details = append(r.Details, fmt.Sprintf("Gateway ping: %d ms", gwMs))
	} else {
		r.Details = append(r.Details, "Gateway ping: failed")
	}

	switch {
	case !dnsOK && !gwOK:
		r.Status = Critical
		r.Summary = "No network connectivity"
		r.Suggestions = append(r.Suggestions, "Check Wi-Fi or Ethernet connection")
	case !dnsOK:
		r.Status = Warning
		r.Summary = "DNS resolution failing"
		r.Suggestions = append(r.Suggestions, "Check DNS settings or try 1.1.1.1 / 8.8.8.8")
	case !gwOK:
		r.Status = Warning
		r.Summary = "Gateway unreachable"
		r.Suggestions = append(r.Suggestions, "Check router or network configuration")
	case dnsMs > 200 || gwMs > 100:
		r.Status = Warning
		r.Summary = fmt.Sprintf("High latency — DNS %d ms, GW %d ms", dnsMs, gwMs)
	default:
		r.Status = OK
		r.Summary = fmt.Sprintf("DNS %d ms, Gateway %d ms", dnsMs, gwMs)
	}

	return r
}

// checkDNS and checkGateway are in platform-specific files:
// network_helpers_darwin.go — macOS DNS/gateway checks
// network_helpers_linux.go  — Linux DNS/gateway checks
