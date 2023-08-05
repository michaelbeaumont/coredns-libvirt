package libvirt

import (
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"

	"golang.org/x/exp/maps"
)

const stateDir = "/var/lib/libvirt/dnsmasq"

type MacEntry struct {
	Domain string   `json:"domain"`
	MACs   []string `json:"macs"`
}

type AddressEntry struct {
	MACAddress string `json:"mac-address"`
	IPAddress  string `json:"ip-address"`
}

type Bridge struct {
	Addresses []AddressEntry
	MACs      []MacEntry
}

func processEntries(name string, bridges []Bridge) []net.IP {
	var addresses []net.IP

	for _, bridge := range bridges {
		macs := map[string]struct{}{}
		for _, domainMacs := range bridge.MACs {
			if domainMacs.Domain != name {
				continue
			}
			for _, mac := range domainMacs.MACs {
				macs[mac] = struct{}{}
			}
		}

		for _, address := range bridge.Addresses {
			if _, ok := macs[address.MACAddress]; ok {
				addresses = append(addresses, net.ParseIP(address.IPAddress))
			}
		}
	}

	return addresses
}

func findGuestIPs(name string) ([]net.IP, error) {
	files, err := os.ReadDir(stateDir)
	if err != nil {
		return nil, err
	}

	bridges := map[string]Bridge{}

	for _, file := range files {
		ext := path.Ext(file.Name())
		bridgeName := strings.TrimSuffix(file.Name(), ext)
		bridge := bridges[bridgeName]
		switch ext {
		case ".macs":
			contents, err := os.ReadFile(path.Join(stateDir, file.Name()))
			if err != nil {
				return nil, err
			}
			var macs []MacEntry
			if err := json.Unmarshal(contents, &macs); err != nil {
				return nil, err
			}
			bridge.MACs = macs
		case ".status":
			contents, err := os.ReadFile(path.Join(stateDir, file.Name()))
			if err != nil {
				return nil, err
			}
			var addresses []AddressEntry
			if err := json.Unmarshal(contents, &addresses); err != nil {
				return nil, err
			}
			bridge.Addresses = addresses
		default:
			continue
		}
		bridges[bridgeName] = bridge
	}

	return processEntries(name, maps.Values(bridges)), nil
}
