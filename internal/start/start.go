package start

import (
	"fmt"
	"log"
	"strings"

	"github.com/hamzaanjum9696/server-patching/internal/prepare"
	"github.com/hamzaanjum9696/server-patching/internal/util"
)

func getApacheName(path string) string {
	parts := strings.Split(path, "/")
	apacheName := parts[3]
	return apacheName
}

func ReadSnapshotFromDir(serverAddr string, port int, username string, password string) (string, error) {

	// Create file path
	remoteFilePath := "/u/Server_Patching_Automation/" + serverAddr + "-Snapshot.txt"

	command := fmt.Sprintf("cat %s", remoteFilePath)
	apacheProcessNamesToReturn, err := util.RunRemoteCommand(serverAddr, port, username, password, command)
	if err != nil {
		log.Fatal("Error reading Snapshot file:", err)
		return "", err
	}
	return string(apacheProcessNamesToReturn), nil
}

func StartApacheProcesses(serverAddr string, port int, username string, password string, apachesNames string) bool {
	var allApacheStartedStatus = true
	lines := strings.Split(apachesNames, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine != "" {
			currApacheName := getApacheName(trimmedLine)
			dirPath := strings.TrimSuffix(trimmedLine, "/httpd")
			startCommand := dirPath + "/apachectl start"
			log.Printf("Starting Apache: %s\n", currApacheName)
			_, err := util.RunRemoteCommand(serverAddr, port, username, password, startCommand)
			if err != nil {
				log.Printf("Error Executing Command: %s\n", startCommand)
				log.Fatal("Error: \n", err)
				allApacheStartedStatus = false
			}
		}
	}
	return allApacheStartedStatus
}

func ValidateSnapshot(serverAddr string, port int, username string, password string, apachesNamesSnapshot string) bool {

	var allApacheStartedStatus = true
	liveServerProcesses, err := prepare.GetApacheProcessNames(serverAddr, port, username, password)
	//lengthLiveServerProcesses := len(liveServerProcesses)

	if err != nil {
		// Error getting updated Values from Server so validation failed
		allApacheStartedStatus = false
	} else {
		lines := strings.Split(apachesNamesSnapshot, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			if trimmedLine != "" {
				currApacheName := getApacheName(trimmedLine)
				log.Printf("Validating Apache: %s\n", currApacheName)
				if !util.StringInArray(trimmedLine, liveServerProcesses) {
					log.Printf("No Process Present for Apache: %s\n", currApacheName)
					allApacheStartedStatus = false
				}
			}
		}
	}

	return allApacheStartedStatus
}
