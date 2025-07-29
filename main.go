package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/shlex"
	"github.com/urfave/cli/v3"
)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
var appDir = "/comfyui"
var dataDir = "/data"
var userName = "comfyui"
var defaultUid = "1000"
var defaultGid = "1000"

func runCmd(command ...string) error {
	logger.Info("running command", "command", strings.Join(command, " "))
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

type SetupOpts struct {
	ComfyUIVersion     string
	TorchIndexUrl      string
	TorchVersion       string
	TorchaudioVersion  string
	TorchvisionVersion string
}

func setup(opts SetupOpts) error {
	torch := fmt.Sprintf("torch==%s", opts.TorchVersion)
	torchaudio := fmt.Sprintf("torchaudio==%s", opts.TorchaudioVersion)
	torchvision := fmt.Sprintf("torchvision==%s", opts.TorchvisionVersion)
	comfy := fmt.Sprintf("https://github.com/comfyanonymous/ComfyUI/archive/refs/tags/v%s.tar.gz", opts.ComfyUIVersion)
	archive := "/tmp/archive.tar.gz"

	comfyId := fmt.Sprintf("%s:%s", userName, userName)
	commands := [][]string{
		// update apt cache
		{"apt", "-y", "update"},
		// install utilities
		{"apt", "-y", "install", "gosu"},
		// create comfyui user and group
		{"groupadd", "-g", defaultGid, userName},
		{"useradd", "-u", defaultUid, "-g", defaultGid, userName},
		// install torch
		{"pip", "install", "--no-cache-dir", torch, torchaudio, torchvision, "--index-url", opts.TorchIndexUrl},
		// download and extract comfyui
		{"mkdir", "-p", appDir, dataDir},
		{"curl", "-o", archive, "-fsSL", comfy},
		{"tar", "xvzf", archive, "-C", appDir, "--strip-components", "1"},
		{"rm", "-f", archive},
		// install comfyui dependencies
		{"pip", "install", "--no-cache-dir", "-r", filepath.Join(appDir, "requirements.txt")},
		// ensure directories are owned by non-root user
		{"chown", "-R", comfyId, appDir, dataDir},
	}
	for _, command := range commands {
		err := runCmd(command...)
		if err != nil {
			return err
		}
	}

	return nil
}

type InstallNodesOpts struct {
	Nodes []string
}

func installNodes(opts InstallNodesOpts) error {
	nodesDir := filepath.Join(appDir, "custom_nodes")
	for _, node := range opts.Nodes {
		logger.Info("installing node", "node", node)
		nodeDir := filepath.Join(nodesDir, node)
		requirements := filepath.Join(nodeDir, "requirements.txt")
		commands := [][]string{
			// clone node
			{"git", "clone", node, nodeDir},
			// install node dependencies
			{"pip", "install", "-r", requirements},
		}
		for _, command := range commands {
			err := runCmd(command...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type EntrypointOpts struct {
	Arguments []string
	Gid       int
	Uid       int
}

func entrypoint(opts EntrypointOpts) error {
	launchCommand := append([]string{"python", filepath.Join(appDir, "main.py")}, opts.Arguments...)

	comfyUser, err := user.Lookup(userName)
	if err != nil {
		return err
	}
	comfyUid, err := strconv.Atoi(comfyUser.Uid)
	if err != nil {
		return err
	}
	comfyGid, err := strconv.Atoi(comfyUser.Gid)
	if err != nil {
		return err
	}

	current := []int{os.Getuid(), os.Getgid()}
	comfy := []int{comfyUid, comfyGid}
	desired := []int{opts.Uid, opts.Gid}

	if current[0] != 0 {
		// as a non-root user, the entrypoint cannot modify permissions nor do privilege de-escalation
		// simply run the launch command
		if !(reflect.DeepEqual(current, comfy) && reflect.DeepEqual(comfy, desired)) {
			return fmt.Errorf("cannot modify permissions as a non-root user")
		}
		return runCmd(launchCommand...)
	}

	commands := [][]string{}
	if comfy[0] != desired[0] {
		// update user uid
		command := []string{"usermod", "-u", strconv.Itoa(desired[0]), userName}
		commands = append(commands, command)
	}
	if comfy[1] != desired[1] {
		// update user gid
		command := []string{"groupmod", "-g", strconv.Itoa(desired[1]), userName}
		commands = append(commands, command)
	}

	commands = append(commands, [][]string{
		// ensure data dir exists with correct ownership
		{"mkdir", "-p", dataDir}}...,
	)
	dataDirs := []string{
		"checkpoints",
		"clip_vision",
		"controlnet",
		"diffusion_models",
		"embeddings",
		"clip",
		"diffusers",
		"gligen",
		"hypernetworks",
		"loras",
		"photomaker",
		"style_models",
		"text_encoders",
		"unet",
		"upscale_models",
		"vae",
		"vae_approx",
	}
	for _, dir := range dataDirs {
		srcPath := filepath.Join(appDir, "models", dir)
		dstPath := filepath.Join(dataDir, dir)
		commands = append(commands, [][]string{
			// delete comfyui path
			{"rm", "-rf", srcPath},
			// create data path
			{"mkdir", "-p", dstPath},
			// symlink comfyui path to data path
			{"ln", "-s", dstPath, srcPath},
		}...)
	}

	// modify the launch command to perform privilege de-escalation
	comfyId := fmt.Sprintf("%s:%s", userName, userName)
	launchCommand = append([]string{"gosu", comfyId}, launchCommand...)
	commands = append(commands, [][]string{
		{"chown", "-R", comfyId, dataDir, appDir},
		// launch the app
		launchCommand,
	}...)

	for _, command := range commands {
		err := runCmd(command...)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := (&cli.Command{
		Commands: []*cli.Command{
			{
				Name: "setup",
				Action: func(ctx context.Context, c *cli.Command) error {
					return setup(SetupOpts{
						ComfyUIVersion:     c.String("comfyui-version"),
						TorchIndexUrl:      c.String("torch-index-url"),
						TorchVersion:       c.String("torch-version"),
						TorchaudioVersion:  c.String("torchaudio-version"),
						TorchvisionVersion: c.String("torchvision-version"),
					})
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "comfyui-version",
						Required: true,
						Sources:  cli.EnvVars("COMFYUI_VERSION"),
					},
					&cli.StringFlag{
						Name:     "torch-index-url",
						Required: true,
						Sources:  cli.EnvVars("TORCH_INDEX_URL"),
					},
					&cli.StringFlag{
						Name:     "torch-version",
						Required: true,
						Sources:  cli.EnvVars("TORCH_VERSION"),
					},
					&cli.StringFlag{
						Name:     "torchaudio-version",
						Required: true,
						Sources:  cli.EnvVars("TORCHAUDIO_VERSION"),
					},
					&cli.StringFlag{
						Name:     "torchvision-version",
						Required: true,
						Sources:  cli.EnvVars("TORCHVISION_VERSION"),
					},
				},
			},
			{
				Name: "install-nodes",
				Action: func(ctx context.Context, c *cli.Command) error {
					nodes := c.StringArgs("nodes")
					var err error
					if len(nodes) == 0 {
						nodes, err = shlex.Split(c.String("nodes"))
						if err != nil {
							return err
						}
					}
					return installNodes(InstallNodesOpts{
						Nodes: nodes,
					})
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "nodes",
						Sources: cli.EnvVars("NODES"),
					},
				},
			},
			{
				Name: "entrypoint",
				Action: func(ctx context.Context, c *cli.Command) error {
					var err error
					arguments := c.StringArgs("arguments")
					if len(arguments) == 0 {
						arguments, err = shlex.Split(c.String("arguments"))
						if err != nil {
							return err
						}
					}
					return entrypoint(EntrypointOpts{
						Arguments: arguments,
						Gid:       c.Int("gid"),
						Uid:       c.Int("uid"),
					})
				},
				Arguments: []cli.Argument{
					&cli.StringArgs{
						Name: "arguments",
						Min:  0,
						Max:  -1,
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "arguments",
						Sources: cli.EnvVars("ARGUMENTS"),
					},
					&cli.IntFlag{
						Name:     "gid",
						Required: true,
						Sources:  cli.EnvVars("GID"),
					},
					&cli.IntFlag{
						Name:     "uid",
						Required: true,
						Sources:  cli.EnvVars("UID"),
					},
				},
			},
		},
	}).Run(context.Background(), os.Args)

	code := 0
	if err != nil {
		logger.Error("command failed", "error", err.Error())
		code = 1
	}

	os.Exit(code)
}
