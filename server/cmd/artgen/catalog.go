package main

import (
	"fmt"
	"strings"
)

// Post-processing modes.
const (
	modeSprite = "sprite" // one object → fitted onto a transparent canvas
	modeSheet  = "sheet"  // many frames / objects in rows → separate PNGs
	modeTile   = "tile"   // opaque square texture
	modeWide   = "wide"   // opaque illustration, keep aspect
	modeRaw    = "raw"    // keep as-is (reference sheets)
)

// Seg is one animation (or set of objects) inside a sheet. Frames are read left-to-right,
// top-to-bottom across the sheet's rows; segments consume frames in order.
type Seg struct {
	Anim   string   `json:"anim"`
	Frames int      `json:"frames"`
	Out    string   `json:"out"`             // output directory
	Names  []string `json:"names,omitempty"` // explicit file names instead of {anim}_{NN}
	W      int      `json:"w"`
	H      int      `json:"h"`
	Pivot  string   `json:"pivot"`          // "feet" (y=460 on 512) | "center"
	Each   bool     `json:"each,omitempty"` // scale every frame on its own (different objects)
	Poses  string   `json:"-"`
}

// Job is one minigen request plus how to turn its image into game assets.
// One image holds several animations to save the minigen quota (~60-100 images / 5h).
type Job struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Prompt   string `json:"prompt"`
	Mode     string `json:"mode"`
	Out      string `json:"out"` // sprite/tile/raw/wide output file
	W        int    `json:"w"`
	H        int    `json:"h"`
	Pivot    string `json:"pivot"`
	Rows     []int  `json:"rows,omitempty"` // sheet: frames per row
	Segs     []Seg  `json:"segs,omitempty"`
}

// Style lock from ART_PROMPTS §2. Gemini cannot output alpha, so we ask for pure white
// and remove it in post-processing.
const style = "STYLE: 2D mobile game sprite, chibi cartoon caricature, exaggerated comedic facial expressions, " +
	"big head small body (head about 40% of height), thick black outline, flat colors with simple cel shading, vibrant saturated palette, " +
	"clean vector-like look, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, " +
	"plain solid pure white background (#FFFFFF), no text, no watermark, no ground shadow, full body visible."

const styleObject = "STYLE: 2D mobile game asset, cartoon, thick black outline, flat colors with simple cel shading, " +
	"vibrant saturated palette, clean vector-like look, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, " +
	"plain solid pure white background (#FFFFFF), no text, no watermark, no ground shadow."

const styleOpaque = "STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, " +
	"clean vector-like look, no text, no watermark."

const negative = " Avoid: realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, " +
	"watermark, background scenery, grid lines, frame borders, inconsistent outfit."

// sheetPrompt describes a multi-row sheet: which frames go on which row and what each segment shows.
func sheetPrompt(st, subject string, rows []int, segs []Seg, sameCharacter bool) string {
	var b strings.Builder
	total := 0
	for _, n := range rows {
		total += n
	}
	fmt.Fprintf(&b, "%s Sprite sheet of %s with %d rows and %d frames in total. ", st, subject, len(rows), total)
	for i, n := range rows {
		fmt.Fprintf(&b, "Row %d has exactly %d frames. ", i+1, n)
	}
	b.WriteString("Frames are read left-to-right, top-to-bottom: ")
	for _, s := range segs {
		fmt.Fprintf(&b, "[%s, %d frames] %s. ", s.Anim, s.Frames, s.Poses)
	}
	if sameCharacter {
		b.WriteString("Identical character in every frame: same outfit, colors, proportions and size, facing right in 3/4 top-down view, feet aligned on the same baseline in each row. ")
	}
	b.WriteString("Leave plenty of empty white space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers.")
	b.WriteString(negative)
	return b.String()
}

type anim struct {
	name, poses string
	frames      int
}

var charAnims = map[string]anim{
	"idle":         {"idle", "1) standing breathing out 2) breathing in, shoulders raised 3) yawning with mouth wide open 4) scratching head, bored sleepy face", 4},
	"run":          {"run", "cartoon run cycle with spinning legs, arms flailing wildly, mouth wide open screaming AAAA, hair flying; 8 phases: contact, down, pass, up for each leg", 8},
	"chase":        {"chase", "sneaky tiptoe run leaning forward, evil grin with squinting eyes, rubbing both hands together; 8 run phases", 8},
	"flee":         {"flee", "panicked run with both arms raised to the sky, sweat drops flying, bulging eyes, wobbly mouth; 8 run phases", 8},
	"hit":          {"hit", "1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering", 4},
	"levelup":      {"levelup", "1) ready stance 2) golden aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance", 6},
	"breakthrough": {"breakthrough", "1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh", 10},
	"realm_drop":   {"realm_drop", "1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up", 6},
	"death":        {"death", "1) getting smashed 2) smashed harder 3) flattened like rice paper 4) completely flat 5) white ghost soul rising out 6) ghost waving goodbye floating up 7) ghost floating higher 8) ghost almost gone at the top", 8},
	"respawn":      {"respawn", "1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose", 6},
	"teleport_out": {"teleport_out", "1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left", 4},
	"teleport_in":  {"teleport_in", "1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy", 4},
}

