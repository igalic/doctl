/*
Copyright 2016 The Doctl Authors All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DoitCmd is the base command.
var DoitCmd = &Command{
	Command: &cobra.Command{
		Use:   "doctl",
		Short: "doctl is a command line interface for the DigitalOcean API.",
	},
}

// Token holds the global authorization token.
var Token string

// Output holds the global output format.
var Output string

// Verbose toggles verbose output.
var Verbose bool

var requiredColor = color.New(color.Bold, color.FgWhite).SprintfFunc()

// Writer is where output should be written to.
var Writer = os.Stdout

// Trace toggles http tracing output.
var Trace bool

func init() {
	viper.SetConfigType("yaml")

	DoitCmd.PersistentFlags().StringVarP(&Token, "access-token", "t", "", "DigitalOcean API V2 Access Token")
	DoitCmd.PersistentFlags().StringVarP(&Output, "output", "o", "text", "output formt [text|json]")
	DoitCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	DoitCmd.PersistentFlags().BoolVarP(&Trace, "trace", "", false, "verbose output")
}

// LoadConfig loads out configuration.
func LoadConfig() error {
	cf, err := doit.NewConfigFile()
	if err != nil {
		return err
	}

	r, err := cf.Open()
	if err != nil {
		return fmt.Errorf("can't open configuration file: %v", err)
	}

	return viper.ReadConfig(r)
}

// Init initializes the root command.
func Init() *Command {
	initializeConfig()
	addCommands()

	return DoitCmd
}

// AddCommands adds sub commands to the base command.
func addCommands() {
	DoitCmd.AddCommand(Account())
	DoitCmd.AddCommand(Auth())
	DoitCmd.AddCommand(computeCmd())
	DoitCmd.AddCommand(Version())
}

func computeCmd() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "compute",
			Short: "compute commands",
			Long:  "compute commands are for controlling and managing infrastructure",
		},
	}

	cmd.AddCommand(Actions())
	cmd.AddCommand(DropletAction())
	cmd.AddCommand(Droplet())
	cmd.AddCommand(Domain())
	cmd.AddCommand(FloatingIP())
	cmd.AddCommand(FloatingIPAction())
	cmd.AddCommand(Images())
	cmd.AddCommand(ImageAction())
	cmd.AddCommand(Plugin())
	cmd.AddCommand(Region())
	cmd.AddCommand(Size())
	cmd.AddCommand(SSHKeys())
	cmd.AddCommand(SSH())

	cmd.AddCommand(Drive())
	cmd.AddCommand(DriveAction())
	return cmd
}

func initFlags() {
	viper.SetEnvPrefix("DIGITALOCEAN")
	viper.BindEnv("access-token", "DIGITALOCEAN_ACCESS_TOKEN")
	viper.BindPFlag("access-token", DoitCmd.PersistentFlags().Lookup("access-token"))
	viper.BindPFlag("output", DoitCmd.PersistentFlags().Lookup("output"))
}

func loadDefaultSettings() {
	viper.SetDefault("output", "text")
}

// InitializeConfig initializes the doit configuration.
func initializeConfig() {
	loadDefaultSettings()
	LoadConfig()
	initFlags()

	if DoitCmd.PersistentFlags().Lookup("access-token").Changed {
		viper.Set("access-token", Token)
	}

	if DoitCmd.PersistentFlags().Lookup("output").Changed {
		viper.Set("output", Output)
	}
}

type flagOpt func(c *Command, name, key string)

func requiredOpt() flagOpt {
	return func(c *Command, name, key string) {
		c.MarkFlagRequired(key)
		key = requiredKey(key)
		viper.Set(key, true)

		u := c.Flag(name).Usage
		c.Flag(name).Usage = fmt.Sprintf("%s %s", u, requiredColor("(required)"))
	}
}

func requiredKey(key string) string {
	return fmt.Sprintf("%s.required", key)
}

// AddStringFlag adds a string flag to a command.
func AddStringFlag(cmd *Command, name, dflt, desc string, opts ...flagOpt) {
	fn := flagName(cmd, name)
	cmd.Flags().String(name, dflt, desc)

	for _, o := range opts {
		o(cmd, name, fn)
	}

	viper.BindPFlag(fn, cmd.Flags().Lookup(name))
}

// AddIntFlag adds an integr flag to a command.
func AddIntFlag(cmd *Command, name string, def int, desc string, opts ...flagOpt) {
	fn := flagName(cmd, name)
	cmd.Flags().Int(name, def, desc)
	viper.BindPFlag(fn, cmd.Flags().Lookup(name))

	for _, o := range opts {
		o(cmd, name, fn)
	}
}

// AddBoolFlag adds a boolean flag to a command.
func AddBoolFlag(cmd *Command, name string, def bool, desc string, opts ...flagOpt) {
	fn := flagName(cmd, name)
	cmd.Flags().Bool(name, def, desc)
	viper.BindPFlag(fn, cmd.Flags().Lookup(name))

	for _, o := range opts {
		o(cmd, name, fn)
	}
}

// AddStringSliceFlag adds a string slice flag to a command.
func AddStringSliceFlag(cmd *Command, name string, def []string, desc string, opts ...flagOpt) {
	fn := flagName(cmd, name)
	cmd.Flags().StringSlice(name, def, desc)
	viper.BindPFlag(fn, cmd.Flags().Lookup(name))

	for _, o := range opts {
		o(cmd, name, fn)
	}
}

func flagName(cmd *Command, name string) string {
	parentName := doit.NSRoot
	if cmd.Parent() != nil {
		parentName = cmd.Parent().Name()
	}

	return fmt.Sprintf("%s.%s.%s", parentName, cmd.Name(), name)
}

func cmdNS(cmd *cobra.Command) string {
	parentName := doit.NSRoot
	if cmd.Parent() != nil {
		parentName = cmd.Parent().Name()
	}

	return fmt.Sprintf("%s.%s", parentName, cmd.Name())
}

// CmdRunner runs a command and passes in a cmdConfig.
type CmdRunner func(*CmdConfig) error

// CmdConfig is a command configuration.
type CmdConfig struct {
	NS   string
	Doit doit.Config
	Out  io.Writer
	Args []string

	// services
	Keys              func() do.KeysService
	Sizes             func() do.SizesService
	Regions           func() do.RegionsService
	Images            func() do.ImagesService
	ImageActions      func() do.ImageActionsService
	FloatingIPs       func() do.FloatingIPsService
	FloatingIPActions func() do.FloatingIPActionsService
	Droplets          func() do.DropletsService
	DropletActions    func() do.DropletActionsService
	Domains           func() do.DomainsService
	Actions           func() do.ActionsService
	Account           func() do.AccountService
	Drives            func() do.DriveService
	DriveActions      func() do.DriveActionService
}

// NewCmdConfig creates an instance of a CmdConfig.
func NewCmdConfig(ns string, dc doit.Config, out io.Writer, args []string) *CmdConfig {
	godoClient := dc.GetGodoClient(Trace)

	return &CmdConfig{
		NS:   ns,
		Doit: dc,
		Out:  out,
		Args: args,

		Keys:              func() do.KeysService { return do.NewKeysService(godoClient) },
		Sizes:             func() do.SizesService { return do.NewSizesService(godoClient) },
		Regions:           func() do.RegionsService { return do.NewRegionsService(godoClient) },
		Images:            func() do.ImagesService { return do.NewImagesService(godoClient) },
		ImageActions:      func() do.ImageActionsService { return do.NewImageActionsService(godoClient) },
		FloatingIPs:       func() do.FloatingIPsService { return do.NewFloatingIPsService(godoClient) },
		FloatingIPActions: func() do.FloatingIPActionsService { return do.NewFloatingIPActionsService(godoClient) },
		Droplets:          func() do.DropletsService { return do.NewDropletsService(godoClient) },
		DropletActions:    func() do.DropletActionsService { return do.NewDropletActionsService(godoClient) },
		Domains:           func() do.DomainsService { return do.NewDomainsService(godoClient) },
		Actions:           func() do.ActionsService { return do.NewActionsService(godoClient) },
		Account:           func() do.AccountService { return do.NewAccountService(godoClient) },
		Drives:            func() do.DriveService { return do.NewDriveService(godoClient) },
		DriveActions:      func() do.DriveActionService { return do.NewDriveActionService(godoClient) },
	}

}

// Display displayes the output from a command.
func (c *CmdConfig) Display(d Displayable) error {
	dc := &displayer{
		ns:     c.NS,
		config: c.Doit,
		item:   d,
		out:    c.Out,
	}

	return dc.Display()
}

// CmdBuilder builds a new command.
func CmdBuilder(parent *Command, cr CmdRunner, cliText, desc string, out io.Writer, options ...cmdOption) *Command {
	cc := &cobra.Command{
		Use:   cliText,
		Short: desc,
		Long:  desc,
		Run: func(cmd *cobra.Command, args []string) {
			c := NewCmdConfig(
				cmdNS(cmd),
				doit.DoitConfig,
				out,
				args,
			)

			err := cr(c)
			checkErr(err, cmd)
		},
	}

	c := &Command{Command: cc}

	if parent != nil {
		parent.AddCommand(c)
	}

	for _, co := range options {
		co(c)
	}

	if cols := c.fmtCols; cols != nil {
		formatHelp := fmt.Sprintf("Columns for output in a comma seperated list. Possible values: %s",
			strings.Join(cols, ","))
		AddStringFlag(c, doit.ArgFormat, "", formatHelp)
		AddBoolFlag(c, doit.ArgNoHeader, false, "hide headers")
	}

	return c
}
