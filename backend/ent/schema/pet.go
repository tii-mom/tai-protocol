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

type Pet struct {
	ent.Schema
}

func (Pet) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "pets"}}
}

func (Pet) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("owner_id", uuid.UUID{}),
		field.String("name").MaxLen(50).Default("未命名"),
		field.String("species").MaxLen(30), // dragon/falcon/wolf/cat/rabbit/turtle
		field.String("quality").MaxLen(10).Default("N"), // N/R/SR/SSR/UR
		field.Int("generation").Default(0),
		field.Float("growth_rate").SchemaType(map[string]string{dialect.Postgres: "decimal(3,2)"}).Default(1.0),
		field.Int("apt_atk").Default(0),
		field.Int("apt_def").Default(0),
		field.Int("apt_spd").Default(0),
		field.Int("apt_int").Default(0),
		field.Int("skill_slots").Default(1),
		field.String("personality").MaxLen(20).Default("loyal"),
		field.Int("level").Default(1),
		field.Int64("exp").Default(0),
		field.Int("mood").Default(100),
		field.Int("energy").Default(100),
		field.String("status").MaxLen(20).Default("idle"),
		field.String("sprite_id").MaxLen(50).Optional(),
		field.Bool("is_on_chain").Default(false),
		field.String("chain_nft_address").MaxLen(64).Optional(),
		field.Time("chain_minted_at").Optional(),
		field.Time("breed_cooldown_until").Optional(),
		field.Float("total_earned_usdt").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Float("total_spent_tai").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Time("created_at").Default(func() interface{} { return nil }), // auto
		field.Time("updated_at").Default(func() interface{} { return nil }),
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("pets").Unique().Field("owner_id"),
		edge.To("skills", PetSkill.Type),
	}
}

func (Pet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id"),
		index.Fields("quality"),
		index.Fields("status"),
	}
}
