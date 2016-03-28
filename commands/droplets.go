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
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"

	"github.com/digitalocean/doctl"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// Droplet creates the droplet command.
func Droplet() *Command {
	cmd := &Command{
		Command: &cobra.Command{
			Use:     "droplet",
			Aliases: []string{"d"},
			Short:   "droplet commands",
			Long:    "droplet is used to access droplet commands",
		},
		DocCategories: []string{"droplet"},
		IsIndex:       true,
	}

	CmdBuilder(cmd, RunDropletActions, "actions <droplet id>", "droplet actions", Writer,
		aliasOpt("a"), displayerType(&action{}), docCategories("droplet"))

	CmdBuilder(cmd, RunDropletBackups, "backups <droplet id>", "droplet backups", Writer,
		aliasOpt("b"), displayerType(&image{}), docCategories("droplet"))

	cmdDropletCreate := CmdBuilder(cmd, RunDropletCreate, "create NAME [NAME ...]", "create droplet", Writer,
		aliasOpt("c"), displayerType(&droplet{}), docCategories("droplet"))
	AddStringSliceFlag(cmdDropletCreate, doit.ArgSSHKeys, []string{}, "SSH Keys or fingerprints")
	AddStringFlag(cmdDropletCreate, doit.ArgUserData, "", "User data")
	AddStringFlag(cmdDropletCreate, doit.ArgUserDataFile, "", "User data file")
	AddBoolFlag(cmdDropletCreate, doit.ArgCommandWait, false, "Wait for droplet to be created")
	AddStringFlag(cmdDropletCreate, doit.ArgRegionSlug, "", "Droplet region",
		requiredOpt())
	AddStringFlag(cmdDropletCreate, doit.ArgSizeSlug, "", "Droplet size",
		requiredOpt())
	AddBoolFlag(cmdDropletCreate, doit.ArgBackups, false, "Backup droplet")
	AddBoolFlag(cmdDropletCreate, doit.ArgIPv6, false, "IPv6 support")
	AddBoolFlag(cmdDropletCreate, doit.ArgPrivateNetworking, false, "Private networking")
	AddStringFlag(cmdDropletCreate, doit.ArgImage, "", "Droplet image",
		requiredOpt())
	// Note: Adding Drives will only work for accounts in the block storage private beta
	AddStringSliceFlag(cmdDropletCreate, doit.ArgDriveList, []string{}, "Drives to attach")

	CmdBuilder(cmd, RunDropletDelete, "delete ID [ID|Name ...]", "Delete droplet by id or name", Writer,
		aliasOpt("d", "del", "rm"), docCategories("droplet"))

	CmdBuilder(cmd, RunDropletGet, "get", "get droplet", Writer,
		aliasOpt("g"), displayerType(&droplet{}), docCategories("droplet"))

	CmdBuilder(cmd, RunDropletKernels, "kernels <droplet id>", "droplet kernels", Writer,
		aliasOpt("k"), displayerType(&kernel{}), docCategories("droplet"))

	cmdRunDropletList := CmdBuilder(cmd, RunDropletList, "list [GLOB]", "list droplets", Writer,
		aliasOpt("ls"), displayerType(&droplet{}), docCategories("droplet"))
	AddStringFlag(cmdRunDropletList, doit.ArgRegionSlug, "", "Droplet region")

	CmdBuilder(cmd, RunDropletNeighbors, "neighbors <droplet id>", "droplet neighbors", Writer,
		aliasOpt("n"), displayerType(&droplet{}), docCategories("droplet"))

	CmdBuilder(cmd, RunDropletSnapshots, "snapshots <droplet id>", "snapshots", Writer,
		aliasOpt("s"), displayerType(&image{}), docCategories("droplet"))

	return cmd
}

// RunDropletActions returns a list of actions for a droplet.
func RunDropletActions(c *CmdConfig) error {

	ds := c.Droplets()

	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	list, err := ds.Actions(id)
	item := &action{actions: list}
	return c.Display(item)
}

