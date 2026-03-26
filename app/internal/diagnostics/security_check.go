package diagnostics

// secStatus holds the result of a single security feature check
type secStatus struct {
	enabled bool
	label   string
	fix     string
}

// CheckSecurity is implemented in platform-specific files:
// security_check_darwin.go — macOS (Firewall, FileVault, SIP, Gatekeeper)
// security_check_linux.go  — Linux (UFW/iptables, LUKS)
