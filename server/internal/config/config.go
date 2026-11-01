// Package config loads config/game.json. Every balance number lives there.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Cultivation struct {
	// TienOverflowScore: đã Đạo Tổ thì không còn cảnh giới để lên, nên mỗi điểm Tiên khí
	// nhặt thêm đổi thành ngần này điểm Uy Danh thay vì bỏ phí. 0 = tắt.
	// Hiện là 0,1 — tỉ lệ 10 Tiên khí ăn 1 Uy Danh (chủ dự án quyết 29/10/2026).
	TienOverflowScore float64  `json:"tien_overflow_score"`
	Realms            []string `json:"realms"`
	TiersPerRealm     int      `json:"tiers_per_realm"`
	ExpBase           float64  `json:"exp_base"`
	RealmMult         float64  `json:"realm_mult"`
	TierStep          float64  `json:"tier_step"`
	ScaleMin          float64  `json:"scale_min"`
	ScaleMax          float64  `json:"scale_max"`

	// Tiên giới (spec §3.4): 5 cảnh giới sau Đại Thừa, không chia tầng, chỉ lên được
	// bằng sát mạng chứ không dùng EXP.
	TienRealms []string `json:"tien_realms"`
	// TienKhiCeilings: TRẦN Tiên khí cộng dồn của từng cảnh giới Tiên (spec v5.7 §3.4).
	// [100,300,600,1000,1500] nghĩa là 1–100 Chân Tiên, 101–300 Kim Tiên, … 1001–1500 Đạo Tổ.
	//
	// Cảnh giới Tiên suy ra THẲNG từ tổng Tiên khí đang có, không cộng dồn từng chặng như
	// bản v5.6 (`tien_khi_per_step`). Khác biệt thật: bị chém thì Tiên khí về đáy cảnh giới
	// mới, và cảnh giới tự tụt theo — không còn hai nguồn sự thật để lệch nhau.
	TienKhiCeilings []float64 `json:"tien_khi_ceilings"`
	// TienEntryRequiresFullExp: phải ĐẦY thanh Đại Thừa tầng 10 mới nhặt được Tiên khí.
	TienEntryRequiresFullExp bool    `json:"tien_entry_requires_full_exp"`
	TienRobShare             float64 `json:"tien_rob_share"`
	TienDropShare            float64 `json:"tien_drop_share"`
	// TienDropScatterMin/Max: viên Tiên khí rơi ra khi bị chém văng xa ngần này (spec v5.9).
	// Phải xa hơn hẳn thân người Tiên giới: họ to gấp 5–6 lần, và vật phẩm vẽ DƯỚI nhân vật,
	// nên viên rơi gần là nằm khuất hẳn dưới xác — đó chính là lỗi "chém mà không thấy rơi".
	TienDropScatterMin float64 `json:"tien_drop_scatter_min"`
	TienDropScatterMax float64 `json:"tien_drop_scatter_max"`
	// TienDropUnits: mệnh giá viên Tiên khí rơi ra, LỚN TRƯỚC; TienDropMaxOrbs: trần số viên
	// mỗi lần rơi, vượt thì gộp phần nhỏ vào viên cuối (spec v6.4 §3.4).
	TienDropUnits   []float64 `json:"tien_drop_units"`
	TienDropMaxOrbs int       `json:"tien_drop_max_orbs"`
	TienScaleStep   float64   `json:"tien_scale_step"`
	TienAuraColors  []string  `json:"tien_aura_colors"`
}

