package main

import (
	"fmt"
	"strings"
)

// Post-processing modes.
const (
	modeSprite = "sprite" // one object → fitted onto a transparent canvas
	modeSheet  = "sheet"  // N frames in a grid → frames/{anim}_{NN}.png
	modeTile   = "tile"   // opaque square texture
	modeWide   = "wide"   // opaque illustration, keep aspect
	modeRaw    = "raw"    // keep as-is (reference sheets)
)

// Job is one minigen request plus how to turn its image into game assets.
type Job struct {
	Name     string   `json:"name"`
	Priority int      `json:"priority"`
	Prompt   string   `json:"prompt"`
	Mode     string   `json:"mode"`
	Out      string   `json:"out"`             // file (sprite/tile/raw) or directory (sheet)
	Anim     string   `json:"anim,omitempty"`  // sheet: animation name
	Frames   int      `json:"frames,omitempty"`
	Cols     int      `json:"cols,omitempty"`
	Rows     int      `json:"rows,omitempty"`
	Names    []string `json:"names,omitempty"` // sheet: explicit output names instead of {anim}_{NN}
	W        int      `json:"w"`
	H        int      `json:"h"`
	Pivot    string   `json:"pivot"` // "feet" (y=460 on 512) | "center"
}

// Style lock from ART_PROMPTS §2. Gemini cannot output alpha, so we ask for pure white
// and remove it in post-processing.
const style = "STYLE: 2D mobile game sprite, chibi cartoon caricature, exaggerated comedic facial expressions, " +
	"big head small body, thick black outline, flat colors with simple cel shading, vibrant saturated palette, " +
	"clean vector-like look, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, " +
	"plain solid pure white background (#FFFFFF), no text, no watermark, no ground shadow, centered, full body visible."

const styleObject = "STYLE: 2D mobile game asset, cartoon, thick black outline, flat colors with simple cel shading, " +
	"vibrant saturated palette, clean vector-like look, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, " +
	"plain solid pure white background (#FFFFFF), no text, no watermark, no ground shadow, centered."

const styleOpaque = "STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, " +
	"clean vector-like look, no text, no watermark."

const negative = " Avoid: realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, " +
	"watermark, background scenery, inconsistent outfit."

type character struct{ id, desc, palette string }

var characters = []character{
	{"a_hon", "a loud-mouthed young street vendor wearing a giant steamed bun as a hat, sleeveless white undershirt, baggy red shorts, flip-flops, huge buck teeth", "#FFFFFF #E63946 #F4A261 #2B2D42"},
	{"ba_tam_loa", "a sturdy middle-aged fish market lady holding a megaphone, floral blouse, rubber apron, tall green rubber boots, bun hair with a pencil stuck in it", "#2A9D8F #E9C46A #F28482 #264653"},
	{"ong_ba_vong", "a sleepy skinny old man carrying a folded hammock on his back like a cape, white tank top, striped pajama pants, half-closed eyes, toothpick in mouth", "#8D99AE #EDF2F4 #6D597A #B56576"},
	{"dai_ca_dep", "a pompous gang boss in an oversized shiny purple suit, tiny black sunglasses, gold chain, wearing cheap honeycomb plastic sandals, slicked-back hair", "#7209B7 #FFD60A #000000 #F72585"},
	{"map_tra_sua", "a round chubby teen always clutching a giant bubble tea cup with a fat straw, oversized hoodie, cheeks puffed, tapioca pearls flying around", "#CDB4DB #A47148 #FFC8DD #3D405B"},
	{"co_ba_ganh", "a fast nimble young woman balancing a bamboo shoulder pole with two baskets of fruit, conical hat, simple brown blouse, determined eyebrows", "#BC6C25 #DDA15E #606C38 #FEFAE0"},
	{"su_phu_mi", "an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, orange robe, holding chopsticks like weapons", "#FB8500 #FFB703 #8ECAE6 #023047"},
	{"be_ti", "a very tiny man with absurdly huge bodybuilder muscles, baby face, sweatband, tiny shorts, flexing constantly", "#E76F51 #F4A261 #2A9D8F #1D3557"},
	{"shipper", "a delivery rider in a bright helmet and reflective vest carrying a big cube delivery box on his back, lightning bolt stickers, phone in hand", "#FFBE0B #FB5607 #3A86FF #222222"},
	{"cu_rua", "a slow ancient grandpa wearing a turtle shell backpack, long white eyebrows covering eyes, walking cane, peaceful smile", "#52B788 #D8F3DC #6F4E37 #F1FAEE"},
}

