package commands

import (
	"fmt"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"

	"github.com/spf13/cobra"
)

// Drive creates the Drive command
// NOTE: This command will currently only work for those in the
// block storage private beta on DigitalOcean.
func Drive() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "drive",
			Short: "drive commands",
			Long:  "drive is used to access drive commands",
		},
	}

	cmdDriveList := CmdBuilder(cmd, RunDriveList, "list", "list drive", Writer,
		aliasOpt("ls"), displayerType(&drive{}))

	AddStringFlag(cmdDriveList, doit.ArgDriveRegion, "", "Drive Region")

	cmdDriveCreate := CmdBuilder(cmd, RunDriveCreate, "create [name]", "create a drive", Writer,
		aliasOpt("c"), displayerType(&drive{}))

	AddIntFlag(cmdDriveCreate, doit.ArgDriveSize, 100, "Size of the drive (GiB)",
		requiredOpt())
	AddStringFlag(cmdDriveCreate, doit.ArgDriveDesc, "", "Drive Description",
		requiredOpt())
	AddStringFlag(cmdDriveCreate, doit.ArgDriveRegion, "", "Drive Region",
		requiredOpt())

	CmdBuilder(cmd, RunDriveDelete, "delete [ID]", "delete a drive", Writer,
		aliasOpt("rm"))

	cmdDriveGet := CmdBuilder(cmd, RunDriveGet, "get", "get a drive", Writer, aliasOpt("g"),
		displayerType(&drive{}))

	AddStringFlag(cmdDriveGet, doit.ArgDriveID, "", "id to fetch", requiredOpt())

	AddStringFlag(cmdDriveGet, "region", "", "region the drive is in", requiredOpt())

	return cmd

}

func RunDriveList(c *CmdConfig) error {
	region, err := c.Doit.GetString(c.NS, doit.ArgDriveRegion)
	if err != nil {
		return err

	}
	al := c.Drives()
	d, err := al.List(region)
	if err != nil {
		return err

	}
	item := &drive{drives: d}
	return c.Display(item)

}

func RunDriveCreate(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doit.NewMissingArgsErr(c.NS)

	}

	size, err := c.Doit.GetInt(c.NS, doit.ArgDriveSize)
	if err != nil {
		return err

	}

	name := c.Args[0]

	desc, err := c.Doit.GetString(c.NS, doit.ArgDriveDesc)
	if err != nil {
		return err
	}

	region, err := c.Doit.GetString(c.NS, doit.ArgDriveRegion)
	if err != nil {
		return err

	}

	var createDrive godo.DriveCreateRequest

	createDrive.Name = name
	createDrive.SizeGB = int64(size)
	createDrive.Description = desc
	createDrive.Region = region

	al := c.Drives()

	d, err := al.CreateDrive(&createDrive)
	if err != nil {
		return err

	}
	item := &drive{drives: []do.Drive{*d}}
	return c.Display(item)

}

func RunDriveDelete(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doit.NewMissingArgsErr(c.NS)

	}

	id := c.Args[0]

	al := c.Drives()

	if err := al.DeleteDrive(id); err != nil {
		return err

	}

	fmt.Printf("Deleted Drive: %s\n", id)

	return nil

}

func RunDriveGet(c *CmdConfig) error {
	id, err := c.Doit.GetString(c.NS, doit.ArgDriveID)
	if err != nil {
		return err

	}

	region, err := c.Doit.GetString(c.NS, doit.ArgDriveRegion)
	if err != nil {
		return err

	}

	al := c.Drives()

	d, err := al.Get(id, region)
	if err != nil {
		return err

	}

	item := &drive{drives: []do.Drive{*d}}
	return c.Display(item)

}