type Sword struct {
	StartCount int   `json:"start_count"`
	MinCount   int   `json:"min_count"`
	MaxSwords  []int `json:"max_swords"`
	// RingStep: vòng thứ k chứa RingStep × k kiếm (vòng 1 = 10, vòng 2 = 20… spec §4.1a).
	RingStep       int      `json:"ring_capacity_step"`
	Length         float64  `json:"length"`
	Width          float64  `json:"width"`
	HiltGap        float64  `json:"hilt_gap"`
	RingGrowth     float64  `json:"ring_growth"`
	RingGrowthMode string   `json:"ring_growth_mode"`
	HiltStep       float64  `json:"hilt_step"`
	TiltDeg        float64  `json:"tilt_deg"`
	TiltTweenSecs  float64  `json:"tilt_tween_secs"`
	RingTints      []string `json:"ring_tints"`
	RevSecs        float64  `json:"rev_secs"`
	// FlipCooldown: chỉ giới hạn việc ĐỔI CHIỀU quay, để vòng kiếm không giật 30 lần/giây.
	FlipCooldown float64 `json:"flip_cooldown"`
	// ClashCooldown: giãn cách giữa hai lần TÍNH BỂ KIẾM của cùng một cặp (A, B).
	// 0 = mỗi tick một lần, tức mài hết một vòng 30 kiếm trong đúng 1 giây (spec §5.2).
	ClashCooldown      float64 `json:"clash_cooldown"`
	ClashUnit          int     `json:"clash_unit"`
	FlipOnlySameRealm  bool    `json:"flip_only_same_realm"`
	ClashKnockback     float64 `json:"clash_knockback"`
	ClashKnockbackSecs float64 `json:"clash_knockback_secs"`
	DropBundleMax      int     `json:"drop_bundle_max"`
	// Độ bền kiếm (spec v4.1 §5.2): mỗi cảnh giới bền hơn cảnh giới dưới 20%.
	// "compound" → k = 1.2^d; "linear" → k = 1 + 0.2 × d.
	DurabilityPerRealm float64 `json:"durability_per_realm"`
	DurabilityMode     string  `json:"durability_mode"`
	// BodyHitMaxRings: kiếm địch chỉ chém được vào thân khi nạn nhân còn tối đa bấy nhiêu
	// vòng (spec v4.5 §5.2 — thay cho lá chắn chặn đường đi của bản trước).
	BodyHitMaxRings int     `json:"body_hit_max_rings"`
	TweenSecs       float64 `json:"tween_secs"`
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
	// RobCapRealms: chém chết thì người chém nhận Linh khí tối đa tới sàn cảnh giới r+N
	// (spec v6.26 §5.1, N = 2). RobCapTienRealms: Tiên khí tối đa tới trần cảnh giới Tiên +N.
	RobCapRealms     int `json:"rob_cap_realms"`
	RobCapTienRealms int `json:"rob_cap_tien_realms"`
	// DeathOrbsMax: xác rơi tối đa ngần này viên Linh khí, mệnh giá lớn trước (40).
	DeathOrbsMax  int     `json:"death_orbs_max"`
	Knockback     float64 `json:"knockback"`
	KnockbackSecs float64 `json:"knockback_secs"`
	DeathAnimSecs float64 `json:"death_anim_secs"`
	NameMin       int     `json:"name_min"`
	NameMax       int     `json:"name_max"`
}

// Burn is Thiêu Đốt Tinh Huyết (spec §4.1c): hold to run faster, paying EXP every second.
// ExpFracPerSec is a fraction of the CURRENT tier's requirement, so the cost scales with
// cultivation instead of being free at high realms and crippling at low ones.
type Burn struct {
	ExpFracPerSec float64 `json:"exp_frac_per_sec"`
	// TienFracPerSec: người Tiên giới đốt **Tiên khí** thay cho EXP (spec v5.6 §4.1c) —
	// mỗi giây ngần này phần của BƯỚC Tiên khí hiện tại. Spec §16 ghi khoá này trong khối
	// `player` tên `burn_tien_frac_per_sec`; ở đây để cạnh `exp_frac_per_sec` cho khớp
	// cách config đang xếp (khối `burn` riêng), đúng như `burn_exp_per_sec` đã làm.
	TienFracPerSec float64 `json:"tien_frac_per_sec"`
	KeepRealm      bool    `json:"keep_realm"` // never burn below the floor of the current realm
}

// Speed: tốc độ di chuyển CỘNG trên base rồi kẹp (spec v6.25 §4.1):
//
//	move_speed = move_base × clamp(1 + Σ thưởng + Σ phạt, floor, cap)
//
// Mỗi nguồn cộng đúng một lần (hai vùng bùn chồng nhau vẫn chỉ −0.4).
type Speed struct {
	Cap       float64 `json:"cap"`         // trần thường (×2 base)
	CapGomDon float64 `json:"cap_gom_don"` // trần khi Shipper đang Gom Đơn (×3)
	Floor     float64 `json:"floor"`       // sàn khi bị làm chậm (không áp cho khống chế đứng im)
	Haste     float64 `json:"haste"`       // Thần Hành
	Burn      float64 `json:"burn"`        // Thiêu Đốt Tinh Huyết
	GomDon    float64 `json:"gom_don"`     // Gom Đơn Hỏa Tốc
	WindWith  float64 `json:"wind_with"`
	WindAgain float64 `json:"wind_against"`
	Mud       float64 `json:"mud"`  // bùn + nước
	Snow      float64 `json:"snow"` // bờ tuyết
	// SlowStatus: trạng thái "chậm" do kỹ năng (spec: "theo kỹ năng"). Hiện chưa kỹ năng nào
	// gây chậm; giữ khoá để kỹ năng sau dùng, cùng mức với bùn.
	SlowStatus float64 `json:"slow_status"`
}

