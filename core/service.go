package core

// Service : Definition of service interface
type Service interface {
	CreateServer() (*Service, error)
	Start() error
	Close() error
}
