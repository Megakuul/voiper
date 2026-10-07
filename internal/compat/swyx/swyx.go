// Package swyx keeps vendor protocol extensions outside the reusable SIP core.
package swyx

import "strings"

type Capabilities struct {
	Server       string
	SwyxDetected bool
	Presence     string
	Messaging    string
	Directory    string
	Detail       string
}

func Detect(server, presenceMode, messagingMode string) Capabilities {
	c := Capabilities{Server: server, Presence: "standard", Messaging: "sip", Directory: "local", Detail: "Standard SIP; directory contacts can be imported."}
	name := strings.ToLower(server)
	c.SwyxDetected = strings.Contains(name, "swyx") || strings.Contains(name, "netphone")
	if c.SwyxDetected {
		c.Detail = "Swyx server hint detected. Classic PIDF status extensions are supported; support is confirmed when received. SIP messaging can be tried separately; legacy and modern Swyx chat require a verified service contract."
		if presenceMode == "" || presenceMode == "auto" {
			c.Presence = "swyx"
		}
	}
	switch presenceMode {
	case "disabled":
		c.Presence = "disabled"
	case "swyx":
		c.Presence = "swyx"
	case "dialog":
		c.Presence = "dialog"
	}
	switch messagingMode {
	case "disabled":
		c.Messaging = "disabled"
	case "swyx":
		c.Messaging = "unavailable"
	}
	return c
}