type anim struct {
	name           string
	frames, c, r   int
	poses          string
	expression     string
	priority       int
}

// Animation list of ART_PROMPTS §4.4.
var anims = []anim{
	{"idle", 4, 4, 1, "1) standing breathing out 2) breathing in, shoulders raised 3) yawning with mouth wide open 4) scratching head", "bored, sleepy", 1},
	{"run", 8, 4, 2, "cartoon run cycle with spinning legs, arms flailing wildly, mouth wide open screaming, hair flying; 8 phases of the run cycle (contact, down, pass, up for each leg)", "screaming AAAA", 1},
	{"chase", 8, 4, 2, "sneaky tiptoe run leaning forward, evil grin with squinting eyes, rubbing both hands together; 8 run phases", "evil grin", 3},
	{"flee", 8, 4, 2, "panicked run with both arms raised to the sky, sweat drops flying, bulging eyes, wobbly mouth; 8 run phases", "panicked", 3},
	{"hit", 4, 4, 1, "1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering", "dazed", 2},
	{"levelup", 6, 3, 2, "1) ready stance 2) aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance", "proud", 4},
	{"breakthrough", 10, 5, 2, "1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh", "intense", 4},
	{"realm_drop", 6, 3, 2, "1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up", "crying", 4},
	{"death", 8, 4, 2, "1) getting smashed 2) smashed harder 3) flattened like rice paper 4) completely flat 5) white ghost soul rising out 6) ghost waving goodbye floating up 7) ghost floating higher 8) ghost almost gone at the top", "comic defeat", 2},
	{"respawn", 6, 3, 2, "1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose", "determined", 4},
	{"teleport_out", 4, 4, 1, "1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left", "shocked", 4},
	{"teleport_in", 4, 4, 1, "1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy", "dizzy", 4},
}

func sheetPrompt(st, subject, animName string, n, c, r int, poses string) string {
	return fmt.Sprintf("%s Sprite sheet, %d columns by %d rows grid, %d frames of %s facing right in 3/4 top-down view, "+
		"animation \"%s\" in %d sequential frames read left-to-right, top-to-bottom: %s. "+
		"Identical character in every cell, same size, feet aligned on the same baseline, each frame clearly separated "+
		"by plenty of white space, no grid lines, no frame numbers.%s", st, c, r, n, subject, animName, n, poses, negative)
}

func characterJobs() []Job {
	var jobs []Job
	for _, ch := range characters {
		subject := ch.desc + ", color palette " + ch.palette
		jobs = append(jobs, Job{
			Name: "char_" + ch.id + "_ref", Priority: 9, Mode: modeRaw, Out: "characters/" + ch.id + "/ref_sheet.png",
			Prompt: style + " Character reference sheet of " + subject + ". Show 3 views in a row: front, 3/4 facing right, " +
				"side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). " +
				"Same outfit and colors in every view. White background separated panels." + negative,
		})
		for _, a := range anims {
			jobs = append(jobs, Job{
				Name: "char_" + ch.id + "_" + a.name, Priority: a.priority, Mode: modeSheet, Out: "characters/" + ch.id + "/frames",
				Anim: a.name, Frames: a.frames, Cols: a.c, Rows: a.r, W: 512, H: 512, Pivot: "feet",
				Prompt: sheetPrompt(style, subject+", expression "+a.expression, a.name, a.frames, a.c, a.r, a.poses),
			})
		}
	}
	return jobs
}