type Buff struct {
	Mult    float64 `json:"mult"`
	Secs    float64 `json:"secs"`
	RevealR float64 `json:"reveal_r"`
	// Cương Kiếm only (spec v4.4 §6.2): kiếm cứng hơn khi va (DurabilityMult) và chém quái /
	// vật cản mạnh hơn (DamageMult).
	DurabilityMult float64 `json:"durability_mult"`
	DamageMult     float64 `json:"damage_mult"`
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

// ExpEconomy: mọi hệ số kinh tế Linh khí (spec v6.24 §6.0). Bảng nhịp §6.0 là CHUẨN; các
// số ở đây được `cmd/sim --farm` chỉnh cho trúng bảng (±25 %) và Nam Châm ≤ 1,5 lần.
type ExpEconomy struct {
	// OrbValueMult: viên EXP môi trường và EXP quái × OrbValueMult^r (r = cảnh giới của người
	// nhặt / người chém). Không nguồn nào khác nhân theo cảnh giới.
	OrbValueMult float64 `json:"orb_value_mult"`
	// PickupScaleExp: bán kính hút = pickup_r × scale^PickupScaleExp (0.5 = √scale, 1 = thẳng).
	PickupScaleExp float64 `json:"pickup_scale_exp"`
	// BotStealShare: phần EXP người chém nhận khi nạn nhân là BOT (người thật: attacker_share).
	BotStealShare float64 `json:"bot_steal_share"`
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
	FootprintW     float64 `json:"footprint_w"`
	FootprintH     float64 `json:"footprint_h"`
	OcclusionAlpha float64 `json:"occlusion_alpha"`
	RegrowSecs     float64 `json:"regrow_secs"`
	RegrowWarnSecs float64 `json:"regrow_warn_secs"`

	// MinRealm: **bậc** của từng loại vật cản — chỉ số cảnh giới TỐI THIỂU để chém được
	// (0 = Luyện Khí … 8 = Đại Thừa, 9 = Chân Tiên … 13 = Đạo Tổ). Spec v6.7/v6.8 §8.2.
	//
	// Thay hẳn cho `hp` + `dmg_*` của bản cũ. Vật cản không còn máu: đủ bậc thì chạm là
	// biến mất, chưa đủ thì kiếm đi xuyên không tác dụng. Bỏ HP không phải để đơn giản hoá
	// mà vì luật cũ cho người cảnh giới thấp **mài mãi cũng phá được** mọi thứ — tức bậc
	// không chặn được gì, chỉ làm chậm.
	MinRealm map[string]int `json:"min_realm"`
	// CorridorMin: khe hở tối thiểu giữa hai vật cản CÙNG BẬC, khoá là bậc dạng chuỗi.
	// Người to nhất bị bậc đó chặn phải lọt qua được (§8.1b).
	CorridorMin map[string]float64 `json:"corridor_min"`
	// HardBuffExtraTier: Cương Kiếm cho chém thêm ngần này bậc (§6.2, v6.8 = 1).
	HardBuffExtraTier int `json:"hard_buff_extra_tier"`
	// CutHintCooldownSecs: chữ "Cần {cảnh giới}" hiện tối đa 1 lần / ngần này giây / vật cản.
	CutHintCooldownSecs float64 `json:"cut_hint_cooldown_secs"`

	StallItemDrop float64 `json:"stall_item_drop"`
	HitCooldown   float64 `json:"hit_cooldown"`
}

// BorderKind là loại vật cản BIÊN KHU — không có bậc, không bao giờ chém được (§8.2).
//
// Tách thành loại riêng thay vì để chung `stone_wall`: biên khu hiện là 4 dãy
// `stone_wall` dài 11 400, mà `stone_wall` có bậc 5 — tức từ Luyện Hư trở lên là **chém
// thủng được biên khu và đi ra ngoài bản đồ**.
const BorderKind = "border"

// ObstacleTier trả về bậc của một loại vật cản. Loại không khai báo coi như bậc 0 (ai cũng
// chém được) — thà để người chơi phá được một thứ lẽ ra không phá được, còn hơn dựng một
// bức tường vô hình không ai hiểu vì sao không qua nổi.
// MaxTier là bậc cao nhất CÓ THẬT trong khu (không tính biên khu).
//
// Dùng để chốt trần bậc sạp của Cô Ba: spec nói "tối đa 13", mà 13 là bậc thạch bi — tức
// "cao nhất trong bảng". Đọc từ bảng thay vì viết 13 vào mã: đổi bảng mà quên đổi mã là
// sạp của Đạo Tổ bỗng thành bất khả phá với chính Đạo Tổ.
func (o Obstacles) MaxTier() int {
	top := 0
	for _, t := range o.MinRealm {
		if t < TierBorder && t > top {
			top = t
		}
	}
	return top
}

func (o Obstacles) ObstacleTier(kind string) int {
	if kind == BorderKind {
		return TierBorder
	}
	if t, ok := o.MinRealm[kind]; ok {
		return t
	}
	return 0
}

// TierBorder là bậc của biên khu: cao hơn mọi cảnh giới nên không ai chém được.
const TierBorder = 99

// Ice: mặt băng ở Băng Nguyệt Hồ (spec §8.0). Di chuyển có quán tính:
//
//	vận tốc mới = vận tốc cũ × friction + hướng input × accel × dt
//
// Tốc độ tới hạn = accel × dt / (1 − friction). Với friction 0,92, dt 1/30 và accel 468 thì
// tới hạn ≈ 195 = đúng `player.move_base`, nên chạy thẳng trên băng vẫn nhanh bằng trên
// đất — cái khó là **dừng và đổi hướng**, đúng ý spec.
type Ice struct {
	Friction float64 `json:"friction"`
	Accel    float64 `json:"accel"`
}

// SkillU là một giá trị trong `skills.ranges_u`.
//
// Khối đó trong spec §16 trộn ba kiểu: hệ số nhân với U (`ru_ngu_r: 4`), số thuần không
// nhân (`ru_ngu_secs: 2.5`, `bay_sap_count: 5`) và **một boolean** (`bay_sap_owner_passes:
// true`). Giữ nguyên hình dạng đó thay vì đổi `true` thành `1.0` trong config: `config/
// game.json` phải dán khớp được với spec, nếu không thì mỗi lần chủ dự án sửa spec rồi chép
// sang là server từ chối nạp và không ai biết vì sao.
//
// Việc phân biệt "nhân với U hay không" nằm ở CHỖ DÙNG (`Game.skillRange` vs đọc thẳng),
// không nằm ở kiểu dữ liệu — tên khoá đã nói rõ: `_r`, `_cone`, `_len`, `_dist`, `_w`,
// `_wave`, `_push`, `_follow`, `_seek` là khoảng cách; `_secs`, `_count`, `_steal`,
// `_channel`, `_bonus`, `_of_own_max`, `_passes` thì không.
type SkillU float64

func (u *SkillU) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case "true":
		*u = 1
		return nil
	case "false":
		*u = 0
		return nil
	}
	var f float64
	if err := json.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("ranges_u: %q không phải số hay boolean: %w", b, err)
	}
	*u = SkillU(f)
	return nil
}

