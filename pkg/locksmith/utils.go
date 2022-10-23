package locksmith

import "fmt"

func WrapError(err error, message string) error {
	return fmt.Errorf("%s; %s", message, err.Error())
}