var monsters = []struct{ id, prompt string }{
	{"banh_bao_cau", "an angry living steamed bun creature with tiny legs, steam puffing from head, round belly area left blank in the center"},
	{"dep_lao", "a sentient flip-flop sandal creature hopping, googly eyes, sticking tongue out, blank round circle on its sole"},
	{"noi_com", "a rice cooker monster with its lid as a biting mouth full of rice teeth, blank round panel on front"},
	{"ga_hoang", "a panicking chubby chicken monster, feathers flying, bulging eyes, blank round badge on chest"},
	{"muoi", "a giant fat mosquito monster with a straw-like nose, tiny wings, smug face, blank round area on belly"},
	{"vua_ghe", "a towering boss made of stacked red plastic stools, crown of a plastic basin, angry eyes, blank big circle in the middle"},
}

func monsterJobs() []Job {
	var jobs []Job
	type ma struct {
		name         string
		n, c, r, pri int
		poses        string
	}
	list := []ma{
		{"idle", 4, 4, 1, 1, "1) leaning left 2) upright 3) leaning right 4) upright squash"},
		{"hurt", 2, 2, 1, 3, "1) flinching with angry tears 2) squashed in pain"},
		{"death", 6, 3, 2, 3, "1) inflating 2) cracking 3) bursting 4) exploding into glowing spirit orbs 5) orbs scattering 6) only sparkles left"},
	}
	for _, m := range monsters {
		for _, a := range list {
			jobs = append(jobs, Job{
				Name: "mon_" + m.id + "_" + a.name, Priority: a.pri, Mode: modeSheet, Out: "monsters/" + m.id + "/frames",
				Anim: a.name, Frames: a.n, Cols: a.c, Rows: a.r, W: 512, H: 512, Pivot: "feet",
				Prompt: sheetPrompt(style, m.prompt, a.name, a.n, a.c, a.r, a.poses),
			})
		}
	}
	return jobs
}

var swords = []string{
	"a pair of bamboo chopsticks tied together as a sword",
	"a kitchen cleaver with a wooden handle",
	"a peach wood taoist sword with red tassel",
	"a polished steel jian sword",
	"a flaming red-orange sword with fire trail",
	"an ice crystal blue sword with frost sparkles",
	"a crackling lightning yellow sword",
	"a dark purple cursed sword with smoky aura",
	"a radiant jade and gold celestial sword glowing white",
}

func swordJobs() []Job {
	var jobs []Job
	for i, s := range swords {
		jobs = append(jobs, Job{
			Name: fmt.Sprintf("sword_realm_%d", i), Priority: 0, Mode: modeSprite, Out: fmt.Sprintf("swords/realm_%d.png", i),
			W: 256, H: 256, Pivot: "center",
			Prompt: styleObject + " Single game weapon icon, " + s + ", perfectly vertical, tip pointing straight up, handle at the bottom, centered." + negative,
		})
	}
	return jobs
}

var buffIcons = []struct{ id, icon, color string }{
	{"spin", "a cyan spinning whirlwind around a small sword", "cyan"},
	{"haste", "a winged sandal", "orange"},
	{"magnet", "a red horseshoe magnet", "red"},
	{"long", "a very long thin sword with an up arrow", "purple"},
	{"shield", "a golden buddhist-style round shield", "gold"},
	{"exp2", "a glowing magic pill made of light", "green"},
	{"stealth", "a paper talisman with swirling ink", "grey"},
}