func (u SkillU) Float() float64 { return float64(u) }
func (u SkillU) Int() int       { return int(u) }
func (u SkillU) Bool() bool     { return u != 0 }

// Skills: mọi con số của 10 kỹ năng nhân vật (spec §6.3). Bán kính và tầm xa còn được
// nhân theo scale người dùng lúc thi hành.
type Skills struct {
	// Cooldowns: hồi chiêu theo nhân vật (spec v6.25 §6.3) — `[giây hồi 1 lượt, lượt tối đa]`.
	Cooldowns map[string][2]float64 `json:"cooldowns"`
	// StartCharges: vào trận / hồi sinh có sẵn ngần này lượt (1).
	StartCharges int `json:"start_charges"`
	// TienExtraCharges: Tiên giới +1 lượt tích tối đa; rớt về Đại Thừa thì cắt bớt.
	TienExtraCharges int     `json:"tien_extra_charges"`
	LockSecs         float64 `json:"lock_secs"` // khoá giữa hai lần dùng (3 s)
	KnockSecs        float64 `json:"knock_secs"`

	// ——— Phạm vi kỹ năng theo đơn vị U (spec v6.5 §6.3) ———
	//
	// `U = pickup_r × scale_NGƯỜI_DÙNG` — đúng bán kính vòng hút Linh khí của họ. Mọi chiều
	// dài, bán kính và khoảng đẩy của kỹ năng là `ranges_u[...] × U`.
	//
	// Vì sao U chứ không phải tầm chém: vòng hút chỉ phụ thuộc cảnh giới (scale), không đổi
	// khi mất hay thêm kiếm — nên kỹ năng ổn định và người chơi đoán được. Tầm chém thì co
	// giãn theo số kiếm đang có, dùng nó thì cùng một kỹ năng lúc xa lúc gần.
	//
	// RangeUnit chỉ để GHI lại U lấy từ đâu; `config.Validate` từ chối giá trị lạ, nên
	// không ai đổi được nó thành thứ code không hiểu mà vẫn chạy.
	RangeUnit string `json:"range_unit"`
	// PushSecs: mọi đòn đẩy của kỹ năng diễn ra trong ngần này giây (spec: 0.3s), trượt dọc
	// vật cản, không xuyên. Khác `KnockSecs` (0.25) của đòn đẩy khi bể kiếm — hai thứ khác
	// nhau nên để hai khoá, đừng gộp.
	PushSecs float64           `json:"push_secs"`
	RangesU  map[string]SkillU `json:"ranges_u"`

	// BanhBaoMaxShare: hồi ngay ngần này phần của số kiếm TỐI ĐA ở cảnh giới hiện tại
	// (spec v5.5 §6.3). Bản v5.0 hồi theo số kiếm vừa bể trong 10s — đổi rồi, không còn
	// sổ ghi kiếm bể nữa.
	BanhBaoMaxShare float64 `json:"banh_bao_max_share"`

	HutMiConeDeg float64 `json:"hut_mi_cone_deg"`

	DanEmCount int     `json:"dan_em_count"`
	DanEmSecs  float64 `json:"dan_em_secs"`
	// DanEmSwordShare: mỗi đàn em có ngần này phần số kiếm HIỆN TẠI của đại ca (làm tròn
	// xuống, tối thiểu 1). Đại ca không bị trừ kiếm — đây là bản sao, không phải chia của.
	DanEmSwordShare float64 `json:"dan_em_sword_share"`
	// DanEmScale: đàn em nhỏ hơn đại ca (0.7 = nhỏ hơn 30%).
	DanEmScale float64 `json:"dan_em_scale"`

	TranChauConeDeg float64 `json:"tran_chau_cone_deg"`
	TranChauSecs    float64 `json:"tran_chau_secs"`
	// TranChauNoPickup: người đang bị bịt mắt không nhặt được gì (spec v5.5 §6.3).
	TranChauNoPickup  bool    `json:"tran_chau_no_pickup"`
	TranChauBotVision float64 `json:"tran_chau_bot_vision"`

	DashSecs   float64 `json:"dash_secs"`
	SlickCount int     `json:"slick_count"`
	SlickSecs  float64 `json:"slick_secs"`
	SlipSecs   float64 `json:"slip_secs"`

	BraceSecs       float64 `json:"brace_secs"`
	BraceDurability float64 `json:"brace_durability"`

	// GhostUnstickSecs: hết Giao Hỏa Tốc mà đang đứng trong vật cản thì được đẩy mượt ra
	// chỗ trống gần nhất trong ngần này giây (spec v5.5 §6.3) — không dịch tức thời, vì
	// nhảy một phát là client thấy giật y như lỗi §8.2a vừa sửa xong.
	GhostUnstickSecs float64 `json:"ghost_unstick_secs"`
	// BlindNoBodyHit: người bị bịt mắt (Thằng Mập) — kiếm không chém trúng thân ai, va kiếm
	// vẫn tính (spec v6.27 §6.3).
	BlindNoBodyHit bool `json:"blind_no_body_hit"`

	ShellSecs float64 `json:"shell_secs"`
}

