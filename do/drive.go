package do

import "github.com/digitalocean/godo"

// Drive is a wrapper for godo.Drive.
type Drive struct {
	*godo.Drive
}

// DriveService is an interface for interacting with DigitalOcean's account api.
type DriveService interface {
	List(region string) ([]Drive, error)
	CreateDrive(*godo.DriveCreateRequest) (*Drive, error)
	DeleteDrive(string) error
	Get(string, string) (*Drive, error) // second string is region, ignored now

}

type driveService struct {
	client *godo.Client
}

var _ DriveService = &driveService{}

// NewAccountService builds an NewDriveService instance.
func NewDriveService(godoClient *godo.Client) DriveService {
	return &driveService{
		client: godoClient,
	}

}

func (a *driveService) List(region string) ([]Drive, error) {
	f := func(opt *godo.ListOptions) ([]interface{}, *godo.Response, error) {
		list, resp, err := a.client.Storage.ListDrives(region, opt)
		if err != nil {
			return nil, nil, err

		}

		si := make([]interface{}, len(list))
		for i := range list {
			si[i] = list[i]

		}

		return si, resp, err

	}

	si, err := PaginateResp(f)
	if err != nil {
		return nil, err

	}

	list := make([]Drive, len(si))
	for i := range si {
		a := si[i].(godo.Drive)
		list[i] = Drive{Drive: &a}

	}

	return list, nil

}

func (a *driveService) CreateDrive(r *godo.DriveCreateRequest) (*Drive, error) {
	al, _, err := a.client.Storage.CreateDrive(r)
	if err != nil {
		return nil, err

	}

	return &Drive{Drive: al}, nil

}

func (a *driveService) DeleteDrive(id string) error {

	_, err := a.client.Storage.DeleteDrive(id)
	if err != nil {
		return err

	}

	return nil

}

func (a *driveService) Get(id, region string) (*Drive, error) {
	d, _, err := a.client.Storage.GetDrive(id)
	if err != nil {
		return nil, err

	}

	return &Drive{Drive: d}, nil

}
