package responder

import "fmt"

type RequiredAttributeMissing struct {
	field string
}

func (r *RequiredAttributeMissing) Error() string {
	return fmt.Sprintf("missing required message attribute: \"%s\"", r.field)
}
