package util

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"

	//"github.com/hamzaanjum9696/server-patching/cmd/server-patcher"
	"golang.org/x/crypto/ssh"
)

type ServerType int

const (
	WebApp  ServerType = 0
	Backend ServerType = 1
	Apache  ServerType = 2
	Unknown ServerType = 3
)

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
