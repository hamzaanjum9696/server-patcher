package util

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/crypto/ssh"
)

type ServerType int

type ProcessContext struct {
	PID         int
	ProcessName string
	ProcessPath string
	LaunchPath  string
}

type CommandStep struct {
	Command string `yaml:"command"`
}

type Application struct {
	Name          string        `yaml:"name"`
	ProcessFilter string        `yaml:"process_filter"`
	IPRegex       string        `yaml:"ip_regex"`
	StartSteps    []CommandStep `yaml:"start_steps"`
	StopSteps     []CommandStep `yaml:"stop_steps"`

	// this one doesn't have a yaml key because its not from yaml,
	// we are building it ourselves using process_filter from yaml
	// still need to solve the uniqueness warning problem
	// for now, we can just add a `are you sure? (y/n)` check and list
	// all the processes. Then, user can read and confirm that they
	// are starting/stopping the processes that they wanted to.
	ProcessContext ProcessContext

	HealthCheck struct {
		Command    string `yaml:"command"`
		NumRetries int    `yaml:"num_retries"`
		Timeout    string `yaml:"timeout"`
	} `yaml:"health_check"`
}

const (
	WebApp  ServerType = 0
	Backend ServerType = 1
	Apache  ServerType = 2
	Unknown ServerType = 3
)

func BuildProcessContext(processFilter string) *ProcessContext {
	// using the process filter, extract process name, process id, launch path, process path
	// and return a ProcessContext struct
	// example: processFilter = "httpd"
	// output:
	// ProcessContext{
	// 	  PID: 1234,
	// 	  ProcessName: "httpd",
	// 	  ProcessPath: "/usr/sbin/httpd",
	// 	  LaunchPath: "/u/apps/apache/bin"
	// }

	matches := findProcesses(processFilter)
	log.Printf("Processes: %s\n", matches)
	return nil

}

func findProcesses(processFilter string) []*process.Process {
	processes, err := process.Pids()
	matches := make([]*process.Process, 0)

	if err != nil {
		log.Fatalf("Error getting process ids: %s\n", err)
	}

	log.Printf("Found %d processes\n", len(processes))

	for _, pid := range processes {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}
		if pname, err := p.Name(); err == nil && strings.Contains(pname, processFilter) {
			log.Printf("Found matching process %s with PID: %d\n", pname, pid)
			matches = append(matches, p)
		}
	}

	return matches
}

func SendEmail(notifyFromEmail string, notifyToEmails []string, ipAddress string, subject string, body []string) error {
	bodyText := strings.Join(body, "\n")

	cmd := exec.Command(
		"mail",
		"-s", subject,
		"-r", notifyFromEmail,
		strings.Join(notifyToEmails, ","),
	)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("IP Address: %s\n\n%s", ipAddress, bodyText))
	// in case extra dot (.) is coming in email at the end of apache, try below command
	// cmd.Stdin = strings.NewReader(fmt.Sprintf("IP Address: %s\n\n%s", ipAddress, bodyText))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error sending email: %s\nOutput: %s", err, output)
	}

	return nil
}

func SendSnapshotEmail(notifyFromEmail string, notifyToEmails []string, notifyCCEmails []string, ipAddress string, subject string, body []string) error {
	bodyText := strings.Join(body, "\n")

	// in case extra dot (.) is coming in email at the end of apache, try below command
	// cmd.Stdin = strings.NewReader(fmt.Sprintf("IP Address: %s\n\n%s", ipAddress, bodyText))

	cmd := exec.Command(
		"mail",
		"-s", subject,
		"-c", strings.Join(notifyCCEmails, ","),
		"-r", notifyFromEmail,
		strings.Join(notifyToEmails, ","),
	)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("IP Address: %s\n\n%s", ipAddress, bodyText))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error sending email: %s\nOutput: %s", err, output)
	}

	return nil
}

func RunRemoteCommand(serverAddr string, port int, username string, password string, command string) (string, error) {

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{

			ssh.Password(password),
		},
		Config: ssh.Config{
			KeyExchanges: []string{"diffie-hellman-group-exchange-sha256", "diffie-hellman-group14-sha256", "diffie-hellman-group14-sha1"},
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// Always return nil to ignore host key verification
			return nil
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", serverAddr, port), config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(command)
	if err != nil {
		log.Printf("Error in Executing Command: %s\n", command)
		log.Printf("Error as follows:\n%s", err)
		return "", err
	}

	return stdoutBuf.String(), nil
}

func StringInArray(target string, array []string) bool {
	for _, item := range array {
		item = strings.TrimSpace(item)
		target = strings.TrimSpace(target)
		if item == target {
			return true
		}
	}
	return false
}

func IsValidIP(provided_ip string) bool {
	parsedIP := net.ParseIP(provided_ip)
	return parsedIP != nil
}

func matchIPPattern(ip string, pattern string) bool {
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(ip)
}

func DetermineServerType(ip_address string, ipMappings map[string]string) ServerType {

	// define all possible patterns here
	ipPatternWEBServers := ipMappings["WebApp"]
	ipPatternBEServers := ipMappings["Backend"]
	ipPatternApacheServers := ipMappings["Apache"]

	if matchIPPattern(ip_address, ipPatternWEBServers) {
		return WebApp
	} else if matchIPPattern(ip_address, ipPatternBEServers) {
		return Backend
	} else if matchIPPattern(ip_address, ipPatternApacheServers) {
		return Apache
	} else {
		return Unknown
	}
}
