// Package config loads config/game.json. Every balance number lives there.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Cultivation struct {
	Realms        []string `json:"realms"`
	TiersPerRealm int      `json:"tiers_per_realm"`
	ExpBase       float64  `json:"exp_base"`
	RealmMult     float64  `json:"realm_mult"`
	TierStep      float64  `json:"tier_step"`
	ScaleMin      float64  `json:"scale_min"`
	ScaleMax      float64  `json:"scale_max"`
}

type Sword struct {
	StartCount         int      `json:"start_count"`
	MinCount           int      `json:"min_count"`
	MaxSwords          []int    `json:"max_swords"`
	SwordsPerRing      int      `json:"swords_per_ring"`
	Length             float64  `json:"length"`
	Width              float64  `json:"width"`
	HiltGap            float64  `json:"hilt_gap"`
	RingGrowth         float64  `json:"ring_growth"`
	RingGrowthMode     string   `json:"ring_growth_mode"`
	HiltStep           float64  `json:"hilt_step"`
	TiltDeg            float64  `json:"tilt_deg"`
	TiltTweenSecs      float64  `json:"tilt_tween_secs"`
	RingTints          []string `json:"ring_tints"`
	RevSecs            float64  `json:"rev_secs"`
	FlipCooldown       float64  `json:"flip_cooldown"`
	ClashUnit          int      `json:"clash_unit"`
	ClashKnockback     float64  `json:"clash_knockback"`
	ClashKnockbackSecs float64  `json:"clash_knockback_secs"`
	DropBundleMax      int      `json:"drop_bundle_max"`
	ClashRatioBase     int      `json:"clash_ratio_base"`
	ClashRatioPerRealm int      `json:"clash_ratio_per_realm"`
	TweenSecs          float64  `json:"tween_secs"`
}

type Player struct {
	BodyR          float64 `json:"body_r"`
	MoveBase       float64 `json:"move_base"`
	PickupR        float64 `json:"pickup_r"`
	InvulnOnHit    float64 `json:"invuln_on_hit"`
	InvulnOnSpawn  float64 `json:"invuln_on_spawn"`
	AttackerShare  float64 `json:"attacker_share"`
	DropShare      float64 `json:"drop_share"`
	KillBonusExp   float64 `json:"kill_bonus_exp"`
	DeathSwordDrop float64 `json:"death_sword_drop"`
	OrbOwnerSecs   float64 `json:"orb_owner_secs"`
	CameraViewH    float64 `json:"camera_view_h"`
	CameraExp      float64 `json:"camera_exp"`
	CameraReachMul float64 `json:"camera_reach_mult"`
	HitDropsRealm  bool    `json:"hit_drops_realm"`
	Knockback      float64 `json:"knockback"`
	KnockbackSecs  float64 `json:"knockback_secs"`
	DeathAnimSecs  float64 `json:"death_anim_secs"`
	NameMin        int     `json:"name_min"`
	NameMax        int     `json:"name_max"`
}

type Buff struct {
	Mult    float64 `json:"mult"`
	Secs    float64 `json:"secs"`
	RevealR float64 `json:"reveal_r"`
	// Ẩn Thân only (spec §6.2): the buff never stacks past MaxSecs, viewers see the player
	// at VisibleAlpha instead of not at all, and bots only notice them within
	// BotDetectMult × their normal sight range.
	MaxSecs       float64 `json:"max_secs"`
	VisibleAlpha  float64 `json:"visible_alpha"`
	BotDetectMult float64 `json:"bot_detect_mult"`
}