type Zones struct {
	WindWithMaxDeg      float64 `json:"wind_with_max_deg"`
	WindAgainstMinDeg   float64 `json:"wind_against_min_deg"`
	BushRevealR         float64 `json:"bush_reveal_r"`
	BushRevealOnHitSecs float64 `json:"bush_reveal_on_hit_secs"`
}

type Gates struct {
	// RandomExit: bước vào một cổng thì ra ở một cổng BẤT KỲ khác cùng mạng, thay vì cổng
	// đôi cố định. Đặt false để quay về kiểu cũ.
	RandomExit bool `json:"random_exit"`
	// ExitPool: "all" (mặc định) = ra ở cổng bất kỳ trong cả thế giới; "other_zone" = luôn
	// sang khu khác (spec §8.4).
	ExitPool string  `json:"exit_pool"`
	Radius   float64 `json:"radius"`
	// HeavenRadius: bán kính kích hoạt Thiên Môn — to gấp đôi cổng thường (spec §8.4a).
	HeavenRadius float64 `json:"heaven_radius"`
	TransitSecs  float64 `json:"transit_secs"`
	ExitWarnSecs float64 `json:"exit_warn_secs"`
	CooldownSecs float64 `json:"cooldown_secs"`
	FeedSecs     float64 `json:"feed_secs"`
}