func itemJobs() []Job {
	jobs := []Job{{
		Name: "item_sword", Priority: 0, Mode: modeSprite, Out: "items/sword.png", W: 256, H: 256, Pivot: "center",
		Prompt: styleObject + " A glossy transparent soap bubble containing a small glowing sword, shiny highlight." + negative,
	}}
	for _, c := range []struct{ id, color string }{{"orb_small", "green"}, {"orb_medium", "blue"}, {"orb_large", "gold"}} {
		jobs = append(jobs, Job{Name: "item_" + c.id, Priority: 0, Mode: modeSprite, Out: "items/" + c.id + ".png", W: 128, H: 128, Pivot: "center",
			Prompt: styleObject + " A glowing " + c.color + " spirit orb gem, sparkly, round." + negative})
	}
	for _, b := range buffIcons {
		jobs = append(jobs, Job{Name: "buff_" + b.id, Priority: 0, Mode: modeSprite, Out: "buffs/" + b.id + ".png", W: 256, H: 256, Pivot: "center",
			Prompt: styleObject + " A glossy transparent soap bubble containing " + b.icon + ", tinted " + b.color + ", shiny highlight." + negative})
	}
	chest := "an ornate wooden treasure chest wrapped in red ribbons, glowing golden light leaking out"
	jobs = append(jobs,
		Job{Name: "chest_idle", Priority: 2, Mode: modeSheet, Out: "items/chest", Anim: "idle", Frames: 4, Cols: 4, Rows: 1, W: 512, H: 512, Pivot: "feet",
			Prompt: sheetPrompt(styleObject, chest, "idle", 4, 4, 1, "1) still 2) rattling left 3) still 4) rattling right")},
		Job{Name: "chest_open", Priority: 3, Mode: modeSheet, Out: "items/chest", Anim: "open", Frames: 6, Cols: 3, Rows: 2, W: 512, H: 512, Pivot: "feet",
			Prompt: sheetPrompt(styleObject, chest, "open", 6, 3, 2, "1) closed shaking 2) lid cracking open 3) lid bursting open with light 4) items flying out 5) light fading 6) open empty chest")},
	)
	return jobs
}

func mapJobs() []Job {
	tile := func(id, biome string) Job {
		return Job{Name: "tile_" + id, Priority: 1, Mode: modeTile, Out: "tiles/" + id + ".png", W: 512, H: 512,
			Prompt: styleOpaque + " Seamless tileable top-down ground texture, " + biome + ", square image, cartoon game style, " +
				"no objects, no characters, even lighting, edges must tile perfectly."}
	}
	jobs := []Job{
		tile("grass", "soft green grass with small clover patches and tiny flowers"),
		tile("concrete", "cracked old concrete street with faded painted lines"),
		tile("dirt", "dry dusty brown dirt with pebbles"),
		tile("mud", "muddy brown puddle texture with bubbles"),
		{Name: "zone_bush", Priority: 1, Mode: modeSprite, Out: "zones/bush.png", W: 512, H: 512, Pivot: "center",
			Prompt: styleObject + " Top-down dense leafy bush cluster, round, lush green." + negative},
		{Name: "zone_wind", Priority: 3, Mode: modeSheet, Out: "zones", Anim: "wind", Frames: 4, Cols: 4, Rows: 1, W: 256, H: 256, Pivot: "center",
			Prompt: sheetPrompt(styleObject, "a white cartoon wind gust arrow with swirl lines pointing right, light blue outline", "wind", 4, 4, 1, "the gust arrow sliding to the right in 4 steps")},
		{Name: "wall_border", Priority: 3, Mode: modeSprite, Out: "decor/wall_border.png", W: 1024, H: 256, Pivot: "center",
			Prompt: styleObject + " A seamless horizontal stone and bamboo fence segment, top-down 3/4 view, tileable left-right, wide." + negative},
	}
	for _, d := range []string{"a red plastic stool", "an old bicycle leaning", "a pile of watermelons", "a street food cart", "a banana tree",
		"a rusty water jar", "a broken TV", "a lazy sleeping dog", "a bush with flowers", "a stone lantern"} {
		id := decorID(d)
		jobs = append(jobs, Job{Name: "decor_" + id, Priority: 3, Mode: modeSprite, Out: "decor/" + id + ".png", W: 256, H: 256, Pivot: "center",
			Prompt: styleObject + " Top-down 3/4 view decoration prop, " + d + "." + negative})
	}
	return jobs
}

