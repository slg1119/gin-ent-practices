package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User는 저장 형태를 정의한다. 업무 규칙은 internal/user에 둔다.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("email").NotEmpty().MaxLen(254).Unique(),
		field.Bool("is_active").Default(true),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}