type Room struct {
	// Zones: số khu mỗi phòng (spec §8.0). Để 1 là quay về thế giới một bản đồ như trước.
	Zones int `json:"zones"`
	// ZoneMinChars: sàn nhân vật mỗi khu — bot lấp TỪNG KHU cho đủ số này (spec v6.28 §9.3).
	// MaxHumans: trần người thật/phòng, bot không tính. Cả hai là MẶC ĐỊNH của môi trường
	// (`ApplyEnv` chép từ `room_env`); admin ghi đè lúc chạy (room.Manager.SetSettings).
	ZoneMinChars   int     `json:"zone_min_chars"`
	MaxHumans      int     `json:"max_humans"`
	MapSize        float64 `json:"map_size"`
	MonsterCount   int     `json:"monster_count"`
	EmptyCloseSecs float64 `json:"empty_close_secs"`
	SafeSpawnR     float64 `json:"safe_spawn_r"`
}

type Bots struct {
	SpawnEverySecs float64 `json:"spawn_every_secs"`
	SpawnBatch     int     `json:"spawn_batch"`
	RespawnSecs    float64 `json:"respawn_secs"`
	// Hành vi giống người khi chết (spec v6.1 §9.4). Bot cũ hồi sinh sau đúng 3 giây mọi
	// lần — nhìn kill feed là thấy nhịp máy.
	RespawnSecsRange     []float64 `json:"respawn_secs_range"`
	RejoinSameNameChance float64   `json:"rejoin_same_name_chance"`
	IdleAfterJoinChance  float64   `json:"idle_after_join_chance"`
	IdleAfterJoinSecs    []float64 `json:"idle_after_join_secs"`
	// Chế độ bot (spec v6.22 §9.5): "auto" | "on" | "off". Admin đổi lúc chạy; đây là giá
	// trị khi server khởi động. Bot sinh ra như người chơi mới (Luyện Khí 1, 1 kiếm, không
	// buff) — đã bỏ `spawn_rank` / `max_spawn_realm`.
	ModeDefault string `json:"mode_default"`
	// Tự động (spec v6.28 §9.3): người thật ≥ AutoOffAt → ngừng sinh bot; < AutoOffAt −
	// AutoOffHysteresis → sinh lại. Trễ để phòng quanh ngưỡng không bật/tắt liên tục.
	// AutoOffAt là mặc định môi trường (`room_env.bot_auto_off_at`), admin ghi đè lúc chạy.
	AutoOffAt         int `json:"auto_off_at"`
	AutoOffHysteresis int `json:"auto_off_hysteresis"`
	// WithdrawIdleSecs: "Rút bot ngay" chỉ rút bot không giao tranh trong ngần ấy giây.
	WithdrawIdleSecs float64  `json:"withdraw_idle_secs"`
	DecideMin        float64  `json:"decide_min"`
	DecideMax        float64  `json:"decide_max"`
	ReactionMin      float64  `json:"reaction_min"`
	ReactionMax      float64  `json:"reaction_max"`
	HuntR            float64  `json:"hunt_r"`
	HuntRatio        float64  `json:"hunt_ratio"`
	FleeR            float64  `json:"flee_r"`
	FleeRatio        float64  `json:"flee_ratio"`
	PathGrid         float64  `json:"path_grid"`
	AmbushR          float64  `json:"ambush_r"`
	AmbushSecs       float64  `json:"ambush_secs"`
	Names            []string `json:"names"`
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
// AltarKind is one Đài Tiên Khí type (spec §8.5).
type AltarKind struct {
	Count          int     `json:"count"`
	EverySecs      float64 `json:"every_secs"`
	Orbs           []int   `json:"orbs"` // [min, max] viên mỗi đợt thường
	OrbValue       float64 `json:"orb_value"`
	BurstEverySecs float64 `json:"burst_every_secs"`
	BurstSmall     int     `json:"burst_small"`
	BurstMedium    int     `json:"burst_medium"`
	MediumValue    float64 `json:"medium_value"`
	WarnSecs       float64 `json:"warn_secs"`
}

// Altars is the Đài Tiên Khí section (spec §8.5).
type Altars struct {
	Small      AltarKind `json:"small"`
	Center     AltarKind `json:"center"`
	SprayR     []float64 `json:"spray_r"`
	OrbLifeSec float64   `json:"orb_lifetime_secs"`
	ClearR     float64   `json:"clear_r"`
}

// Mapgen is the §8.1 generator settings.
type Mapgen struct {
	CenterPlazaR float64 `json:"center_plaza_r"`
	SectR        float64 `json:"sect_r"`
	SectInset    float64 `json:"sect_inset"`
	RoadW        float64 `json:"road_w"`
	RingRoadR    float64 `json:"ring_road_r"`
	Symmetry     string  `json:"symmetry"`
	Jitter       float64 `json:"jitter"`
	MinOpenRatio float64 `json:"min_open_ratio"`
	MaxRetries   int     `json:"max_retries"`
}

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
	// MaxRingFlips: tối đa bao nhiêu sự kiện va kiếm được gửi cho MỘT người trong một gói
	// (giữ cái gần họ nhất). Va kiếm không mang luật chơi — chiều quay mới nằm ở trường `d`
	// mỗi gói — nên đây thuần là ngân sách băng thông cho tia lửa. 0 = không giới hạn.
	MaxRingFlips       int     `json:"max_ring_flips"`
	TickHz             int     `json:"tick_hz"`
	SnapshotHz         int     `json:"snapshot_hz"`
	InterpMs           int     `json:"interp_ms"`
	AOIMargin          float64 `json:"aoi_margin"`
	AOIAspect          float64 `json:"aoi_aspect"`
	LeaderboardHz      int     `json:"leaderboard_hz"`
	MaxMsgsPerSec      int     `json:"max_msgs_per_sec"`
	MaxConnsPerIP      int     `json:"max_conns_per_ip"`
	LeaderboardCacheSc float64 `json:"leaderboard_cache_secs"`
	// CorrectionSnap / CorrectionTauMs: client dùng để sửa vị trí dự đoán (spec §8.2a).
	// Server không dùng, nhưng phải có ở đây thì `/api/config` mới gửi đi được.
	CorrectionSnap  float64 `json:"correction_snap"`
	CorrectionTauMs float64 `json:"correction_tau_ms"`
}

