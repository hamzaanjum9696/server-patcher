package main

import (
	"log"
	"os"
	"strconv"

	"github.com/hamzaanjum9696/server-patching/internal/prepare"
	"github.com/hamzaanjum9696/server-patching/internal/start"
	"github.com/hamzaanjum9696/server-patching/internal/util"
	"gopkg.in/yaml.v2"
)

const RequiredArgs int = 6

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
	log.Println("Usage: ./server-patcher <command> <ip> <port> <username> <pass>")
	os.Exit(exitCode)
}

func main() {

	log.SetOutput(os.Stdout)

	if len(os.Args) == 2 {
		util.BuildProcessContext(os.Args[1])
		os.Exit(0)
	}

	// deserialize into Configuration struct from a local config.yaml file
	configFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer configFile.Close()

	var config AutomationConfiguration
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

	options := CLIOptions{
		Command:  os.Args[1],
		IP:       os.Args[2],
		Port:     port,
		Username: os.Args[4],
		Password: os.Args[5],
	}

	switch options.Command {
	case "start":
		switch options.ServerType {
		case util.WebApp:
			panic("unimplemented")
		case util.Backend:
			panic("unimplemented")
		case util.Apache, util.Unknown:
			apacheProcessNames, err := start.ReadSnapshotFromDir(options.IP, options.Port, options.Username, options.Password)
			if err != nil {
				log.Fatal(err)
			}

			status := start.StartApacheProcesses(options.Password, options.Port, options.Username, options.Password, apacheProcessNames)
			log.Printf("Status is: %t\n", status)

			validationStatus := start.ValidateSnapshot(options.IP, options.Port, options.Username, options.Password, apacheProcessNames)
			if validationStatus {
				log.Println("All Server Present in Snapshot Started Successfully!!!")
			} else {
				log.Println("ALERT!!!All Server Present in Snapshot are not Started!!!ALERT")
			}
		}
	case "stop":
		switch options.ServerType {
		case util.WebApp:
			panic("unimplemented")
		case util.Backend:
			panic("unimplemented")
		case util.Apache, util.Unknown:
			apacheProcessNames, err := prepare.GetApacheProcessNames(options.IP, options.Port, options.Username, options.Password)
			if err != nil {
				log.Fatal(err)
			}

			numberOfApachesRunning := len(apacheProcessNames)
			log.Printf("Number of Apaches Running: %d\n", numberOfApachesRunning)
			for i, apache := range apacheProcessNames {
				log.Printf("Apache %d: %s\n", i+1, apache)
			}

			err = prepare.SaveSnapshotInDir(options.IP, options.Port, options.Username, options.Password, apacheProcessNames)
			if err != nil {
				log.Fatal(err)
			}

			if len(apacheProcessNames) > 0 {
				log.Printf("Stopping Apache Servers for Node: %s\n", options.IP)
				status := prepare.StopApacheProcesses(options.IP, options.Port, options.Username, options.Password, apacheProcessNames)
				if !status {
					log.Fatal("Apache Servers Could Not be Stopped Properly.")
				} else {
					log.Printf("Apache Servers Stopped on Node: %s\n", options.IP)
				}
			} else {
				log.Printf("Nothing to Stop on Node: %s\n", options.IP)
			}

		}

	}
}
