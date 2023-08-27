package prepare

import (
	"fmt"
	"log"
	"strings"

	"github.com/hamzaanjum9696/server-patching/internal/util"
)

// normal parameters sequence
// /usr/local/apache2/bin/apachectl start
// /u/bin/tomcat-name-here/bin/startup.sh

// "/u/Server_Patching_Automation/snapshot-2021-07-21-12-00-00.json"

// SNAPSHOT FORMAT:
// {
// "shutdown_time": "2021-07-21 12:00:00",
// "processes": [
// 	{
// 		"process_name": "httpd",
// 		"process_path": "/usr/local/apache2/bin/httpd",
// 		"launch_path": "/usr/local/apache2/bin/",
// 		"process_owner": "root"
// 	},
// 	{
// 		"process_name": "tomcat",
// 		"process_path": "/u/bin/tomcat-name-here/bin/tomcat",
// 		"launch_path": "/u/bin/tomcat-name-here/bin/",
// 		"process_owner": "root"
// 	},
// 	{
// 		"process_name": "tomcat",
// 		"process_path": "/u/bin/tomcat-name-here/bin/tomcat",
// 		"launch_path": "/u/bin/tomcat-name-here/bin/",
// 		"process_owner": "app_user"
// 	}
// ]
// }
//

func SaveSnapshotInDir(serverAddr string, port int, username string, password string, processNames []string) error {

	// Create the remote directory
	command := "bash -c 'mkdir -p /u/Server_Patching_Automation/'"
	_, err := util.RunRemoteCommand(serverAddr, port, username, password, command)
	if err != nil {
		log.Fatal("Error creating directory:", err)
		return err
	}

	// Create the file
	directoryPath := "/u/Server_Patching_Automation/"
	filePath := directoryPath + serverAddr + "-Snapshot.txt"
	log.Printf("Snapshot Saveed in Path: %s\n", filePath)

	command = fmt.Sprintf("bash -c 'touch %s && chmod 644 %s'", filePath, filePath)
	_, err = util.RunRemoteCommand(serverAddr, port, username, password, command)
	if err != nil {
		log.Printf("Error creating file: %s\n", err)
		return err
	}

	escapedContent := strings.Join(processNames, "\n")
	heredoc := fmt.Sprintf("bash -c 'cat <<EOT > %s\n%s\nEOT'", filePath, escapedContent)
	_, err = util.RunRemoteCommand(serverAddr, port, username, password, heredoc)
	if err != nil {
		log.Printf("Error Writing in file: %s", err)
		return err
	}
	log.Println("Snapshot File Saved Successfully!!!")
	return nil
}

// Apache server snapshot function
func GetApacheProcessNames(serverAddr string, port int, username string, password string) ([]string, error) {

	// command to get apache processes
	commandToGetProcessNames := "ps -ef | grep [h]ttpd | grep bin | awk '{print $8}' | grep -v grep | sort | uniq"
	names := []string{}
	separator := "\n"

	output, err := util.RunRemoteCommand(serverAddr, port, username, password, commandToGetProcessNames)
	if err != nil {
		return names, err
	}

	lines := strings.Split(string(output), separator)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine != "" {
			names = append(names, trimmedLine)
		}
	}

	return names, nil
}

func StopApacheProcesses(serverAddr string, port int, username string, password string, apachePaths []string) bool {

	// preparing commands to stop apache processes
	allProcessesStopped := true

	stopCommands := make([]string, len(apachePaths))
	for i, path := range apachePaths {
		// Extract the directory path without the executable name
		dirPath := strings.TrimSuffix(path, "/httpd")

		// Append the "apachectl stop" command to stop the server
		stopCommands[i] = dirPath + "/apachectl stop"
	}

	// code to stop apache processess
	for _, command := range stopCommands {
		log.Printf("Executing Command: %s\n", command)

		_, err := util.RunRemoteCommand(serverAddr, port, username, password, command)
		if err != nil {
			log.Printf("Error Executing Command: %s\n", command)
			log.Fatal("Error: \n", err)
			allProcessesStopped = false
		}
	}

	// code to validate if apache processess have been stopped
	apacheProcessNamesValidation, err := GetApacheProcessNames(serverAddr, port, username, password)
	if err != nil {
		log.Fatal(err)
		allProcessesStopped = false
	}

	// logic to validate if processess have been stopped
	// if there is any apache still running that means our commands were
	// not able to stop apache. In that case, we return false status to main func
	apacheProcessNamesValidationLength := len(apacheProcessNamesValidation)
	if apacheProcessNamesValidationLength == 0 {
		// all processes stopped as there is no current process running
		allProcessesStopped = true
	} else {
		for _, currApachePath := range apacheProcessNamesValidation {
			parts := strings.Split(currApachePath, "/")
			currApacheName := parts[3]
			log.Printf("ALERT---Apache: %s process is still Running---ALERT\n", currApacheName)
		}
		allProcessesStopped = false
	}

	return allProcessesStopped
}