// Config mirrors config/game.json.
// Brand — GAME_SPEC §20. Mọi chỗ HIỂN THỊ tên game lấy từ đây.
//
// Để trong config thay vì hằng số trong code vì tên game xuất hiện ở bảy chỗ khác nhau
// (tiêu đề tab, màn hình tải, OG, manifest PWA, trang quản trị, trang số liệu, log server)
// và lần đổi tên vừa rồi cho thấy bỏ sót là chuyện bình thường. Một nguồn thì không bỏ sót.
type Brand struct {
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	NameVI    string `json:"name_vi"`
	TaglineVI string `json:"tagline_vi"`
}

type Config struct {
	Brand       Brand              `json:"brand"`
	Skills      Skills             `json:"skills"`
	Ice         Ice                `json:"ice"`
	Cultivation Cultivation        `json:"cultivation"`
	Sword       Sword              `json:"sword"`
	Player      Player             `json:"player"`
	Speed       Speed              `json:"speed"`
	Burn        Burn               `json:"burn"`
	Buffs       map[string]Buff    `json:"buffs"`
	Items       Items              `json:"items"`
	ExpOrbs     ExpOrbs            `json:"exp_orbs"`
	ExpEconomy  ExpEconomy         `json:"exp_economy"`
	Monsters    Monsters           `json:"monsters"`
	Obstacles   Obstacles          `json:"obstacles"`
	Zones       Zones              `json:"zones"`
	Gates       Gates              `json:"gates"`
	Room        Room               `json:"room"`
	RoomEnv     map[string]RoomEnv `json:"room_env"`
	Bots        Bots               `json:"bots"`
	Characters  []string           `json:"characters"`
	Mood        Mood               `json:"mood"`
	Titles      []string           `json:"titles"`
	TitleHold   float64            `json:"title_hold_secs"`
	Minimap     Minimap            `json:"minimap"`
	Altars      Altars             `json:"altars"`
	Mapgen      Mapgen             `json:"mapgen"`
	Score       Score              `json:"score"`
	Net         Net                `json:"net"`
}

