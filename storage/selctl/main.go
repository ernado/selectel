package main

import (
	"errors"
	"fmt"
	"github.com/ernado/selectel/storage"
	"github.com/jwaldrip/odin/cli"
	"log"
	"os"
)

const (
	envKey       = storage.EnvKey
	envUser      = storage.EnvUser
	envContainer = "SELECTEL_CONTAINER"
)

var (
	client    = cli.New("0.1", "Selectel storage command line client", connect)
	user, key string
	container string
	api       storage.API
	debug     bool
)

func init() {
	client.DefineBoolFlagVar(&debug, "debug", false, "debug mode")
	client.DefineStringFlag("key", "", fmt.Sprintf("selectel storage key (%s)", envKey))
	client.DefineStringFlag("user", "", fmt.Sprintf("selectel storage user (%s)", envUser))
	client.DefineStringFlag("container", "", fmt.Sprintf("default container (%s)", envContainer))
	infoCommand := client.DefineSubCommand("info", "print information about storage/container/object", info)
	infoCommand.DefineStringFlag("type", "storage", "storage, container or object")
}

func readFlag(c cli.Command, name, env string) string {
	if len(os.Getenv(env)) > 0 {
		return os.Getenv(env)
	}
	return c.Flag(name).String()
}

func blank(s string) bool {
	return len(s) == 0
}

// connect reads credentials and performs auth
func connect(c cli.Command) {
	key = readFlag(c, "key", envKey)
	user = readFlag(c, "user", envUser)
	container = readFlag(c, "container", envContainer)

	// checking for blank credentials
	if blank(key) || blank(user) {
		log.Fatal(storage.ErrorBadCredentials)
	}

	// connencting to api
	var err error
	api, err = storage.New(user, key)
	if err != nil {
		log.Fatal(err)
	}
	api.Debug(debug)
}

// info prints information about storage
func info(c cli.Command) {
	connect(c.Parent())
	var (
		containerName = container
		objectName    string
		data          interface{}
		err           error
		arglen        = len(c.Args())
		command       = c.Flag("type").String()
	)

	defer func() {
		if err != nil {
			log.Fatal(err)
		}
		if blank(containerName) || command == "storage" {
			data = api.Info()
		} else {
			containerApi := api.Container(containerName)
			if blank(objectName) {
				data, err = containerApi.Info()
			} else {
				data, err = containerApi.Object(objectName).Info()
			}
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", data)
	}()

	if arglen > 0 {
		if command == "container" {
			containerName = c.Arg(0).String()
			return
		}
		command = "object"
		if !blank(containerName) && arglen == 1 {
			objectName = c.Arg(0).String()
			return
		}
		if arglen == 2 {
			containerName = c.Arg(0).String()
			objectName = c.Arg(1).String()
			return
		}
	}
	if command == "container" && !blank(containerName) {
		return
	}
	if command == "storage" {
		return
	}
	err = errors.New("Not enough arguments")
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()
	client.Start()
}