type Items struct {
	Weights        map[string]int     `json:"weights"`
	CountPerRoom   int                `json:"count_per_room"`
	RespawnSecs    float64            `json:"respawn_secs"`
	SwordAmount    int                `json:"sword_item_amount"`
	BundleAmount   int                `json:"bundle_amount"`
	OvercapExp     float64            `json:"overcap_exp"`
	ChestEverySecs float64            `json:"chest_every_secs"`
	ChestItems     int                `json:"chest_items"`
	SpawnMinDist   float64            `json:"spawn_min_dist"`
	ChestWarnSecs  float64            `json:"chest_warn_secs"`
	ChestHP        int                `json:"chest_hp"`
	ItemR          float64            `json:"item_r"`
	OrbValues      map[string]float64 `json:"orb_values"`
	OrbR           float64            `json:"orb_r"`
	ScatterR       float64            `json:"scatter_r"`
	MaxOrbs        int                `json:"max_orbs"`
}

// ExpOrbs is the ambient EXP spread over the whole map (spec §6.1).
type ExpOrbs struct {
	Cell          float64   `json:"cell"`
	Values        []float64 `json:"values"`
	Weights       []int     `json:"weights"`
	RespawnSecs   float64   `json:"respawn_secs"`
	RespawnRegion float64   `json:"respawn_region"`
}

type MonsterKind struct {
	HP        int     `json:"hp"`
	Exp       float64 `json:"exp"`
	Drop      float64 `json:"drop"`
	DropItems int     `json:"drop_items"`
	Weight    int     `json:"weight"`
	Radius    float64 `json:"radius"`
	Max       int     `json:"max"`
}

type Monsters struct {
	HitCooldown float64                `json:"hit_cooldown"`
	WanderSpeed float64                `json:"wander_speed"`
	FleeSpeed   float64                `json:"flee_speed"`
	FleeSecs    float64                `json:"flee_secs"`
	Kinds       map[string]MonsterKind `json:"kinds"`
}

type Obstacles struct {
	FootprintW     float64        `json:"footprint_w"`
	FootprintH     float64        `json:"footprint_h"`
	OcclusionAlpha float64        `json:"occlusion_alpha"`
	RegrowSecs     float64        `json:"regrow_secs"`
	RegrowWarnSecs float64        `json:"regrow_warn_secs"`
	HP             map[string]int `json:"hp"`
	StallItemDrop  float64        `json:"stall_item_drop"`
	HitCooldown    float64        `json:"hit_cooldown"`
}

type Zones struct {
	MudMult             float64 `json:"mud_mult"`
	WindWith            float64 `json:"wind_with"`
	WindAgainst         float64 `json:"wind_against"`
	WindWithMaxDeg      float64 `json:"wind_with_max_deg"`
	WindAgainstMinDeg   float64 `json:"wind_against_min_deg"`
	BushRevealR         float64 `json:"bush_reveal_r"`
	BushRevealOnHitSecs float64 `json:"bush_reveal_on_hit_secs"`
}

type Gates struct {
	Radius       float64 `json:"radius"`
	TransitSecs  float64 `json:"transit_secs"`
	ExitWarnSecs float64 `json:"exit_warn_secs"`
	CooldownSecs float64 `json:"cooldown_secs"`
	FeedSecs     float64 `json:"feed_secs"`
}

type Room struct {
	MaxHumans      int     `json:"max_humans"`
	BotFillTarget  int     `json:"bot_fill_target"`
	MapSize        float64 `json:"map_size"`
	MonsterCount   int     `json:"monster_count"`
	EmptyCloseSecs float64 `json:"empty_close_secs"`
	SafeSpawnR     float64 `json:"safe_spawn_r"`
}

type Bots struct {
	SpawnEverySecs float64  `json:"spawn_every_secs"`
	SpawnBatch     int      `json:"spawn_batch"`
	RespawnSecs    float64  `json:"respawn_secs"`
	RealmWeights   []int    `json:"realm_weights"`
	DecideMin      float64  `json:"decide_min"`
	DecideMax      float64  `json:"decide_max"`
	ReactionMin    float64  `json:"reaction_min"`
	ReactionMax    float64  `json:"reaction_max"`
	HuntR          float64  `json:"hunt_r"`
	HuntRatio      float64  `json:"hunt_ratio"`
	FleeR          float64  `json:"flee_r"`
	FleeRatio      float64  `json:"flee_ratio"`
	PathGrid       float64  `json:"path_grid"`
	AmbushR        float64  `json:"ambush_r"`
	AmbushSecs     float64  `json:"ambush_secs"`
	Names          []string `json:"names"`
}

