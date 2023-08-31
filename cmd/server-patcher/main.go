package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hamzaanjum9696/server-patching/internal/util"
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
	fmt.Printf("Process Context:\n")
	fmt.Printf("  PID: %d\n", pc.PID)
	fmt.Printf("  User is: %s\n", pc.ProcessOwner)
	fmt.Printf("  Process Name: %s\n", pc.ProcessName)
	fmt.Printf("  Process Path: %s\n", pc.ProcessPath)
	fmt.Printf("  Launch Path: %s\n", pc.LaunchPath)
}

func PromtUserForKillingConfirmation(allpc []util.ProcessContext) string {

	fmt.Println("Processes Running...")
	for _, pc := range allpc {
		printProcessContext(pc)
	}

	fmt.Println("Do you want to stop these processes? (yes/no)")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := strings.TrimSpace(scanner.Text())

	if userInput == "yes" {
		return "yes"
	} else {
		return "no"
	}
}

func StopServicesNow(AllProcessContexts []util.ProcessContext, config *util.AutomationConfiguration) {
	// save snapshot now
	err := util.SaveProcessContexts(AllProcessContexts)
	if err != nil {
		log.Fatalf("Error in Saving snapshot file: %s\n", err)
	}
	log.Println("Stopping Services Now")
	// stop services now
	for _, step := range config.Applications {
		fmt.Println("Executing Stop step:", step)
		// Implement your logic for each step here
	}
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
	config, err := util.LoadConfig(configFilePath)
	if err != nil {
		log.Fatal("Error loading configurations:", err)
	}

	options := CLIOptions{
		Command: os.Args[1],
	}

	switch options.Command {
	case "start":
		log.Fatal("Not Implemented Yet")
	case "stop":

		AllProcessContexts := util.BuildProcessContexts(config)
		var userConfirmation string
		if len(AllProcessContexts) > 0 {
			userConfirmation = PromtUserForKillingConfirmation(AllProcessContexts)
		} else {
			log.Println("No Process To Delete")
			return
		}

		if userConfirmation == "yes" {
			StopServicesNow(AllProcessContexts, config)
		} else {
			log.Println("Terminating Process as per user request!!!")
		}
	default:
		usage(1)
		log.Println("Wrong Input!!!")
	}
	os.Exit(0)

}