func decorID(d string) string {
	for _, p := range []string{"a ", "an "} {
		d = strings.TrimPrefix(d, p)
	}
	return strings.ReplaceAll(d, " ", "_")
}

var obstacles = []struct {
	id, desc string
	w, h     int
}{
	{"fence", "a straight horizontal bamboo fence segment tied with rope", 1024, 256},
	{"stall", "a street market stall with a striped canopy and fruit baskets", 512, 512},
	{"cart", "a wooden street food cart with a small stove", 512, 512},
	{"brick", "a short horizontal red brick wall segment with moss", 1024, 512},
	{"boulder", "a huge round grey boulder with cartoon crack lines", 1024, 1024},
	{"stone_wall", "an ancient mossy horizontal stone wall with carved patterns", 1024, 512},
}

func obstacleJobs() []Job {
	var jobs []Job
	for _, o := range obstacles {
		jobs = append(jobs, Job{
			Name: "obs_" + o.id + "_states", Priority: 1, Mode: modeSheet, Out: "obstacles/" + o.id, Anim: "states",
			Frames: 4, Cols: 2, Rows: 2, W: o.w, H: o.h, Pivot: "center",
			Names: []string{"intact", "cracked", "nearly_broken", "rubble"},
			Prompt: sheetPrompt(styleObject, "a top-down 3/4 view game obstacle, "+o.desc, "damage states", 4, 2, 2,
				"1) intact and solid 2) with visible cracks 3) heavily damaged, crumbling, pieces missing 4) flat pile of rubble and debris on the ground"),
		})
		if o.id == "stone_wall" {
			continue
		}
		jobs = append(jobs,
			Job{Name: "obs_" + o.id + "_break", Priority: 4, Mode: modeSheet, Out: "obstacles/" + o.id + "/break", Anim: "break",
				Frames: 6, Cols: 3, Rows: 2, W: o.w, H: o.h, Pivot: "center",
				Prompt: sheetPrompt(styleObject, o.desc, "break", 6, 3, 2, "1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile")},
			Job{Name: "obs_" + o.id + "_regrow", Priority: 4, Mode: modeSheet, Out: "obstacles/" + o.id + "/regrow", Anim: "regrow",
				Frames: 6, Cols: 3, Rows: 2, W: o.w, H: o.h, Pivot: "center",
				Prompt: sheetPrompt(styleObject, o.desc, "regrow", 6, 3, 2, "1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid")},
		)
	}
	return jobs
}

func gateJobs() []Job {
	var jobs []Job
	list := []struct {
		name         string
		n, c, r, pri int
		poses        string
	}{
		{"idle", 6, 3, 2, 1, "slowly swirling vortex, rotating a little more each frame"},
		{"warn", 4, 4, 1, 3, "flashing brighter and brighter, intense glow"},
		{"burst", 6, 3, 2, 3, "bursting with light rays expanding then fading"},
	}
	for _, color := range []string{"purple", "blue", "orange"} {
		for _, a := range list {
			jobs = append(jobs, Job{Name: "gate_" + color + "_" + a.name, Priority: a.pri, Mode: modeSheet, Out: "gates/" + color,
				Anim: a.name, Frames: a.n, Cols: a.c, Rows: a.r, W: 512, H: 512, Pivot: "center",
				Prompt: sheetPrompt(styleObject, "a top-down magical teleport portal on the ground, swirling "+color+
					" vortex inside a carved stone ring with taoist symbols, glowing", a.name, a.n, a.c, a.r, a.poses)})
		}
	}
	return jobs
}

