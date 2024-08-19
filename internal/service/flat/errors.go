package flat

import "errors"

var (
	ErrHouseNotFound = errors.New("house with the given ID does not exist")
	ErrFlatNotFound  = errors.New("flat with the given ID does not exist")
	ErrDuplicateFlat = errors.New("flat with this number already exists in the specified house")
)
