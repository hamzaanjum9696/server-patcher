package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	//"github.com/hamzaanjum9696/server-patching/internal/prepare"
	//"github.com/hamzaanjum9696/server-patching/internal/start"
	"github.com/hamzaanjum9696/server-patching/internal/util"
	"gopkg.in/yaml.v2"
)

const RequiredArgs int = 6

type Options struct {
	Command    string
	IP         string
	Port       int
	Username   string
	Password   string
	ServerType util.ServerType
}

type CommandStep struct {
	Command        string   `yaml:"command"`
	Args           []string `yaml:"args"`
	CaptureOutput  bool     `yaml:"capture_output"`
	OutputVariable string   `yaml:"output_variable"`
}

type Application struct {
	Name        string        `yaml:"name"`
	IPRegex     string        `yaml:"ip_regex"`
	StartSteps  []CommandStep `yaml:"start_steps"`
	StopSteps   []CommandStep `yaml:"stop_steps"`
	HealthCheck struct {
		Command    string `yaml:"command"`
		NumRetries int    `yaml:"num_retries"`
		Timeout    string `yaml:"timeout"`
	} `yaml:"health_check"`
}

type Configuration struct {
	PatchNotifyEnabled bool              `yaml:"patch_notifications_enabled"`
	IpMappings         map[string]string `yaml:"application_ip_mappings"`
	PatchNotifyFrom    string            `yaml:"patch_notify_from_email"`
	PatchNotifyTo      []string          `yaml:"patch_notify_to_emails"`
	PatchNotifyCC      []string          `yaml:"patch_notify_cc_emails"`
	Applications       []Application     `yaml:"applications"`
}

func usage(exitCode int) {
	log.Println("Usage: ./server-patcher <command> <ip> <port> <username> <pass>")
	os.Exit(exitCode)
}

func main() {

	// deserialize into Configuration struct from a local config.yaml file
	configFile, err := os.Open("../../config.yaml")
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer configFile.Close()

	var config Configuration
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error decoding config file:", err)
	}

	log.Printf("Running with configuration: %#v\n", config)

	if len(os.Args) < RequiredArgs {
		usage(1)
	}

	if os.Args[1] == "help" {
		usage(0)
	}

	if !util.IsValidIP(os.Args[2]) {
		log.Fatal("Provided Argument for IP Address is not Valid")
	}

	port, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal("Unable to parse arg for <port> as int")
	}

	options := Options{
		Command:  os.Args[1],
		IP:       os.Args[2],
		Port:     port,
		Username: os.Args[4],
		Password: os.Args[5],
	}

	options.ServerType = util.DetermineServerType(options.IP, config.IpMappings)

	switch options.Command {
	case "start":
		for _, app := range config.Applications {

			for _, step := range app.StartSteps {
				step.Command = strings.ReplaceAll(step.Command, "{{.IP}}", options.IP)
				if step.CaptureOutput {
					output, err := util.RunRemoteCommand(options.IP, options.Port, options.Username, options.Password, step.Command)
					if err != nil {
						log.Fatalf("Error executing command: %s\n", err)
					}
					step.OutputVariable = string(output)
				} else {
					_, err := util.RunRemoteCommand(options.IP, options.Port, options.Username, options.Password, step.Command)
					if err != nil {
						log.Fatalf("Error executing command: %s\n", err)
					}
				}
			}
		}
	case "stop":
		for _, app := range config.Applications {
			fmt.Printf("Running stop steps for %s...\n", app.Name)

			// Execute stop steps
			for _, step := range app.StopSteps {
				step.Command = strings.ReplaceAll(step.Command, "{{.IP}}", options.IP)
				if step.CaptureOutput {
					output, err := util.RunRemoteCommand(options.IP, options.Port, options.Username, options.Password, step.Command)
					if err != nil {
						log.Fatalf("Error executing command: %s\n", err)
					}
					step.OutputVariable = string(output)
				} else {
					fmt.Printf("No capture Command: %s\n", step.Command)
					// Execute the command without capturing the output
					_, err := util.RunRemoteCommand(options.IP, options.Port, options.Username, options.Password, step.Command)
					if err != nil {
						log.Fatalf("Error executing command: %s\n", err)
					}
				}
			}
		}

		// write function here to send snapshot email
		// subject := "Snapshot: [ " + ip + " ]"
		// err = util.SendSnapshotEmail(config.PatchNotifyFrom, config.PatchNotifyTo, subject, apacheProcessNames)
		// if err != nil {
		// 	log.Fatalf("Error:", err)
		// 	return
		// } else {
		// 	log.Println("Email sent successfully!")
		// }

	}
}