func vfxJobs() []Job {
	list := []struct {
		id, desc string
		n, c, r  int
		pri      int
	}{
		{"clash_spark", "comic sword clash spark burst, white and yellow star shapes, expanding then fading", 6, 3, 2, 2},
		{"stun_stars", "ring of small yellow cartoon stars orbiting", 4, 4, 1, 3},
		{"levelup_aura", "rising golden aura ring with sparkles", 6, 3, 2, 2},
		{"lightning", "thick cartoon lightning bolt striking down from the top", 6, 3, 2, 3},
		{"death_smoke", "comic white-grey smoke puff cloud with dark outline, expanding then dissipating", 6, 3, 2, 2},
		{"run_dust", "small brown dust puffs behind running feet", 4, 4, 1, 4},
		{"sword_shatter", "a cartoon sword shattering into sharp shards flying outward", 6, 3, 2, 2},
		{"shield_break", "golden bubble shield cracking and bursting into glass shards", 6, 3, 2, 3},
	}
	var jobs []Job
	for _, v := range list {
		jobs = append(jobs, Job{Name: "vfx_" + v.id, Priority: v.pri, Mode: modeSheet, Out: "vfx/" + v.id, Anim: v.id,
			Frames: v.n, Cols: v.c, Rows: v.r, W: 256, H: 256, Pivot: "center",
			Prompt: sheetPrompt(styleObject, v.desc, v.id, v.n, v.c, v.r, "the effect progressing frame by frame")})
	}
	jobs = append(jobs, Job{Name: "vfx_crack", Priority: 3, Mode: modeSprite, Out: "vfx/crack.png", W: 256, H: 256, Pivot: "center",
		Prompt: styleObject + " Thin dark grey crack lines overlay pattern like broken glass, sparse." + negative})
	return jobs
}

var badgeThemes = []string{"bamboo green", "kitchen steel grey", "peach wood red", "polished silver", "flame orange",
	"ice crystal blue", "lightning yellow", "cursed dark purple", "celestial jade and gold"}

func uiJobs() []Job {
	var jobs []Job
	for i, t := range badgeThemes {
		jobs = append(jobs, Job{Name: fmt.Sprintf("ui_badge_%d", i), Priority: 2, Mode: modeSprite, Out: fmt.Sprintf("ui/badge_%d.png", i),
			W: 256, H: 256, Pivot: "center",
			Prompt: styleObject + " Round game badge emblem for a cultivation rank, " + t + " theme, with a small sword symbol, no letters." + negative})
	}
	for _, m := range []string{"gold", "silver", "bronze"} {
		jobs = append(jobs, Job{Name: "ui_medal_" + m, Priority: 2, Mode: modeSprite, Out: "ui/medal_" + m + ".png", W: 128, H: 128, Pivot: "center",
			Prompt: styleObject + " A " + m + " medal with a tiny sword emblem, no letters." + negative})
	}
	jobs = append(jobs,
		Job{Name: "ui_speech", Priority: 2, Mode: modeSprite, Out: "ui/speech_bubble.png", W: 512, H: 256, Pivot: "center",
			Prompt: styleObject + " An empty white comic speech bubble with thick black outline, wide oval, tail at bottom left, empty inside." + negative},
		Job{Name: "ui_scroll", Priority: 2, Mode: modeSprite, Out: "ui/scroll.png", W: 1024, H: 768, Pivot: "center",
			Prompt: styleObject + " An empty unrolled ancient paper scroll with wooden handles on left and right, horizontal, nothing written." + negative},
		Job{Name: "ui_home_bg", Priority: 2, Mode: modeWide, Out: "ui/home_bg.png", W: 1920, H: 1080,
			Prompt: styleOpaque + " Wide 16:9 illustration of a chaotic cartoon Vietnamese village square with many chibi characters " +
				"spinning flying swords around them, sunny, bamboo and street stalls, no text."},
	)
	return jobs
}

func allJobs() []Job {
	var jobs []Job
	for _, f := range []func() []Job{swordJobs, itemJobs, characterJobs, monsterJobs, mapJobs, obstacleJobs, gateJobs, vfxJobs, uiJobs} {
		jobs = append(jobs, f()...)
	}
	return jobs
}
