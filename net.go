// utility/net.go
package Utility

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	externalip "github.com/glendc/go-external-ip"
	"github.com/txn2/txeh"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// IPInfo describes a particular IP address (from ipinfo.io).
type IPInfo struct {
	IP       string
	Hostname string
	City     string
	Country  string
	Loc      string
	Org      string
	Postal   string
}

// Ping sends an ICMP echo request to a domain and waits for a reply.
func Ping(domain string) error {
	ipAddr, err := net.ResolveIPAddr("ip4", domain)
	if err != nil {
		return fmt.Errorf("error resolving IP address: %v", err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return fmt.Errorf("error listening for ICMP packets: %v", err)
	}
	defer conn.Close()

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	messageBytes, err := message.Marshal(nil)
	if err != nil {
		return fmt.Errorf("error marshalling ICMP message: %v", err)
	}

	_, err = conn.WriteTo(messageBytes, ipAddr)
	if err != nil {
		return fmt.Errorf("error sending ICMP message: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	responseBytes := make([]byte, 1500)
	_, _, err = conn.ReadFrom(responseBytes)
	if err != nil {
		return fmt.Errorf("error receiving ICMP response: %v", err)
	}
	return nil
}

// MyMacAddr gets the MAC address of the local interface associated with ip.
func MyMacAddr(ip string) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	var currentIP, ifaceName string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if ipnet.IP.String() == ip {
					currentIP = ipnet.IP.String()
					break
				}
			}
		}
	}
	if currentIP == "" {
		return "", errors.New("IP not found on this host")
	}

	interfaces, _ := net.Interfaces()
	for _, interf := range interfaces {
		if addrs, err := interf.Addrs(); err == nil {
			for _, addr := range addrs {
				if strings.Contains(addr.String(), currentIP) {
					ifaceName = interf.Name
					break
				}
			}
		}
	}
	if ifaceName == "" {
		return "", errors.New("no interface for IP found")
	}
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return "", err
	}
	return iface.HardwareAddr.String(), nil
}

// DomainHasIp checks if a DNS lookup for domain resolves to ip.
func DomainHasIp(domain string, ip string) bool {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return false
	}
	for _, ip_ := range ips {
		if ip_.String() == ip {
			return true
		}
	}
	return false
}

// MyIP returns the external IP as seen from outside.
func MyIP() string {
	consensus := externalip.DefaultConsensus(&externalip.ConsensusConfig{Timeout: 500 * time.Millisecond}, nil)
	ip, err := consensus.ExternalIP()
	if err == nil {
		return ip.String()
	}
	return ""
}

// MyIPv6 returns the first non-loopback IPv6 address.
func MyIPv6() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() == nil && ipnet.IP.To16() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("IPv6 address not found")
}

// GetPrimaryIPAddress returns the main non-loopback IPv4 of this machine.
func GetPrimaryIPAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return "", err
			}
			if ip.To4() != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
				return ip.String(), nil
			}
		}
	}
	return "", errors.New("no primary local IP address found")
}

// MyLocalIP returns the local IPv4 for a given MAC.
func MyLocalIP(mac string) (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.HardwareAddr.String() == mac {
			addrs, err := iface.Addrs()
			if err != nil {
				return "", err
			}
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					return "", err
				}
				if ip.To4() != nil && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
					return ip.String(), nil
				}
			}
		}
	}
	return "", errors.New("no local IP found for MAC " + mac)
}

// privateIPCheck checks if an IP is in a private range.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}

// GetIpv4 resolves a hostname into an IPv4 string.
func GetIpv4(address string) (string, error) {
	if strings.Contains(address, ":") {
		address = address[:strings.Index(address, ":")]
	}
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return "", err
	}
	exist, ip, _ := hosts.HostAddressLookup(address)
	if exist {
		return ip, nil
	}
	ips, _ := net.LookupIP(address)
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}
	return "", errors.New("no address found for domain " + address)
}

// IsLocal returns true if a hostname resolves to a private/local IP.
func IsLocal(hostname string) bool {
	if strings.Contains(hostname, ":") {
		hostname = hostname[:strings.Index(hostname, ":")]
	}
	hosts, err := txeh.NewHostsDefault()
	if err != nil {
		return false
	}
	exist, ip, _ := hosts.HostAddressLookup(hostname)
	if exist {
		return privateIPCheck(ip)
	}
	return false
}

// ForeignIP queries ipinfo.io for details about an IP.
func ForeignIP(ip string) (*IPInfo, error) {
	if ip != "" {
		ip += "/" + ip
	}
	resp, err := http.Get("http://ipinfo.io" + ip + "/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ipinfo IPInfo
	if err := json.Unmarshal(data, &ipinfo); err != nil {
		return nil, err
	}
	return &ipinfo, nil
}

// ScanIPs runs `arp -a` and extracts IPv4 addresses.
func ScanIPs() ([]string, error) {
	cmd := exec.Command("arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	re := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	var ips []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		ip := re.FindString(line)
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading output: %w", err)
	}
	return ips, nil
}

// GetHostnameIPMap scans the local network and returns hostname→IP mappings.
func GetHostnameIPMap(localIp string) map[string]string {
	localNetworks := make([]string, 0)
	if localIp != "" {
		if strings.HasPrefix(localIp, "192.168.0.") {
			localNetworks = append(localNetworks, "192.168.0.0/24")
		} else if strings.HasPrefix(localIp, "10.") {
			localNetworks = append(localNetworks, "10.0.0.0/24")
		} else if strings.HasPrefix(localIp, "172.") {
			localNetworks = append(localNetworks, "172.16.0.0/24")
		}
	}
	hostnameIPMap := make(map[string]string)
	for _, netrange := range localNetworks {
		if m, err := getHostnameIPMap(netrange); err == nil {
			for k, v := range m {
				hostnameIPMap[k] = v
			}
		}
	}
	return hostnameIPMap
}

func getHostnameIPMap(localnetwork string) (map[string]string, error) {
	cmd := exec.Command("nmap", "-sn", localnetwork)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running nmap: %v", err)
	}
	awkCmd := exec.Command("awk", "/for/ && $6 != \"\" {gsub(/[()]/, \"\"); print $5, $6}")
	awkCmd.Stdin = strings.NewReader(string(output))

	awkOutput, err := awkCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running awk: %v", err)
	}
	hostnameIPMap := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(awkOutput)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			hostnameIPMap[parts[1]] = parts[0]
		}
	}
	return hostnameIPMap, nil
}

