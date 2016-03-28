package godo

import "fmt"

// StorageActionsService is an interface for interfacing with the
// storage actions endpoints of the Digital Ocean API.
// See: https://developers.digitalocean.com/documentation/v2#storage-actions
type StorageActionsService interface {
	Attach(driveID string, dropletID int) (*Response, error)
	Detach(driveID string) (*Response, error)
}

// StorageActionsServiceOp handles communication with the floating IPs
// action related methods of the DigitalOcean API.
type StorageActionsServiceOp struct {
	client *Client
}

// StorageAttachment represents the attachement of a block storage
// drive to a specific droplet under the device name.
type StorageAttachment struct {
	DropletID int `json:"droplet_id"`
}

// Attach a storage drive to a droplet, using the given device name.
func (s *StorageActionsServiceOp) Attach(driveID string, dropletID int) (*Response, error) {
	request := &ActionRequest{
		"droplet_id": dropletID,
		"drive_id":   driveID,
	}

	path := storageDriveActionPath()
	req, err := s.client.NewRequest("POST", path, request)
	if err != nil {
		return nil, err

	}
	return s.client.Do(req, nil)

}

// Detach a storage drive from a droplet.
func (s *StorageActionsServiceOp) Detach(driveID string) (*Response, error) {
	request := &ActionRequest{
		"drive_id": driveID,
	}

	path := storageDriveActionPath()
	req, err := s.client.NewRequest("DELETE", path, request)
	if err != nil {
		return nil, err

	}
	fmt.Println(req.URL)
	fmt.Println(req.Body)
	return s.client.Do(req, nil)

}

func storageDriveActionPath() string {
	return fmt.Sprintf("%s/attachments", storageDrivePath)

}
