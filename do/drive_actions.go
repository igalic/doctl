package do

import "github.com/digitalocean/godo"

// DriveActionService is an interface for interacting with DigitalOcean's account api.
type DriveActionService interface {
	Attach(string, int) error
	Detach(string) error
}

type driveActionService struct {
	client *godo.Client
}

var _ DriveActionService = &driveActionService{}

// NewAccountService builds an DriveActionService instance.
func NewDriveActionService(godoClient *godo.Client) DriveActionService {
	return &driveActionService{
		client: godoClient,
	}

}

func (a *driveActionService) Attach(driveID string, dropletID int) error {
	_, err := a.client.StorageActions.Attach(driveID, dropletID)
	return err

}

func (a *driveActionService) Detach(driveID string) error {
	_, err := a.client.StorageActions.Detach(driveID)
	return err

}