// RoomEnv: mặc định phòng/bot của một môi trường (spec v6.28 §9.5). LoadtestedHumans là
// mức người thật/phòng đã qua loadtest trên máy của môi trường đó (PERF_LOG) — trang admin
// cảnh báo khi đặt trần cao hơn.
type RoomEnv struct {
	MaxHumans        int `json:"max_humans"`
	ZoneMinChars     int `json:"zone_min_chars"`
	BotAutoOffAt     int `json:"bot_auto_off_at"`
	LoadtestedHumans int `json:"loadtested_humans"`
}

// EnvStaging / EnvProduction: hai môi trường có khối `room_env` riêng.
const (
	EnvStaging    = "staging"
	EnvProduction = "production"
)

// EnvName: `APP_ENV` → tên khối `room_env`. Chỉ "staging" dùng khối staging; mọi giá trị
// khác (production, dev, rỗng) dùng production — máy :9999 là máy đã qua loadtest 50 người.
func EnvName(appEnv string) string {
	if strings.EqualFold(strings.TrimSpace(appEnv), EnvStaging) {
		return EnvStaging
	}
	return EnvProduction
}

// ApplyEnv chép mặc định của môi trường `env` vào Room/Bots, trả khối đã dùng. Không có
// khối thì giữ nguyên số trong `room` / `bots`.
func (c *Config) ApplyEnv(env string) RoomEnv {
	e, ok := c.RoomEnv[env]
	if !ok {
		return RoomEnv{MaxHumans: c.Room.MaxHumans, ZoneMinChars: c.Room.ZoneMinChars,
			BotAutoOffAt: c.Bots.AutoOffAt, LoadtestedHumans: c.Room.MaxHumans}
	}
	c.Room.MaxHumans, c.Room.ZoneMinChars, c.Bots.AutoOffAt = e.MaxHumans, e.ZoneMinChars, e.BotAutoOffAt
	return e
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
	// Tên bot nằm ở FILE RIÊNG `config/bot_names.txt` (spec v6.1 §9.4), không nằm trong
	// game.json: một nghìn dòng tên nhét vào JSON thì mỗi lần sửa một con số cân bằng là
	// phải cuộn qua chúng, và ai cũng ngại mở file.
	//
	// Thiếu file thì giữ danh sách trong game.json làm dự phòng — máy dev chưa kéo file mới
	// vẫn chạy được, chỉ là bot có ít tên hơn.
	if names := loadBotNames(filepath.Join(filepath.Dir(path), "bot_names.txt")); len(names) > 0 {
		c.Bots.Names = names
	}
	if err := c.checkCooldowns(); err != nil {
		return nil, nil, err
	}
	return &c, raw, nil
}

// checkCooldowns: mọi nhân vật phải có hồi chiêu > 0 và ≥ 1 lượt (spec v6.25 §6.3). Thiếu
// một dòng thì nhân vật đó im lặng không bao giờ có lượt — chặn ngay lúc nạp config.
func (c *Config) checkCooldowns() error {
	for _, ch := range c.Characters {
		cd, ok := c.Skills.Cooldowns[ch]
		if !ok || cd[0] <= 0 || cd[1] < 1 {
			return fmt.Errorf("config: skills.cooldowns[%q] thiếu hoặc sai (cần [giây > 0, lượt ≥ 1])", ch)
		}
	}
	return nil
}

// loadBotNames đọc file tên bot: mỗi dòng một tên, bỏ dòng trống và dòng bắt đầu bằng `#`.
func loadBotNames(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(b), "\n") {
		s := strings.TrimSpace(line)
		if s != "" && !strings.HasPrefix(s, "#") {
			out = append(out, s)
		}
	}
	return out
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
