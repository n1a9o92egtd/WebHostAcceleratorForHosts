package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func GetHostsFilePath() (string, error) {
	hosts_dst_file := "/etc/hosts"
	if runtime.GOOS == "windows" {
		win_sys := os.Getenv("windir")
		hosts_dst_file := filepath.Join(win_sys, "System32", "drivers", "etc", "hosts")
		_, err := os.Stat(hosts_dst_file)
		if os.IsExist(err) == false {
			return hosts_dst_file, err
		}
		return "", err
	} else if runtime.GOOS == "linux" {
		_, err := os.Stat(hosts_dst_file)
		if os.IsExist(err) == false {
			return hosts_dst_file, err
		}
		return "", err
	} else if runtime.GOOS == "darwin" {
		_, err := os.Stat(hosts_dst_file)
		if os.IsExist(err) == false {
			return hosts_dst_file, err
		}
		return "", err
	} else {
		return "", nil
	}
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func BackupHostsFile(hosts_file string) (bool, error) {
	backup_ok := true
	hosts_file_bak := hosts_file + ".bak"
	// _, err := os.Stat(hosts_file_bak)
	if _, err := os.Stat(hosts_file_bak); err != nil {
		if os.IsExist(err) {
			return backup_ok, err
		}
	}
	_, err := CopyFile(hosts_file_bak, hosts_file)
	if err == nil {
		return backup_ok, err
	}
	return false, err
}

func IsAdminRunning() bool {
	if runtime.GOOS == "windows" {
		_, err := exec.Command("net", "session").Output()
		if err != nil {
			fmt.Println("Please start the WebHostAcceleratorForHosts as administrator!")
			os.Exit(1)
		}
		// var sid *windows.SID
		// err := windows.AllocateAndInitializeSid(
		// 	&windows.SECURITY_NT_AUTHORITY,
		// 	2,
		// 	windows.SECURITY_BUILTIN_DOMAIN_RID,
		// 	windows.DOMAIN_ALIAS_RID_ADMINS,
		// 	0, 0, 0, 0, 0, 0,
		// 	&sid)
		// if err != nil {
		// 	fmt.Println("Please start the WebHostAcceleratorForHosts as root or sudo:", err)
		// 	os.Exit(1)
		// }
		// token := windows.Token(0)
		// member, err := token.IsMember(sid)
		// if err != nil {
		// 	fmt.Println("Please start the WebHostAcceleratorForHosts as root or sudo:", err)
		// 	os.Exit(1)
		// }
		// if member != true {
		// 	fmt.Println("Please start the WebHostAcceleratorForHosts as root or sudo!")
		// 	os.Exit(1)
		// }
		return true
	} else {
		if os.Getuid() != 0 {
			fmt.Println("Please start the WebHostAcceleratorForHosts as root or sudo!")
			os.Exit(1)
		}
		return true
	}

}

func HijackGithubHosts(hosts_file string, github_host string, domain string) {
	content, err := ioutil.ReadFile(hosts_file)
	if err != nil {
		fmt.Println("WebHostAcceleratorForHosts Read Hosts ERROR:", err)
		return
	}
	var lines []string
	if runtime.GOOS == "windows" {
		lines = strings.Split(string(content), "\r\n")
	} else {
		lines = strings.Split(string(content), "\n")
	}

	is_exist := false
	var new_lines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			new_lines = append(new_lines, line)
			continue
		}
		parts := strings.Fields(line)
		d := parts[1]
		if d == domain {
			is_exist = true
			parts[0] = github_host
			new_lines = append(new_lines, strings.Join(parts, " "))
			continue
		}
		new_lines = append(new_lines, line)
	}
	if !is_exist {
		// _, ip_net, err := net.ParseCIDR(github_host)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
		new_lines = append(new_lines, github_host+"	"+domain)
	}

	file, err := os.OpenFile(hosts_file, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, line := range new_lines {
		if runtime.GOOS == "windows" {
			_, err = writer.WriteString(line + "\r\n")
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			_, err = writer.WriteString(line + "\n")
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	writer.Flush()
}

func OpenLocalWebBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\nWebHostAcceleratorForHosts 140.82.121.4")
		return
	}
	OpenLocalWebBrowser("https://tool.chinaz.com/speedworld/github.com")
	// fmt.Println("\r\n\r\n\r\nOpen https://tool.chinaz.com/speedworld/github.com\r\n\r\n\r\n")
	IsAdminRunning()
	hosts_file, err := GetHostsFilePath()
	if err != nil {
		fmt.Println("WebHostAcceleratorForHosts Find Hosts ERROR:", err)
		return
	}
	_, err = BackupHostsFile(hosts_file)
	if err != nil {
		fmt.Println("WebHostAcceleratorForHosts Bakcup Hosts ERROR:", err)
		return
	}
	fmt.Println("\r\n\r\n\r\nWebHostAcceleratorForHosts Backup Hosts File Success:", hosts_file, "\r\n\r\n\r\n")
	HijackGithubHosts(hosts_file, os.Args[1], "github.com")
}
