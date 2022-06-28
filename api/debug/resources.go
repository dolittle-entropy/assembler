package debug

import "dolittle.io/kokk/resources"

type Repository interface {
	List() []resources.Resource
	Get(id string) (*resources.Resource, error)
}
