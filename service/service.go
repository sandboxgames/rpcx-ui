package service

type Registry interface {
	InitRegistry()
	FetchServices() []*Service
	DeactivateService(name, address string) error
	ActivateService(name, address string) error
	UpdateMetadata(name, address string, metadata string) error
}

// Service is a service endpoint
type Service struct {
	ID       string
	Name     string
	Address  string
	Metadata string
	State    string
	Group    string
}

