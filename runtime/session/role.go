package session

// Role is a representation of a nameable role.
type Role interface {
	Name() string // User readable name.
}

// role implements a session endpoint role.
type role struct {
	name string
}

// Name returns the name of a session endpoint role.
func (r role) Name() string {
	return r.name
}

// NewRole creates a new Role using the given name.
func NewRole(name string) Role {
	return &paramRole{name: name}
}
