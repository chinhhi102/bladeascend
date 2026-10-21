package main

// Character and monster descriptions from docs/ART_PROMPTS.md §5–6.

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

var monsters = []struct{ id, prompt string }{
	{"banh_bao_cau", "an angry living steamed bun creature with tiny legs, steam puffing from head, round belly area left blank in the center"},
	{"dep_lao", "a sentient flip-flop sandal creature hopping, googly eyes, sticking tongue out, blank round circle on its sole"},
	{"noi_com", "a rice cooker monster with its lid as a biting mouth full of rice teeth, blank round panel on front"},
	{"ga_hoang", "a panicking chubby chicken monster, feathers flying, bulging eyes, blank round badge on chest"},
	{"muoi", "a giant fat mosquito monster with a straw-like nose, tiny wings, smug face, blank round area on belly"},
	{"vua_ghe", "a towering boss made of stacked red plastic stools, crown of a plastic basin, angry eyes, blank big circle in the middle"},
}