type Mood struct {
	MinSecs          float64 `json:"min_secs"`
	HitSecs          float64 `json:"hit_secs"`
	LevelupSecs      float64 `json:"levelup_secs"`
	BreakthroughSecs float64 `json:"breakthrough_secs"`
	RealmDropSecs    float64 `json:"realm_drop_secs"`
	RespawnSecs      float64 `json:"respawn_secs"`
	NearR            float64 `json:"near_r"`
}

// Minimap is the client-side minimap (spec §12.1); the server only passes it through.
type Minimap struct {
	SizePC     int  `json:"size_pc"`
	SizeMobile int  `json:"size_mobile"`
	ShowTop1   bool `json:"show_top1"`
	ShowChest  bool `json:"show_chest"`
	ShowOthers bool `json:"show_others"`
}

// Score is the Uy Danh scoring table (spec §11.1).
type Score struct {
	HitBase              float64 `json:"hit_base"`
	KillBonus            float64 `json:"kill_bonus"`
	UpsetPerRealm        float64 `json:"upset_per_realm"`
	BreakthroughPerRealm float64 `json:"breakthrough_per_realm"`
	Top1PerSec           float64 `json:"top1_per_sec"`
	BotVictimMult        float64 `json:"bot_victim_mult"`
	SameVictimCooldown   float64 `json:"same_victim_cooldown_secs"`
}

type Net struct {
	TickHz             int     `json:"tick_hz"`
	SnapshotHz         int     `json:"snapshot_hz"`
	InterpMs           int     `json:"interp_ms"`
	AOIMargin          float64 `json:"aoi_margin"`
	AOIAspect          float64 `json:"aoi_aspect"`
	LeaderboardHz      int     `json:"leaderboard_hz"`
	MaxMsgsPerSec      int     `json:"max_msgs_per_sec"`
	MaxConnsPerIP      int     `json:"max_conns_per_ip"`
	LeaderboardCacheSc float64 `json:"leaderboard_cache_secs"`
}

// Config mirrors config/game.json.
type Config struct {
	Cultivation Cultivation     `json:"cultivation"`
	Sword       Sword           `json:"sword"`
	Player      Player          `json:"player"`
	Buffs       map[string]Buff `json:"buffs"`
	Items       Items           `json:"items"`
	ExpOrbs     ExpOrbs         `json:"exp_orbs"`
	Monsters    Monsters        `json:"monsters"`
	Obstacles   Obstacles       `json:"obstacles"`
	Zones       Zones           `json:"zones"`
	Gates       Gates           `json:"gates"`
	Room        Room            `json:"room"`
	Bots        Bots            `json:"bots"`
	Characters  []string        `json:"characters"`
	Mood        Mood            `json:"mood"`
	Titles      []string        `json:"titles"`
	TitleHold   float64         `json:"title_hold_secs"`
	Minimap     Minimap         `json:"minimap"`
	Score       Score           `json:"score"`
	Net         Net             `json:"net"`
}

// Load reads and parses a config file. It returns the parsed config and raw bytes
// (the raw bytes are served to clients at GET /api/config).
func Load(path string) (*Config, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read config: %w", err)
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}
	return &c, raw, nil
}

// FindRepoFile walks up from the working directory to locate a repo-relative path
// such as "config/game.json", so commands work from both repo root and server/.
func FindRepoFile(rel string) string {
	dir, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		dir = filepath.Dir(dir)
	}
	return rel
}

// MustLoadDefault loads config/game.json found from the working directory; for tests and tools.
func MustLoadDefault() *Config {
	c, _, err := Load(FindRepoFile("config/game.json"))
	if err != nil {
		panic(err)
	}
	return c
}
