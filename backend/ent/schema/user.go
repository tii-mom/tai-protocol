package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "users"}}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int64("tg_user_id").Unique(),
		field.String("tg_username").MaxLen(100).Optional(),
		field.String("wallet_address").MaxLen(64).Optional(),
		field.Float("balance_tai").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Float("balance_usdt").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Float("frozen_balance").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.String("role").MaxLen(20).Default("user"),
		field.String("referral_code").MaxLen(20).Unique().Optional(),
		field.UUID("referred_by", uuid.UUID{}).Optional(),
		field.Int("invite_count").Default(0),
		field.Float("total_earned").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.UUID("guild_id", uuid.UUID{}).Optional(),
		field.String("status").MaxLen(20).Default("active"),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("pets", Pet.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tg_user_id").Unique(),
		index.Fields("referral_code"),
	}
}
