package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cri-o/cri-o/internal/client"
	"github.com/cri-o/cri-o/internal/pkg/completion"
	"github.com/cri-o/cri-o/internal/version"
	"github.com/urfave/cli"
)

const (
	defaultSocket = "/var/run/crio/crio.sock"
	idArg         = "id"
	socketArg     = "socket"
)

func main() {
	app := cli.NewApp()
	app.Name = "crio-status"
	app.Usage = "A tool for CRI-O status retrieval"
	app.Version = version.Version
	app.CommandNotFound = func(*cli.Context, string) { os.Exit(1) }
	app.OnUsageError = func(c *cli.Context, e error, b bool) error { return e }
	app.Action = func(c *cli.Context) error {
		return fmt.Errorf("expecting a valid subcommand")
	}

	flags := []cli.Flag{
		cli.StringFlag{
			Name: socketArg + ", s",
			Usage: fmt.Sprintf(
				"absolute path to the unix socket (default: %q)",
				defaultSocket,
			),
			TakesFile: true,
		},
	}
	app.Flags = flags
	app.Commands = []cli.Command{{
		Action:  config,
		Aliases: []string{"c"},
		Flags:   flags,
		Name:    "config",
		Usage:   "retrieve the configuration as a TOML string",
	}, {
		Action:  info,
		Aliases: []string{"i"},
		Flags:   flags,
		Name:    "info",
		Usage:   "retrieve generic information",
	}, {
		Action:  containers,
		Aliases: []string{"container", "cs", "s"},
		Flags: append(flags, cli.StringFlag{
			Name:  idArg + ", i",
			Usage: "the container ID",
		}),
		Name:  "containers",
		Usage: "retrieve information about containers",
	},
		completion.Command,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func config(c *cli.Context) error {
	crioClient, err := crioClient(c)
	if err != nil {
		return err
	}

	info, err := crioClient.ConfigInfo()
	if err != nil {
		return err
	}

	fmt.Print(info)
	return nil
}

func containers(c *cli.Context) error {
	crioClient, err := crioClient(c)
	if err != nil {
		return err
	}

	id := c.String(idArg)
	if id == "" {
		return fmt.Errorf("the argument --%s cannot be empty", idArg)
	}

	info, err := crioClient.ContainerInfo(c.String(idArg))
	if err != nil {
		return err
	}

	fmt.Printf("name: %s\n", info.Name)
	fmt.Printf("pid: %d\n", info.Pid)
	fmt.Printf("image: %s\n", info.Image)
	fmt.Printf("image ref: %s\n", info.ImageRef)
	fmt.Printf("created: %v\n", info.CreatedTime)
	fmt.Printf("labels:\n")
	for k, v := range info.Labels {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Printf("annotations:\n")
	for k, v := range info.Annotations {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Printf("CRI-O annotations:\n")
	for k, v := range info.CrioAnnotations {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Printf("log path: %s\n", info.LogPath)
	fmt.Printf("root: %s\n", info.Root)
	fmt.Printf("sandbox: %s\n", info.Sandbox)
	fmt.Printf("ips: %s\n", strings.Join(info.IPs, ", "))

	return nil
}

func info(c *cli.Context) error {
	crioClient, err := crioClient(c)
	if err != nil {
		return err
	}

	info, err := crioClient.DaemonInfo()
	if err != nil {
		return err
	}

	fmt.Printf("cgroup driver: %s\n", info.CgroupDriver)
	fmt.Printf("storage driver: %s\n", info.StorageDriver)
	fmt.Printf("storage root: %s\n", info.StorageRoot)

	fmt.Printf("default GID mappings (format <container>:<host>:<size>):\n")
	for _, m := range info.DefaultIDMappings.Gids {
		fmt.Printf("  %d:%d:%d\n", m.ContainerID, m.HostID, m.Size)
	}
	fmt.Printf("default UID mappings (format <container>:<host>:<size>):\n")
	for _, m := range info.DefaultIDMappings.Uids {
		fmt.Printf("  %d:%d:%d\n", m.ContainerID, m.HostID, m.Size)
	}

	return nil
}

func crioClient(c *cli.Context) (client.CrioClient, error) {
	socket := defaultSocket
	if c.GlobalString(socketArg) != "" {
		socket = c.GlobalString(socketArg)
	} else if c.String(socketArg) != "" {
		socket = c.String(socketArg)
	}
	return client.New(socket)
}
