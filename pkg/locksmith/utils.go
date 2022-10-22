package locksmith

import "fmt"

func wrapError(err error, message string) error {
	return fmt.Errorf("%s; %w", message, err)
}
