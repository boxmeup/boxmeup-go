package locations

import (
	"time"

	"github.com/cjsaylor/boxmeup-go/models/users"
)

// Location structure
type Location struct {
	ID         int
	User       users.User
	UUID       string
	Name       string
	IsMappable bool
	Address    string
	Created    time.Time
	Modified   time.Time
}
