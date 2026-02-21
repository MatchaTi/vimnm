package network

import (
	"fmt"
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
			ssid := strings.TrimSpace(parts[1])

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
		return "[*] " + w.SSID + " (Connected)"
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

func Connect(ssid, password string) error {
	cleanSSID := strings.TrimSpace(ssid)
	cleanPassword := strings.TrimSpace(password)

	if cleanPassword == "" {
		cmdUp := exec.Command("nmcli", "connection", "up", "id", cleanSSID)
		outUp, errUp := cmdUp.CombinedOutput()

		if errUp == nil {
			return nil
		}

		errMsg := strings.TrimSpace(string(outUp))
		if strings.Contains(errMsg, "Secrets were required") || strings.Contains(errMsg, "password") {
			return fmt.Errorf("Password is required for profile '%s'", cleanSSID)
		}

		return fmt.Errorf("Failed to use profile '%s': %s", cleanSSID, errMsg)
	}

	exec.Command("nmcli", "connection", "delete", "id", cleanSSID).Run()

	cmd := exec.Command("nmcli", "device", "wifi", "connect", cleanSSID, "password", cleanPassword)
	out, err := cmd.CombinedOutput()

	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		return fmt.Errorf("%v | Message nmcli: %s", err, errMsg)
	}

	return nil
}

func Disconnect(ssid string) error {
	cleanSSID := strings.TrimSpace(ssid)
	cmd := exec.Command("nmcli", "connection", "down", "id", cleanSSID)
	out, err := cmd.CombinedOutput()

	if err != nil {
		errMsg := strings.TrimSpace(string(out))
		return fmt.Errorf("%v | Message nmcli: %s", err, errMsg)
	}

	return nil
}