// charSheets groups the 12 animations of ART_PROMPTS §4.4 into 5 images per character.
var charSheets = []struct {
	key      string
	rows     []int
	anims    []string
	priority int
}{
	{"core", []int{4, 8}, []string{"idle", "run"}, 1},
	{"combat", []int{4, 8}, []string{"hit", "death"}, 2},
	{"moods", []int{8, 8}, []string{"chase", "flee"}, 3},
	{"events", []int{6, 6, 8}, []string{"levelup", "realm_drop", "teleport_out", "teleport_in"}, 4},
	{"big", []int{5, 5, 6}, []string{"breakthrough", "respawn"}, 4},
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
		for _, cs := range charSheets {
			var segs []Seg
			for _, a := range cs.anims {
				an := charAnims[a]
				segs = append(segs, Seg{Anim: an.name, Frames: an.frames, Out: "characters/" + ch.id + "/frames",
					W: 512, H: 512, Pivot: "feet", Poses: an.poses})
			}
			jobs = append(jobs, Job{Name: "char_" + ch.id + "_" + cs.key, Priority: cs.priority, Mode: modeSheet,
				Rows: cs.rows, Segs: segs, Prompt: sheetPrompt(style, subject, cs.rows, segs, true)})
		}
	}
	return jobs
}

// One image per monster: idle 4 / hurt 2 + death 6 (ART_PROMPTS §6).
func monsterJobs() []Job {
	var jobs []Job
	for _, m := range monsters {
		out := "monsters/" + m.id + "/frames"
		segs := []Seg{
			{Anim: "idle", Frames: 4, Out: out, W: 512, H: 512, Pivot: "feet", Poses: "1) leaning left 2) upright 3) leaning right 4) upright squashed"},
			{Anim: "hurt", Frames: 2, Out: out, W: 512, H: 512, Pivot: "feet", Poses: "1) flinching with angry tears 2) squashed in pain"},
			{Anim: "death", Frames: 6, Out: out, W: 512, H: 512, Pivot: "feet", Poses: "1) inflating 2) cracking 3) bursting 4) exploding into glowing spirit orbs 5) orbs scattering 6) only sparkles left"},
		}
		rows := []int{4, 8}
		jobs = append(jobs, Job{Name: "mon_" + m.id, Priority: 1, Mode: modeSheet, Rows: rows, Segs: segs,
			Prompt: sheetPrompt(style, m.prompt+", 3/4 top-down view facing right", rows, segs, true)})
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

// Swords and item bubbles were generated one per image before the sheet format existed.
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
	sprite := func(name, out string, size int, prompt string) Job {
		return Job{Name: name, Priority: 0, Mode: modeSprite, Out: out, W: size, H: size, Pivot: "center", Prompt: styleObject + " " + prompt + negative}
	}
	jobs := []Job{sprite("item_sword", "items/sword.png", 256, "A glossy transparent soap bubble containing a small glowing sword, shiny highlight.")}
	for _, c := range []struct{ id, color string }{{"orb_small", "green"}, {"orb_medium", "blue"}, {"orb_large", "gold"}} {
		jobs = append(jobs, sprite("item_"+c.id, "items/"+c.id+".png", 128, "A glowing "+c.color+" spirit orb gem, sparkly, round."))
	}
	for _, b := range buffIcons {
		jobs = append(jobs, sprite("buff_"+b.id, "buffs/"+b.id+".png", 256,
			"A glossy transparent soap bubble containing "+b.icon+", tinted "+b.color+", shiny highlight."))
	}
	chest := "an ornate wooden treasure chest wrapped in red ribbons, glowing golden light leaking out, 3/4 top-down view"
	segs := []Seg{
		{Anim: "idle", Frames: 4, Out: "items/chest", W: 512, H: 512, Pivot: "feet", Poses: "1) still 2) rattling left 3) still 4) rattling right"},
		{Anim: "open", Frames: 6, Out: "items/chest", W: 512, H: 512, Pivot: "feet", Poses: "1) closed shaking 2) lid cracking open 3) lid bursting open with light 4) items flying out 5) light fading 6) open empty chest"},
	}
	jobs = append(jobs, Job{Name: "chest", Priority: 2, Mode: modeSheet, Rows: []int{4, 6}, Segs: segs,
		Prompt: sheetPrompt(styleObject, chest, []int{4, 6}, segs, true)})
	return jobs
}

var decorList = []string{"a red plastic stool", "an old bicycle leaning", "a pile of watermelons", "a street food cart", "a banana tree",
	"a rusty water jar", "a broken TV", "a lazy sleeping dog", "a bush with flowers", "a stone lantern"}

func mapJobs() []Job {
	tile := func(id, biome string) Job {
		return Job{Name: "tile_" + id, Priority: 1, Mode: modeTile, Out: "tiles/" + id + ".png", W: 512, H: 512,
			Prompt: styleOpaque + " Seamless tileable top-down ground texture, " + biome + ", square image filling the whole frame, cartoon game style, " +
				"no objects, no characters, even lighting, edges must tile perfectly."}
	}
	jobs := []Job{
		tile("grass", "soft green grass with small clover patches and tiny flowers"),
		tile("concrete", "cracked old concrete street with faded painted lines"),
		tile("dirt", "dry dusty brown dirt with pebbles"),
		tile("mud", "muddy brown puddle texture with bubbles"),
		{Name: "zone_bush", Priority: 1, Mode: modeSprite, Out: "zones/bush.png", W: 512, H: 512, Pivot: "center",
			Prompt: styleObject + " Top-down view of a dense leafy bush cluster, round, lush green, centered." + negative},
	}
	names := make([]string, len(decorList))
	for i, d := range decorList {
		names[i] = decorID(d)
	}
	segs := []Seg{
		{Anim: "decor", Frames: 10, Out: "decor", Names: names, W: 256, H: 256, Pivot: "center", Each: true,
			Poses: "ten different props in this order: " + strings.Join(decorList, ", ")},
		{Anim: "wind", Frames: 4, Out: "zones", W: 256, H: 256, Pivot: "center",
			Poses: "a white cartoon wind gust arrow with swirl lines and light blue outline pointing right, sliding right a little more each frame"},
		{Anim: "wall", Frames: 1, Out: "decor", Names: []string{"wall_border"}, W: 1024, H: 256, Pivot: "center", Each: true,
			Poses: "one wide seamless horizontal stone and bamboo fence segment, tileable left-right"},
	}
	rows := []int{5, 5, 4, 1}
	jobs = append(jobs, Job{Name: "decor_sheet", Priority: 3, Mode: modeSheet, Rows: rows, Segs: segs,
		Prompt: sheetPrompt(styleObject, "top-down 3/4 view decoration props", rows, segs, false)})
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
		dir := "obstacles/" + o.id
		subject := "a top-down 3/4 view game obstacle, " + o.desc
		states := []Seg{{Anim: "states", Frames: 4, Out: dir, Names: []string{"intact", "cracked", "nearly_broken", "rubble"},
			W: o.w, H: o.h, Pivot: "center",
			Poses: "1) intact and solid 2) with visible cracks 3) heavily damaged, crumbling, pieces missing 4) flat pile of rubble and debris on the ground"}}
		jobs = append(jobs, Job{Name: "obs_" + o.id + "_states", Priority: 1, Mode: modeSheet, Rows: []int{2, 2}, Segs: states,
			Prompt: sheetPrompt(styleObject, subject, []int{2, 2}, states, true)})
		if o.id == "stone_wall" {
			continue
		}
		anims := []Seg{
			{Anim: "break", Frames: 6, Out: dir + "/break", W: o.w, H: o.h, Pivot: "center", Poses: "1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile"},
			{Anim: "regrow", Frames: 6, Out: dir + "/regrow", W: o.w, H: o.h, Pivot: "center", Poses: "1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid"},
		}
		jobs = append(jobs, Job{Name: "obs_" + o.id + "_anims", Priority: 4, Mode: modeSheet, Rows: []int{6, 6}, Segs: anims,
			Prompt: sheetPrompt(styleObject, subject, []int{6, 6}, anims, true)})
	}
	return jobs
}

