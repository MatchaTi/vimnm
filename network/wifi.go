package network

import (
	"os/exec"
	"strings"
)

type Wifi struct {
	Active   bool
	SSID     string
	Signal   string
	Security string
}

func GetNetworks() []Wifi {
	out, err := exec.Command("nmcli", "-t", "-f", "ACTIVE,SSID,SIGNAL,SECURITY", "dev", "wifi").Output()
	if err != nil {
		return nil
	}

	var networks []Wifi
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ":")

		if len(parts) >= 4 {
			ssid := parts[1]

			if ssid == "" {
				continue
			}

			networks = append(networks, Wifi{
				Active:   parts[0] == "yes",
				SSID:     ssid,
				Signal:   parts[2],
				Security: parts[3],
			})
		}
	}

	return networks
}

func (w Wifi) Title() string {
	if w.Active {
		return "[*]" + w.SSID + "(Connected)"
	}

	return w.SSID
}

func (w Wifi) Description() string {
	sec := w.Security

	if sec == "" {
		sec = "Open"
	}

	return "Signal: " + w.Signal + "% | Security: " + sec
}

func (w Wifi) FilterValue() string {
	return w.SSID
}
