package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type ServerType int

type AutomationConfiguration struct {
	PatchNotifyEnabled bool          `yaml:"patch_notifications_enabled"`
	PatchNotifyFrom    string        `yaml:"patch_notify_from_email"`
	PatchNotifyTo      string        `yaml:"patch_notify_to_emails"`
	Applications       []Application `yaml:"applications"`
}

type ProcessContext struct {
	PID          int32
	ProcessName  string
	ProcessPath  string
	LaunchPath   string
	ProcessOwner string
}

type CommandStep struct {
	Command string `yaml:"command"`
}

type Application struct {
	Name              string        `yaml:"name"`
	ProcessFilter     string        `yaml:"process_filter"`
	IPRegex           string        `yaml:"ip_regex"`
	LaunchPathCommand string        `yaml:"launch_path_command"`
	StartSteps        []CommandStep `yaml:"start_steps"`
	StopSteps         []CommandStep `yaml:"stop_steps"`

	HealthCheck struct {
		Command    string `yaml:"command"`
		NumRetries int    `yaml:"num_retries"`
		Timeout    string `yaml:"timeout"`
	} `yaml:"health_check"`
}

func getLaunchPath(processFilter string, pid int32, launchPathCommand string) string {
	var launchPath string
	if launchPathCommand != "" {
		tmpl, err := template.New("launch_path_command").Parse(launchPathCommand)
		if err != nil {
			log.Fatalf("Error parsing launch path command: %s\n", err)
		}

		var tpl bytes.Buffer
		err = tmpl.Execute(&tpl, struct {
			ProcessFilter string
			Pid           int32
		}{
			ProcessFilter: processFilter,
			Pid:           pid,
		})
		if err != nil {
			log.Fatalf("Error executing launch path command: %s\n", err)
		}

		launchPathCommand = tpl.String()

		var stdout, stderr bytes.Buffer

		cmd := exec.Command("bash", "-c", launchPathCommand)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			log.Fatalf("Error executing launch path command: %s\n", err)
		}
		if stderr.String() != "" {
			log.Fatalf("Error executing launch path command: %s\n", stderr.String())
		}
		launchPath = strings.TrimSpace(stdout.String())
	}

	return launchPath
}

func BuildProcessContexts(config *AutomationConfiguration) []ProcessContext {

	AllProcessContexts := make([]ProcessContext, 0)

	for _, app := range config.Applications {
		processes := findProcesses(app.ProcessFilter)
		for _, p := range processes {
			if ppid, err := p.Ppid(); err == nil && ppid != 1 {
				continue
			}
			processName, _ := p.Name()
			processPath, _ := p.Exe()
			var processLaunchPath string
			if app.LaunchPathCommand == "" {
				cmdLine, _ := p.CmdlineSlice()
				processLaunchPath = strings.Join(cmdLine, " ")
			} else {
				processLaunchPath = getLaunchPath(app.ProcessFilter, p.Pid, app.LaunchPathCommand)
			}
			processUser, _ := p.Username()

			// Create a new ProcessContext object and populate its values
			processContext := ProcessContext{
				PID:          p.Pid,
				ProcessName:  processName,
				ProcessPath:  processPath,
				LaunchPath:   processLaunchPath,
				ProcessOwner: processUser,
			}
			// Append the ProcessContext object to the array
			AllProcessContexts = append(AllProcessContexts, processContext)
		}
	}
	return AllProcessContexts
}

func LoadConfig(filePath string) (*AutomationConfiguration, error) {
	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config AutomationConfiguration
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveProcessContexts(processContexts []ProcessContext) error {

	directoryPath := "/u/Server-Patcher-Automation/"
	currentTime := time.Now()
	dateTimeString := currentTime.Format("02012006_150405")
	filename := dateTimeString + "-snapshot.json"
	filePath := filepath.Join(directoryPath, filename)

	if err := os.MkdirAll(directoryPath, os.ModePerm); err != nil {
		return err
	}

	// Open the file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	// Convert processContexts to JSON format
	jsonData, err := json.MarshalIndent(processContexts, "", "    ")
	if err != nil {
		return err
	}

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	log.Printf("Snapshot Saved at path: %s\n", filePath)
	return nil
}

func findProcesses(processFilter string) []*process.Process {
	processIDs, err := process.Pids()
	matches := make([]*process.Process, 0)

	if err != nil {
		log.Fatalf("Error getting process ids: %s\n", err)
	}

	for _, pid := range processIDs {
		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		pname, err := p.Name()
		if err != nil {
			continue
		}

		processPath, err := p.Exe()
		if err != nil {
			continue
		}

		processCWD, err := p.Cwd()
		if err != nil {
			continue
		}

		processCmdLine, err := p.CmdlineSlice()
		if err != nil {
			continue
		}

		if strings.Contains(pname, processFilter) ||
			strings.Contains(processPath, processFilter) ||
			strings.Contains(processCWD, processFilter) ||
			strings.Contains(strings.Join(processCmdLine, " "), processFilter) {
			matches = append(matches, p)
		}
	}

	//log.Printf("Found processes matching filter %s: %v\n", processFilter, matches)
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
		log.Fatalf("error sending email: %s; output: %s", err, output)
		return err
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
