package commands

import (
	"fmt"
	"strconv"

	"github.com/digitalocean/doctl"
	"github.com/spf13/cobra"
)

// DriveAction creates the drive command
// NOTE: This command will only work for those accepted
// into the block storage private beta on DigitalOcean
func DriveAction() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:   "drive-action",
			Short: "drive action commands",
			Long:  "drive-action is used to access drive action commands",
		},
	}

	CmdBuilder(cmd, RunDriveAttach, "attach <drive-id> <droplet-id>", "attach a drive", Writer,
		aliasOpt("a"))

	CmdBuilder(cmd, RunDriveDetach, "detach <drive-id>", "detach a drive", Writer,
		aliasOpt("d"))

	return cmd

}

func RunDriveAttach(c *CmdConfig) error {
	if len(c.Args) != 2 {
		doit.NewMissingArgsErr(c.NS)
	}

	aID := c.Args[0]
	dID, err := strconv.Atoi(c.Args[1])
	if err != nil {
		return err

	}

	al := c.DriveActions()

	if err := al.Attach(aID, dID); err != nil {
		return err

	}

	fmt.Printf("attached %s to %d\n", aID, dID)
	return nil

}

func RunDriveDetach(c *CmdConfig) error {
	if len(c.Args) == 0 {
		return doit.NewMissingArgsErr(c.NS)
	}

	aID := c.Args[0]

	al := c.DriveActions()

	if err := al.Detach(aID); err != nil {
		return err

	}

	fmt.Printf("detached %s\n", aID)
	return nil

}
