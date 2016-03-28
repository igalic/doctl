package godo

import (
	"fmt"
	"time"
)

const (
	storageBasePath  = "v2/storage"
	storageDrivePath = storageBasePath + "/drives"
	storageSnapPath  = storageBasePath + "/snapshots"
)

// StorageService is an interface for interfacing with the storage
// endpoints of the Digital Ocean API.
// See: https://developers.digitalocean.com/documentation/v2#storage
type StorageService interface {
	ListDrives(string, *ListOptions) ([]Drive, *Response, error)
	GetDrive(string) (*Drive, *Response, error)
	CreateDrive(*DriveCreateRequest) (*Drive, *Response, error)
	DeleteDrive(string) (*Response, error)
}

// StorageServiceOp handles communication with the storage Drives related methods of the
// DigitalOcean API.
type StorageServiceOp struct {
	client *Client
}

var _ StorageService = &StorageServiceOp{}

// Drive represents a Digital Ocean block store Drive.
type Drive struct {
	ID          string    `json:"id"`
	Region      Region    `json:"region"`
	Name        string    `json:"name"`
	SizeGB      int64     `json:"size_gigabytes"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	DropletID   int64     `json:"attached_to_droplet_id"`
}

func (f Drive) String() string {
	return Stringify(f)

}

type storageDrivesRoot struct {
	Drives []Drive `json:"drives"`
}

type storageDriveRoot struct {
	Drive *Drive `json:"drive"`
}

// DriveCreateRequest represents a request to create a block store
// Drive.
type DriveCreateRequest struct {
	Region      string `json:"region"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SizeGB      int64  `json:"size_gigabytes"`
}

// ListDrives lists all storage Drives.
func (svc *StorageServiceOp) ListDrives(region string, opt *ListOptions) ([]Drive, *Response, error) {
	path := storageDrivePath

	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err

	}

	req, err := svc.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err

	}

	if region != "" {
		q := req.URL.Query()
		q.Set("region", region)
		req.URL.RawQuery = q.Encode()

	}

	root := new(storageDrivesRoot)
	resp, err := svc.client.Do(req, root)
	if err != nil {
		return nil, resp, err

	}

	return root.Drives, resp, nil

}

// CreateDrive creates a storage Drive. The name must be unique.
func (svc *StorageServiceOp) CreateDrive(createRequest *DriveCreateRequest) (*Drive, *Response, error) {
	path := storageDrivePath

	req, err := svc.client.NewRequest("POST", path, createRequest)
	if err != nil {
		return nil, nil, err

	}

	root := new(storageDriveRoot)
	resp, err := svc.client.Do(req, root)
	if err != nil {
		return nil, resp, err

	}
	return root.Drive, resp, nil

}

// GetDrive retrieves an individual storage Drive.
func (svc *StorageServiceOp) GetDrive(id string) (*Drive, *Response, error) {
	path := fmt.Sprintf("%s/%s", storageDrivePath, id)

	req, err := svc.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err

	}

	root := new(storageDriveRoot)
	resp, err := svc.client.Do(req, root)
	if err != nil {
		return nil, resp, err

	}

	return root.Drive, resp, nil

}

// DeleteDrive deletes a storage Drive.
func (svc *StorageServiceOp) DeleteDrive(id string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", storageDrivePath, id)

	req, err := svc.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err

	}
	return svc.client.Do(req, nil)
}
