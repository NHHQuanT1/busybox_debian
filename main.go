package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	pathProcKernel       = "/proc/sys/kernel/osrelease"
	pathProcUptime       = "/proc/uptime"
	pathProcCPUInfo      = "/proc/cpuinfo"
	pathProcMemInfo      = "/proc/meminfo"
	pathProcNetRoute     = "/proc/net/route"
	pathDMIProductSerial = "/sys/class/dmi/id/product_serial"
	pathDMIProductName   = "/sys/class/dmi/id/product_name"
	pathDMIBiosVersion   = "/sys/class/dmi/id/bios_version"
	pathResolvConf       = "/etc/resolv.conf"
)

func readSysFile(path string) string {
	content, err := os.ReadFile(path)
	if path == pathDMIProductSerial {
		if err != nil {
			return ""
		}
		return "Not Specified"
	} else {
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(content))
	}

}

func getSystemUptime(path string) string {
	data := readSysFile(path)
	if data == "" {
		return ""
	}
	parts := strings.Fields(data)
	if len(parts) == 0 {
		return ""
	}
	totalSecondsFloat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return ""
	}
	uptime := int64(totalSecondsFloat)
	const (
		Minute = 60
		Hour   = 60 * Minute
		Day    = 24 * Hour
		Week   = 7 * Day
	)

	weeks := uptime / Week
	uptime %= Week

	days := uptime / Day
	uptime %= Day

	hours := uptime / Hour
	uptime %= Hour

	minutes := uptime / Minute
	uptime %= Minute
	var resultParts []string

	if weeks > 0 {
		resultParts = append(resultParts, fmt.Sprintf("%d week%s", weeks, plural(weeks)))
	}
	if days > 0 {
		resultParts = append(resultParts, fmt.Sprintf("%d day%s", days, plural(days)))
	}
	if hours > 0 {
		resultParts = append(resultParts, fmt.Sprintf("%d hour%s", hours, plural(hours)))
	}
	if minutes > 0 {
		resultParts = append(resultParts, fmt.Sprintf("%d minute%s", minutes, plural(minutes)))
	}

	if len(resultParts) == 0 {
		return fmt.Sprintf("%d second%s", uptime, plural(uptime))
	}

	return strings.Join(resultParts, ", ")
}

func plural(value int64) string {
	if value > 1 {
		return "s"
	}
	return ""
}

func getCPUModel(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func getTotalMemory(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var kb float64
				fmt.Sscanf(fields[1], "%f", &kb)
				return fmt.Sprintf("%.4f GB", kb/1024/1024)
			}
			return ""
		}
	}
	return ""
}

func getPrimaryNetwork() (string, string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					return v.String(), i.Name
				}
			}
		}
	}
	return "", ""
}

func getDefaultGateway(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "00000000" {
			gwHex := fields[2]
			if len(gwHex) == 8 {
				ip := net.IP(make([]byte, 4))
				fmt.Sscanf(gwHex, "%02x%02x%02x%02x", &ip[3], &ip[2], &ip[1], &ip[0])
				return ip.String()
			}
		}
	}
	return ""
}

func getDNS(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	var dnsList []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dnsList = append(dnsList, parts[1])
			}
		}
	}
	return strings.Join(dnsList, ",")
}
func main() {
	m := map[string]string{}
	m["hostname"], _ = os.Hostname()
	m["kernel"] = readSysFile(pathProcKernel)
	m["serial"] = readSysFile(pathDMIProductSerial)
	m["model"] = readSysFile(pathDMIProductName)
	m["bios"] = readSysFile(pathDMIBiosVersion)
	m["uptime"] = getSystemUptime(pathProcUptime)
	m["cpu"] = getCPUModel(pathProcCPUInfo)
	m["mem"] = getTotalMemory(pathProcMemInfo)
	m["gateway"] = getDefaultGateway(pathProcNetRoute)
	m["dns"] = getDNS(pathResolvConf)
	ip, nic := getPrimaryNetwork()
	m["mgmt_ip"] = ip
	m["mgmt_nic"] = nic
	keys := []string{
		"hostname",
		"kernel",
		"serial",
		"model",
		"bios",
		"uptime",
		"cpu",
		"mem",
		"mgmt_ip",
		"mgmt_nic",
		"gateway",
		"dns",
	}

	for _, k := range keys {
		fmt.Printf("%s: %s\n", k, m[k])
	}

}
