package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hamzaanjum9696/server-patching/internal/util"
	"gopkg.in/yaml.v2"
)

const RequiredArgs int = 2

type AutomationConfiguration struct {
	PatchNotifyEnabled bool               `yaml:"patch_notifications_enabled"`
	PatchNotifyFrom    string             `yaml:"patch_notify_from_email"`
	PatchNotifyTo      string             `yaml:"patch_notify_to_emails"`
	Applications       []util.Application `yaml:"applications"`
}

type CLIOptions struct {
	Command    string
	IP         string
	Port       int
	Username   string
	Password   string
	ServerType util.ServerType
}

func usage(exitCode int) {
	log.Println("Usage: ./server-patcher start/stop")
	os.Exit(exitCode)
}

func printProcessContext(pc util.ProcessContext) {
	log.Printf("Process Context:\n")
	log.Printf("  PID: %d\n", pc.PID)
	log.Printf("  User is: %s\n", pc.ProcessOwner)
	log.Printf("  Process Name: %s\n", pc.ProcessName)
	log.Printf("  Process Path: %s\n", pc.ProcessPath)
	log.Printf("  Launch Path: %s\n", pc.LaunchPath)
}

func loadConfig(filePath string) (*AutomationConfiguration, error) {
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

func main() {

	log.SetOutput(os.Stdout)

	if len(os.Args) != RequiredArgs {
		usage(1)
	}

	if os.Args[1] == "help" {
		usage(0)
	}

	configFilePath := "config.yaml"
	config, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatal("Error loading config file:", err)
	}

	options := CLIOptions{
		Command: os.Args[1],
	}

	switch options.Command {
	case "start":
		log.Fatal("Not Implemented Yet")
	case "stop":

		AllProcessContexts := make([]util.ProcessContext, 0)
		for _, app := range config.Applications {
			fmt.Println("Application:", app.Name)
			processContexts := util.BuildProcessContexts(app.ProcessFilter, app.LaunchPathCommand)
			for _, pc := range processContexts {
				//printProcessContext(pc)
				AllProcessContexts = append(AllProcessContexts, pc)
			}
		}

		fmt.Println("Do you want to stop these processes? (yes/no)")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		userInput := strings.TrimSpace(scanner.Text())

		if userInput == "yes" {
			//executeStopSteps()
			log.Println("Stopping Services Now")
			// save snapshot now
			err = util.SaveProcessContexts(AllProcessContexts)
			if err != nil {
				log.Fatalf("Error in Saving snapshot file: %s\n", err)
			}
			// stop services now
			for _, step := range config.Applications {
				fmt.Println("Executing Stop step:", step)
				// Implement your logic for each step here
			}
		} else {
			log.Println("Stopping Execution Now!!!")
		}
	default:
		usage(1)
		log.Println("Wrong Input!!!")
	}
	os.Exit(0)

}