// RunDropletBackups returns a list of backup images for a droplet.
func RunDropletBackups(c *CmdConfig) error {

	ds := c.Droplets()

	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	list, err := ds.Backups(id)
	if err != nil {
		return err
	}

	item := &image{images: list}
	return c.Display(item)
}

// RunDropletCreate creates a droplet.
func RunDropletCreate(c *CmdConfig) error {

	if len(c.Args) < 1 {
		return doit.NewMissingArgsErr(c.NS)
	}

	region, err := c.Doit.GetString(c.NS, doit.ArgRegionSlug)
	if err != nil {
		return err
	}

	size, err := c.Doit.GetString(c.NS, doit.ArgSizeSlug)
	if err != nil {
		return err
	}

	backups, err := c.Doit.GetBool(c.NS, doit.ArgBackups)
	if err != nil {
		return err
	}

	ipv6, err := c.Doit.GetBool(c.NS, doit.ArgIPv6)
	if err != nil {
		return err
	}

	privateNetworking, err := c.Doit.GetBool(c.NS, doit.ArgPrivateNetworking)
	if err != nil {
		return err
	}

	keys, err := c.Doit.GetStringSlice(c.NS, doit.ArgSSHKeys)
	if err != nil {
		return err
	}

	sshKeys := extractSSHKeys(keys)

	userData, err := c.Doit.GetString(c.NS, doit.ArgUserData)
	if err != nil {
		return err
	}

	filename, err := c.Doit.GetString(c.NS, doit.ArgUserDataFile)
	if err != nil {
		return err
	}

	userData, err = extractUserData(userData, filename)
	if err != nil {
		return err
	}

	driveList, err := c.Doit.GetStringSlice(c.NS, doit.ArgDriveList)
	if err != nil {
		return err
	}
	drives := extractDrives(driveList)

	var createImage godo.DropletCreateImage

	imageStr, err := c.Doit.GetString(c.NS, doit.ArgImage)
	if i, err := strconv.Atoi(imageStr); err == nil {
		createImage = godo.DropletCreateImage{ID: i}
	} else {
		createImage = godo.DropletCreateImage{Slug: imageStr}
	}

	wait, err := c.Doit.GetBool(c.NS, doit.ArgCommandWait)
	if err != nil {
		return err
	}

	ds := c.Droplets()

	var wg sync.WaitGroup
	errs := make(chan error, len(c.Args))
	for _, name := range c.Args {
		dcr := &godo.DropletCreateRequest{
			Name:              name,
			Region:            region,
			Size:              size,
			Image:             createImage,
			Backups:           backups,
			IPv6:              ipv6,
			PrivateNetworking: privateNetworking,
			SSHKeys:           sshKeys,
			UserData:          userData,
			Drives:            drives,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			d, err := ds.Create(dcr, wait)
			if err != nil {
				errs <- err
				return
			}

			item := &droplet{droplets: do.Droplets{*d}}
			c.Display(item)
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func extractSSHKeys(keys []string) []godo.DropletCreateSSHKey {
	sshKeys := []godo.DropletCreateSSHKey{}

	for _, rawKey := range keys {
		rawKey = strings.TrimPrefix(rawKey, "[")
		rawKey = strings.TrimSuffix(rawKey, "]")

		keys := strings.Split(rawKey, ",")

		for _, k := range keys {
			if i, err := strconv.Atoi(k); err == nil {
				if i > 0 {
					sshKeys = append(sshKeys, godo.DropletCreateSSHKey{ID: i})
				}
				continue
			}

			if k != "" {
				sshKeys = append(sshKeys, godo.DropletCreateSSHKey{Fingerprint: k})
			}
		}
	}

	return sshKeys
}

func extractDrives(driveList []string) []godo.DropletDriveRequest {
	drives := []godo.DropletDriveRequest{}

	for _, rawDrive := range driveList {
		rawDrive = strings.TrimPrefix(rawDrive, "[")
		rawDrive = strings.TrimSuffix(rawDrive, "]")

		list := strings.Split(rawDrive, ",")
		for _, v := range list {
			drives = append(drives, godo.DropletDriveRequest{ID: v})
		}
	}
	return drives
}

func extractUserData(userData, filename string) (string, error) {
	if userData == "" && filename != "" {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			return "", err
		}
		userData = string(data)
	}

	return userData, nil
}

// RunDropletDelete destroy a droplet by id.
func RunDropletDelete(c *CmdConfig) error {

	ds := c.Droplets()

	if len(c.Args) < 1 {
		return doit.NewMissingArgsErr(c.NS)
	}

	listedDroplets := false
	list := do.Droplets{}

	for _, idStr := range c.Args {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			if !listedDroplets {
				list, err = ds.List()
				if err != nil {
					return errors.New("unable to build list of droplets")
				}
				listedDroplets = true
			}

			var matchedDroplet *do.Droplet
			for _, d := range list {
				if d.Name == idStr {
					matchedDroplet = &d
					break
				}
			}

			if matchedDroplet == nil {
				return fmt.Errorf("unable to find droplet with name %q", idStr)
			}

			id = matchedDroplet.ID
		}

		err = ds.Delete(id)
		if err != nil {
			return fmt.Errorf("unable to delete droplet %d: %v", id, err)
		}

		fmt.Printf("deleted droplet %d\n", id)
	}

	return nil
}

// RunDropletGet returns a droplet.
func RunDropletGet(c *CmdConfig) error {
	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	ds := c.Droplets()

	d, err := ds.Get(id)
	if err != nil {
		return err
	}

	item := &droplet{droplets: do.Droplets{*d}}
	return c.Display(item)
}

// RunDropletKernels returns a list of available kernels for a droplet.
func RunDropletKernels(c *CmdConfig) error {

	ds := c.Droplets()
	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	list, err := ds.Kernels(id)
	if err != nil {
		return err
	}

	item := &kernel{kernels: list}
	return c.Display(item)
}

// RunDropletList returns a list of droplets.
func RunDropletList(c *CmdConfig) error {

	ds := c.Droplets()

	region, err := c.Doit.GetString(c.NS, doit.ArgRegionSlug)
	if err != nil {
		return err
	}

	matches := []glob.Glob{}
	for _, globStr := range c.Args {
		g, err := glob.Compile(globStr)
		if err != nil {
			return fmt.Errorf("unknown glob %q", globStr)
		}

		matches = append(matches, g)
	}

	var matchedList do.Droplets

	list, err := ds.List()
	if err != nil {
		return err
	}

	for _, droplet := range list {
		var skip = true
		if len(matches) == 0 {
			skip = false
		} else {
			for _, m := range matches {
				if m.Match(droplet.Name) {
					skip = false
				}
			}
		}

		if !skip && region != "" {
			if region != droplet.Region.Slug {
				skip = true
			}
		}

		if !skip {
			matchedList = append(matchedList, droplet)
		}
	}

	item := &droplet{droplets: matchedList}
	return c.Display(item)
}

// RunDropletNeighbors returns a list of droplet neighbors.
func RunDropletNeighbors(c *CmdConfig) error {

	ds := c.Droplets()

	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	list, err := ds.Neighbors(id)
	if err != nil {
		return err
	}

	item := &droplet{droplets: list}
	return c.Display(item)
}

// RunDropletSnapshots returns a list of available kernels for a droplet.
func RunDropletSnapshots(c *CmdConfig) error {

	ds := c.Droplets()
	id, err := getDropletIDArg(c.NS, c.Args)
	if err != nil {
		return err
	}

	list, err := ds.Snapshots(id)
	if err != nil {
		return err
	}

	item := &image{images: list}
	return c.Display(item)
}

func getDropletIDArg(ns string, args []string) (int, error) {
	if len(args) != 1 {
		return 0, doit.NewMissingArgsErr(ns)
	}

	return strconv.Atoi(args[0])
}
