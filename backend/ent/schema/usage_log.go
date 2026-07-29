package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

type UsageLog struct {
	ent.Schema
}

func (UsageLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("pet_id").NotEmpty(),
		field.String("model").NotEmpty(),
		field.Int("prompt_tokens").Default(0),
		field.Int("completion_tokens").Default(0),
		field.Float("tai_cost").Default(0),
		field.String("bounty_id").Optional(),
		field.Time("created_at").Default(time.Now),
	}
}

func (UsageLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pet_id"),
		index.Fields("created_at"),
	}
}
