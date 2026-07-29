package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

type Bounty struct {
	ent.Schema
}

func (Bounty) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.String("publisher_id").NotEmpty(),
		field.Enum("difficulty").Values("D", "C", "B", "A", "S").Default("D"),
		field.Enum("status").Values("open", "accepted", "submitted", "confirmed", "expired", "cancelled").Default("open"),
		field.String("acceptor_pet_id").Optional(),
		field.Float("reward_tai").Default(0),
		field.Text("submission").Optional(),
		field.Time("created_at").Default(time.Now),
		field.Time("expires_at").Optional(),
		field.Time("completed_at").Optional(),
	}
}

func (Bounty) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("publisher_id"),
	}
}
