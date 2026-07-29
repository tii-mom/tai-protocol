package schema

import (
	"time"

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
		field.String("species").MaxLen(30),
		field.String("quality").MaxLen(10).Default("common"), // common/rare/epic/legendary/mythic
		field.Int("generation").Default(0),
		field.Float("growth_rate").SchemaType(map[string]string{dialect.Postgres: "decimal(4,2)"}).Default(1.0),
		field.Int("apt_hp").Default(20),
		field.Int("apt_atk").Default(20),
		field.Int("apt_def").Default(20),
		field.Int("apt_spd").Default(20),
		field.Int("apt_int").Default(20),
		field.Int("skill_slots").Default(2),
		field.String("personality").MaxLen(20).Default("balanced"),
		field.Int("level").Default(1),
		field.Int64("exp").Default(0),
		field.Int("mood").Default(80),
		field.Int("energy").Default(100),
		field.String("status").MaxLen(20).Default("idle"), // idle/working/breeding/trading/resting
		field.Float("tai_balance").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.String("image_url").MaxLen(255).Optional(),
		field.Bool("is_on_chain").Default(false),
		field.String("chain_nft_address").MaxLen(64).Optional().Nillable(),
		field.Time("chain_minted_at").Optional().Nillable(),
		field.Time("breed_cooldown_until").Optional().Nillable(),
		field.Float("total_earned_tai").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Float("total_spent_tai").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Int64("total_tokens_used").Default(0),
		field.Int("total_tasks_done").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Pet) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).Ref("pets").Unique().Field("owner_id").Required(),
		edge.To("skills", PetSkill.Type),
	}
}

func (Pet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id"),
		index.Fields("quality"),
		index.Fields("status"),
		index.Fields("species"),
	}
}
