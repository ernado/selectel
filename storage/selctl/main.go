package main

import (
	"errors"
	"fmt"
	"github.com/ernado/selectel/storage"
	"github.com/jwaldrip/odin/cli"
	"github.com/olekukonko/tablewriter"
	"io"
	"log"
	"os"
)

const (
	envKey       = storage.EnvKey
	envUser      = storage.EnvUser
	envContainer = "SELECTEL_CONTAINER"
)

var (
	client         = cli.New("0.1", "Selectel storage command line client", connect)
	user, key      string
	container      string
	api            storage.API
	debug          bool
	errorNotEnough = errors.New("Not enought arguments")
)

func init() {
	client.DefineBoolFlagVar(&debug, "debug", false, "debug mode")
	client.DefineStringFlag("key", "", fmt.Sprintf("selectel storage key (%s)", envKey))
	client.DefineStringFlag("user", "", fmt.Sprintf("selectel storage user (%s)", envUser))
	client.DefineStringFlag("container", "", fmt.Sprintf("default container (%s)", envContainer))
	infoCommand := client.DefineSubCommand("info", "print information about storage/container/object", wrap(info))
	infoCommand.DefineStringFlag("type", "storage", "storage, container or object")
	client.DefineSubCommand("upload", "upload object to container", wrap(upload))
	listCommand := client.DefineSubCommand("list", "list objects in container/storage", wrap(list))
	listCommand.DefineStringFlag("type", "storage", "storage or container")
	downloadCommand := client.DefineSubCommand("download", "download object from container", wrap(download))
	downloadCommand.DefineStringFlag("path", "", "destination path")
	removeCommand := client.DefineSubCommand("remove", "remove object or container", wrap(remove))
	removeCommand.DefineStringFlag("type", "object", "container or object")
	client.DefineSubCommand("create", "create container", wrap(create))
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

func wrap(callback func(cli.Command)) func(cli.Command) {
	return func(c cli.Command) {
		connect(c.Parent())
		callback(c)
	}
}

// info prints information about storage
func info(c cli.Command) {
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
	err = errorNotEnough
}

func remove(c cli.Command) {
	var (
		arglen  = len(c.Args())
		object  string
		err     error
		message string
	)
	if arglen == 2 {
		container = c.Arg(0).String()
		object = c.Arg(1).String()
	}
	if arglen == 1 {
		if c.Flag("type").String() == "container" {
			container = c.Arg(0).String()
		} else {
			object = c.Arg(0).String()
		}
	}
	if blank(container) {
		log.Fatal(errorNotEnough)
	}
	if blank(object) {
		err = api.Container(container).Remove()
		message = fmt.Sprintf("container %s removed", container)
	} else {
		err = api.Container(container).Object(object).Remove()
		message = fmt.Sprintf("object %s removed in container %s", object, container)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(message)
}

func create(c cli.Command) {
	if len(c.Args()) == 0 {
		log.Fatal(errorNotEnough)
	}
	var name = c.Arg(0).String()
	if _, err := api.CreateContainer(name, false); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created container %s\n", name)
}

func upload(c cli.Command) {
	var path string
	switch len(c.Args()) {
	case 1:
		path = c.Arg(0).String()
	case 2:
		container = c.Arg(0).String()
		path = c.Arg(1).String()
	}
	if blank(container) || blank(path) {
		log.Fatal(errorNotEnough)
	}
	if err := api.Container(container).UploadFile(path); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("uploaded to %s\n", container)
}

func list(c cli.Command) {
	var (
		arglen = len(c.Args())
		table  = tablewriter.NewWriter(os.Stdout)
	)
	if arglen == 0 && (blank(container) || c.Flag("type").String() == "storage") {
		containers, err := api.ContainersInfo()
		if err != nil {
			log.Fatal(err)
		}
		table.SetHeader([]string{"Name", "Objects", "Type"})
		for _, cont := range containers {
			v := []string{cont.Name, fmt.Sprint(cont.ObjectCount), cont.Type}
			table.Append(v)
		}
		table.Render()
		return
	}
	if arglen == 1 {
		container = c.Arg(0).String()
	}
	if blank(container) {
		log.Fatal(errorNotEnough)
	}
	objects, err := api.Container(container).ObjectsInfo()
	if err != nil {
		log.Fatal(err)
	}
	table.SetHeader([]string{"Name", "Size", "Downloaded"})
	for _, object := range objects {
		v := []string{object.Name, fmt.Sprint(object.Size), fmt.Sprint(object.Downloaded)}
		table.Append(v)
	}
	table.Render()
}

func download(c cli.Command) {
	var (
		arglen     = len(c.Args())
		objectName string
		path       = c.Flag("path").String()
	)
	switch arglen {
	case 1:
		objectName = c.Arg(0).String()
	case 2:
		objectName = c.Arg(1).String()
		container = c.Arg(0).String()
	}
	if blank(container) || blank(objectName) {
		log.Fatal(errorNotEnough)
	}
	if blank(path) {
		path = objectName
	}
	reader, err := api.Container(container).Object(objectName).GetReader()
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	fmt.Printf("downloading %s->%s from %s\n", objectName, path, container)
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	n, err := io.Copy(f, reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("downloaded %s, %d bytes\n", objectName, n)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()
	client.Start()
}
