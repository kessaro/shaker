package shaker

import "fmt"

func ErrNotFound() errNotFound {
	return errNotFound{}
}

func ErrNotFoundf(resourceName string) errNotFound {
	return errNotFound{
		ResourceName: resourceName,
	}
}

type errNotFound struct {
	ResourceName string
}

func (e errNotFound) Error() string {
	if e.ResourceName == "" {
		e.ResourceName = "resource"
	}
	return fmt.Sprintf("%s not found", e.ResourceName)
}
