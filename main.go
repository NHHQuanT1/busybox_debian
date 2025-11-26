package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func firstField(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return s[:i]
	}
	return s
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func runQuiet(ctx context.Context, name string, args ...string) string {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	return strings.TrimSpace(string(out))
}
func main() {
	m := map[string]string{}
	m["hostname"] = runQuiet(context.Background(), "hostnamectl", "--static")
	m["kernel"] = runQuiet(context.Background(), "uname", "-r")
	m["serial"] = runQuiet(context.Background(), "dmidecode", "-s", "system-serial-number")
	m["model"] = runQuiet(context.Background(), "dmidecode", "-s", "system-product-name")
	//m["bios"] = runQuiet(context.Background(), "bash", "-lc", "dmidecode -t bios | sed -n '1,6p' | tr -s ' '")
	m["bios"] = runQuiet(context.Background(), "busybox", "sh", "-c", "dmidecode -t bios | sed -n '1,6p' | tr -s ' '")
	m["uptime"] = runQuiet(context.Background(), "uptime", "-p")
	//m["cpu"] = runQuiet(context.Background(), "bash", "-lc", "lscpu | awk -F: '/Model name/ {sub(/^ +/,\"\",$2); print $2; exit}'")
	m["cpu"] = runQuiet(context.Background(), "busybox", "sh", "-c", "lscpu | awk -F: '/Model name/ {sub(/^ +/,\"\",$2); print $2; exit}'")
	//m["mem"] = runQuiet(context.Background(), "bash", "-lc", "awk '/MemTotal/ {print $2/1024/1024 \" GB\"}' /proc/meminfo")
	m["mem"] = runQuiet(context.Background(), "busybox", "sh", "-c", "awk '/MemTotal/ {print $2/1024/1024 \" GB\"}' /proc/meminfo")
	m["mgmt_ip"] = firstField(runQuiet(context.Background(), "bash", "-lc", "ip -o -4 addr show scope global | awk '{print $4}' | head -n1"))
	m["mgmt_nic"] = firstField(runQuiet(context.Background(), "bash", "-lc", "ip -o -4 addr show scope global | awk '{print $2}' | head -n1"))
	m["gateway"] = firstField(runQuiet(context.Background(), "bash", "-lc", "ip route | awk '/default/ {print $3; exit}'"))
	m["dns"] = runQuiet(context.Background(), "bash", "-lc", "awk '/^nameserver/ {print $2}' /etc/resolv.conf | paste -sd, -")
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