// One image per gate colour: idle 6 / warn 4 / burst 6 (ART_PROMPTS §9.3).
func gateJobs() []Job {
	var jobs []Job
	for _, color := range []string{"purple", "blue", "orange"} {
		out := "gates/" + color
		segs := []Seg{
			{Anim: "idle", Frames: 6, Out: out, W: 512, H: 512, Pivot: "center", Poses: "slowly swirling vortex, rotating a little more each frame"},
			{Anim: "warn", Frames: 4, Out: out, W: 512, H: 512, Pivot: "center", Poses: "flashing brighter and brighter, intense glow"},
			{Anim: "burst", Frames: 6, Out: out, W: 512, H: 512, Pivot: "center", Poses: "bursting with light rays expanding then fading"},
		}
		rows := []int{6, 4, 6}
		jobs = append(jobs, Job{Name: "gate_" + color, Priority: 1, Mode: modeSheet, Rows: rows, Segs: segs,
			Prompt: sheetPrompt(styleObject, "a top-down magical teleport portal on the ground, swirling "+color+
				" vortex inside a carved stone ring with taoist symbols, glowing", rows, segs, true)})
	}
	return jobs
}

func vfxSeg(id string, n int, poses string) Seg {
	return Seg{Anim: id, Frames: n, Out: "vfx/" + id, W: 256, H: 256, Pivot: "center", Poses: poses}
}

