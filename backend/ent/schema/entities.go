package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Skill struct {
	ent.Schema
}

func (Skill) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "skills"}}
}

func (Skill) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(50),
		field.String("name_en").MaxLen(50).Optional(),
		field.String("quality").MaxLen(10).Default("normal"), // normal/advanced/super/limited
		field.String("category").MaxLen(30),                  // search/analysis/code/monitor/automation
		field.Text("description").Optional(),
		field.String("weapon_visual").MaxLen(50).Optional(),
		field.Int("power_bonus").Default(0),
		field.JSON("task_types", []string{}).Optional(),
		field.Bool("is_platform_exclusive").Default(false),
		field.Int("total_supply").Default(-1), // -1 = unlimited
		field.Int("remaining").Default(0),
		field.Float("price_tai").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}

// === PetSkill (equipped skills on a pet) ===

type PetSkill struct {
	ent.Schema
}

func (PetSkill) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "pet_skills"}}
}

func (PetSkill) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("pet_id", uuid.UUID{}),
		field.UUID("skill_id", uuid.UUID{}),
		field.Int("slot_index").Default(0),
		field.String("source").MaxLen(20).Default("purchase"), // preinstalled/book_drop/purchase/breed_inherit
		field.Time("equipped_at").Default(func() interface{} { return nil }),
	}
}

// === Trade ===

type Trade struct {
	ent.Schema
}

func (Trade) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "trades"}}
}

func (Trade) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("item_type").MaxLen(10), // pet/skill
		field.UUID("item_id", uuid.UUID{}),
		field.UUID("seller_id", uuid.UUID{}),
		field.UUID("buyer_id", uuid.UUID{}),
		field.Float("price").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("currency").MaxLen(10).Default("TAI"),
		field.Float("fee_amount").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.Float("fee_rate").SchemaType(map[string]string{dialect.Postgres: "decimal(5,4)"}).Default(0.03),
		field.Bool("is_market_maker").Default(false),
		field.String("chain_tx_hash").MaxLen(64).Optional(),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}

// === Order ===

type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "orders"}}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.String("item_type").MaxLen(10),
		field.UUID("item_id", uuid.UUID{}),
		field.String("side").MaxLen(4), // buy/sell
		field.Float("price").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("currency").MaxLen(10).Default("TAI"),
		field.String("status").MaxLen(10).Default("open"), // open/filled/cancelled/expired
		field.Time("filled_at").Optional(),
		field.Time("expires_at").Optional(),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}

// === Bounty ===

type Bounty struct {
	ent.Schema
}

func (Bounty) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "bounties"}}
}

func (Bounty) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("publisher_id", uuid.UUID{}),
		field.String("title").MaxLen(200),
		field.Text("description").Optional(),
		field.String("difficulty").MaxLen(1).Default("D"), // D/C/B/A/S
		field.Int("required_int").Default(0),
		field.JSON("required_skills", []string{}).Optional(),
		field.Float("reward_usdt").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("estimated_api_cost").SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).Default(0),
		field.String("status").MaxLen(20).Default("open"),
		field.UUID("assigned_pet_id", uuid.UUID{}).Optional(),
		field.Text("result").Optional(),
		field.Text("reviewer_note").Optional(),
		field.Time("completed_at").Optional(),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}

// === Guild ===

type Guild struct {
	ent.Schema
}

func (Guild) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "guilds"}}
}

func (Guild) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").MaxLen(50),
		field.UUID("leader_id", uuid.UUID{}),
		field.Int("level").Default(1),
		field.Int64("reputation").Default(0),
		field.Int("member_count").Default(1),
		field.Int("max_members").Default(50),
		field.Int("season_rank").Default(0),
		field.Time("created_at").Default(func() interface{} { return nil }),
	}
}
