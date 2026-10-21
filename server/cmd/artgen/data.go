package main

// Character and monster descriptions from docs/ART_PROMPTS.md §5–6 ("bựa" version).
// gags replace or extend the generic pose of an animation (ART_PROMPTS §5: "gag riêng cho từng anim").

type character struct {
	id, desc, palette string
	gags              map[string]string
}

var characters = []character{
	{"a_hon", "a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops",
		"#FFFFFF #E63946 #F4A261 #2B2D42", map[string]string{
			"idle":  "picking his nose, staring at his finger, flicking the booger away",
			"chase": "licking his lips with his tongue out, one eyebrow raised, rubbing his hands",
			"hit":   "his chipped buck tooth flies out of his mouth",
			"death": "squashed flat, the steamed bun hat gets another bite taken out of it",
		}},
	{"ba_tam_loa", "a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots",
		"#2A9D8F #E9C46A #F28482 #264653", map[string]string{
			"run":   "screaming into the megaphone while running, spit spraying out in streams",
			"chase": "puckering her smeared red lips into a threatening air kiss",
			"flee":  "the megaphone stuck upside down over her face, running blind",
			"hit":   "her lipstick mark gets printed on the air in front of her",
		}},
	{"ong_ba_vong", "a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband",
		"#8D99AE #EDF2F4 #6D597A #B56576", map[string]string{
			"idle":         "dozing off while standing, a snot bubble inflating and deflating from his nose",
			"flee":         "pajama pants fallen down to his knees, pulling them up while running",
			"hit":          "his three strands of hair fly up and stand straight",
			"breakthrough": "for one frame his hair grows back thick and lush, then it all falls out again",
		}},
	{"dai_ca_dep", "a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks",
		"#7209B7 #FFD60A #000000 #F72585", map[string]string{
			"idle":       "admiring his gold tooth in a tiny hand mirror",
			"chase":      "snapping his fingers to call henchmen who do not exist, sneering",
			"hit":        "his sunglasses fall off revealing cross-eyes",
			"realm_drop": "all the buttons of his suit pop off",
		}},
	{"map_tra_sua", "a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers",
		"#CDB4DB #A47148 #FFC8DD #3D405B", map[string]string{
			"idle":  "burping out a big bubble",
			"chase": "slurping bubble tea loudly through the straw with sly squinting eyes",
			"hit":   "spitting tapioca pearls out like a machine gun",
			"death": "squashed flat, the bubble tea cup spills into a puddle",
		}},
	{"co_ba_ganh", "a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers",
		"#BC6C25 #DDA15E #606C38 #FEFAE0", map[string]string{
			"chase": "spinning the shoulder pole overhead like helicopter blades",
			"flee":  "fruit flying everywhere and falling on her own head",
			"hit":   "both braids curl up into springs",
		}},
	{"su_phu_mi", "an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon",
		"#FB8500 #FFB703 #8ECAE6 #023047", map[string]string{
			"idle":  "picking up one strand of his own noodle beard with the chopsticks and eating it",
			"chase": "twirling the chopsticks, cross-eyes looking in two directions",
			"hit":   "his noodle beard gets cut and noodles fly everywhere",
		}},
	{"be_ti", "a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing",
		"#E76F51 #F4A261 #2A9D8F #1D3557", map[string]string{
			"idle":  "kissing his own biceps",
			"chase": "flexing so hard his face turns red with forehead veins popping",
			"hit":   "his muscles deflate like balloons and then pump back up",
		}},
	{"shipper", "a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand",
		"#FFBE0B #FB5607 #3A86FF #222222", map[string]string{
			"chase": "holding up his phone with a 'come down and get your order' face",
			"flee":  "the delivery box lid pops open and food flies out",
			"hit":   "his phone flies away and the screen cracks",
		}},
	{"cu_rua", "a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops",
		"#52B788 #D8F3DC #6F4E37 #F1FAEE", map[string]string{
			"run":  "legs spinning super fast but his body barely moves forward",
			"flee": "pulling his head, arms and legs into the turtle shell, the shell slides away on its own",
			"hit":  "his long eyebrows wrap around his head like a turban",
		}},
}

var monsters = []struct{ id, prompt string }{
	{"banh_bao_cau", "an angry living steamed bun creature with tiny legs, steam puffing from head, round belly area left blank in the center"},
	{"dep_lao", "a sentient flip-flop sandal creature hopping, googly eyes, sticking tongue out, blank round circle on its sole"},
	{"noi_com", "a rice cooker monster with its lid as a biting mouth full of rice teeth, blank round panel on front"},
	{"ga_hoang", "a panicking chubby chicken monster, feathers flying, bulging eyes, blank round badge on chest"},
	{"muoi", "a giant fat mosquito monster with a straw-like nose, tiny wings, smug face, blank round area on belly"},
	{"vua_ghe", "a towering boss made of stacked red plastic stools, crown of a plastic basin, angry eyes, blank big circle in the middle"},
}