func vfxJobs() []Job {
	a := []Seg{
		vfxSeg("clash_spark", 6, "comic sword clash spark burst, white and yellow star shapes, expanding then fading"),
		vfxSeg("sword_shatter", 6, "a cartoon sword shattering into sharp shards flying outward"),
		vfxSeg("death_smoke", 6, "comic white-grey smoke puff cloud with dark outline, expanding then dissipating"),
	}
	b := []Seg{
		vfxSeg("levelup_aura", 6, "rising golden aura ring with sparkles"),
		vfxSeg("lightning", 6, "thick cartoon lightning bolt striking down from the top"),
		vfxSeg("shield_break", 6, "golden bubble shield cracking and bursting into glass shards"),
		vfxSeg("stun_stars", 4, "ring of small yellow cartoon stars orbiting"),
		vfxSeg("run_dust", 4, "small brown dust puffs behind running feet"),
	}
	return []Job{
		{Name: "vfx_a", Priority: 2, Mode: modeSheet, Rows: []int{6, 6, 6}, Segs: a,
			Prompt: sheetPrompt(styleObject, "game visual effects", []int{6, 6, 6}, a, false)},
		{Name: "vfx_b", Priority: 3, Mode: modeSheet, Rows: []int{6, 6, 6, 8}, Segs: b,
			Prompt: sheetPrompt(styleObject, "game visual effects", []int{6, 6, 6, 8}, b, false)},
		{Name: "vfx_crack", Priority: 3, Mode: modeSprite, Out: "vfx/crack.png", W: 256, H: 256, Pivot: "center",
			Prompt: styleObject + " Thin dark grey crack lines overlay pattern like broken glass, sparse, centered." + negative},
	}
}

var badgeThemes = []string{"bamboo green", "kitchen steel grey", "peach wood red", "polished silver", "flame orange",
	"ice crystal blue", "lightning yellow", "cursed dark purple", "celestial jade and gold"}

func uiJobs() []Job {
	badgeNames := make([]string, len(badgeThemes))
	for i := range badgeThemes {
		badgeNames[i] = fmt.Sprintf("badge_%d", i)
	}
	segs := []Seg{
		{Anim: "badge", Frames: 9, Out: "ui", Names: badgeNames, W: 256, H: 256, Pivot: "center", Each: true,
			Poses: "nine round game badge emblems for cultivation ranks, each with a small sword symbol and no letters, themes in order: " + strings.Join(badgeThemes, ", ")},
		{Anim: "medal", Frames: 3, Out: "ui", Names: []string{"medal_gold", "medal_silver", "medal_bronze"}, W: 128, H: 128, Pivot: "center", Each: true,
			Poses: "a gold medal, a silver medal and a bronze medal, each with a tiny sword emblem and no letters"},
	}
	rows := []int{5, 4, 3}
	return []Job{
		{Name: "ui_badges", Priority: 2, Mode: modeSheet, Rows: rows, Segs: segs,
			Prompt: sheetPrompt(styleObject, "game UI icons", rows, segs, false)},
		{Name: "ui_speech", Priority: 2, Mode: modeSprite, Out: "ui/speech_bubble.png", W: 512, H: 256, Pivot: "center",
			Prompt: styleObject + " An empty white comic speech bubble with thick black outline, wide oval, tail at bottom left, empty inside, centered." + negative},
		{Name: "ui_scroll", Priority: 3, Mode: modeSprite, Out: "ui/scroll.png", W: 1024, H: 768, Pivot: "center",
			Prompt: styleObject + " An empty unrolled ancient paper scroll with wooden handles on left and right, horizontal, nothing written, centered." + negative},
		{Name: "ui_home_bg", Priority: 2, Mode: modeWide, Out: "ui/home_bg.png", W: 1920, H: 1080,
			Prompt: styleOpaque + " Wide 16:9 illustration of a chaotic cartoon Vietnamese village square with many chibi characters " +
				"spinning flying swords around them, sunny, bamboo and street stalls, no text."},
	}
}

func allJobs() []Job {
	var jobs []Job
	for _, f := range []func() []Job{swordJobs, itemJobs, characterJobs, monsterJobs, mapJobs, obstacleJobs, gateJobs, vfxJobs, uiJobs} {
		jobs = append(jobs, f()...)
	}
	return jobs
}
