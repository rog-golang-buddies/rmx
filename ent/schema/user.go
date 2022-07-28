package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.Enum("role").
			Values("basic", "moderator", "admin").
			Default("basic"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
