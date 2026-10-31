# Tình trạng art

Tạo lại file này: `cd server && go run ./cmd/artgen status`.
Asset thiếu → game tự vẽ placeholder (hình tròn + chữ). Sinh thêm: `go run ./cmd/artgen run -p 2 -n 30` rồi `go run ./cmd/artgen prep`.
Nếu tự tải ảnh từ Gemini: lưu thành `assets/_raw/<tên job>.png` rồi chạy `prep`.

## Còn thiếu (1 ảnh)

| Job | Ưu tiên | Ra file |
|---|---|---|
| `decor_nature` | p3 | decor (decor ×7) |

## Đã có (347 ảnh)

| Job | Ưu tiên | Ra file |
|---|---|---|
| `sword_realm_0` | p0 | swords/realm_0.png |
| `sword_realm_1` | p0 | swords/realm_1.png |
| `sword_realm_2` | p0 | swords/realm_2.png |
| `sword_realm_3` | p0 | swords/realm_3.png |
| `sword_realm_4` | p0 | swords/realm_4.png |
| `sword_realm_5` | p0 | swords/realm_5.png |
| `sword_realm_6` | p0 | swords/realm_6.png |
| `sword_realm_7` | p0 | swords/realm_7.png |
| `sword_realm_8` | p0 | swords/realm_8.png |
| `item_sword` | p0 | items/sword.png |
| `item_bundle` | p0 | items/bundle.png |
| `item_orb_small` | p0 | items/orb_small.png |
| `item_orb_medium` | p0 | items/orb_medium.png |
| `item_orb_large` | p0 | items/orb_large.png |
| `buff_spin` | p0 | buffs/spin.png |
| `buff_haste` | p0 | buffs/haste.png |
| `buff_magnet` | p0 | buffs/magnet.png |
| `buff_hard` | p0 | buffs/hard.png |
| `buff_shield` | p0 | buffs/shield.png |
| `buff_exp2` | p0 | buffs/exp2.png |
| `buff_stealth` | p0 | buffs/stealth.png |
| `char_a_hon_idle` | p1 | characters/a_hon/frames (idle ×4) |
| `char_a_hon_run` | p1 | characters/a_hon/frames (run ×8) |
| `char_ba_tam_loa_idle` | p1 | characters/ba_tam_loa/frames (idle ×4) |
| `char_ba_tam_loa_run` | p1 | characters/ba_tam_loa/frames (run ×8) |
| `char_ong_ba_vong_idle` | p1 | characters/ong_ba_vong/frames (idle ×4) |
| `char_ong_ba_vong_run` | p1 | characters/ong_ba_vong/frames (run ×8) |
| `char_dai_ca_dep_idle` | p1 | characters/dai_ca_dep/frames (idle ×4) |
| `char_dai_ca_dep_run` | p1 | characters/dai_ca_dep/frames (run ×8) |
| `char_map_tra_sua_idle` | p1 | characters/map_tra_sua/frames (idle ×4) |
| `char_map_tra_sua_run` | p1 | characters/map_tra_sua/frames (run ×8) |
| `char_co_ba_ganh_idle` | p1 | characters/co_ba_ganh/frames (idle ×4) |
| `char_co_ba_ganh_run` | p1 | characters/co_ba_ganh/frames (run ×8) |
| `char_su_phu_mi_idle` | p1 | characters/su_phu_mi/frames (idle ×4) |
| `char_su_phu_mi_run` | p1 | characters/su_phu_mi/frames (run ×8) |
| `char_be_ti_idle` | p1 | characters/be_ti/frames (idle ×4) |
| `char_be_ti_run` | p1 | characters/be_ti/frames (run ×8) |
| `char_shipper_idle` | p1 | characters/shipper/frames (idle ×4) |
| `char_shipper_run` | p1 | characters/shipper/frames (run ×8) |
| `char_cu_rua_idle` | p1 | characters/cu_rua/frames (idle ×4) |
| `char_cu_rua_run` | p1 | characters/cu_rua/frames (run ×8) |
| `mon_banh_bao_cau` | p1 | monsters/banh_bao_cau/frames (idle ×4), monsters/banh_bao_cau/frames (hurt ×2), monsters/banh_bao_cau/frames (death ×6) |
| `mon_dep_lao` | p1 | monsters/dep_lao/frames (idle ×4), monsters/dep_lao/frames (hurt ×2), monsters/dep_lao/frames (death ×6) |
| `mon_noi_com` | p1 | monsters/noi_com/frames (idle ×4), monsters/noi_com/frames (hurt ×2), monsters/noi_com/frames (death ×6) |
| `mon_ga_hoang` | p1 | monsters/ga_hoang/frames (idle ×4), monsters/ga_hoang/frames (hurt ×2), monsters/ga_hoang/frames (death ×6) |
| `mon_muoi` | p1 | monsters/muoi/frames (idle ×4), monsters/muoi/frames (hurt ×2), monsters/muoi/frames (death ×6) |
| `mon_vua_ghe` | p1 | monsters/vua_ghe/frames (idle ×4), monsters/vua_ghe/frames (hurt ×2), monsters/vua_ghe/frames (death ×6) |
| `tile_grass` | p1 | tiles/grass.png |
| `tile_concrete` | p1 | tiles/concrete.png |
| `tile_dirt` | p1 | tiles/dirt.png |
| `tile_mud` | p1 | tiles/mud.png |
| `zone_bush` | p1 | zones/bush.png |
| `obs_fence_states` | p1 | obstacles/fence (states ×4) |
| `obs_fence_states_vertical` | p1 | obstacles/fence (states ×4) |
| `obs_stall_states` | p1 | obstacles/stall (states ×4) |
| `obs_stall_states_vertical` | p1 | obstacles/stall (states ×4) |
| `obs_cart_states` | p1 | obstacles/cart (states ×4) |
| `obs_cart_states_vertical` | p1 | obstacles/cart (states ×4) |
| `obs_brick_states` | p1 | obstacles/brick (states ×4) |
| `obs_brick_states_vertical` | p1 | obstacles/brick (states ×4) |
| `obs_boulder_states` | p1 | obstacles/boulder (states ×4) |
| `obs_boulder_states_vertical` | p1 | obstacles/boulder (states ×4) |
| `obs_stone_wall_states` | p1 | obstacles/stone_wall (states ×4) |
| `obs_stone_wall_states_vertical` | p1 | obstacles/stone_wall (states ×4) |
| `obs_bamboo_wall_states` | p1 | obstacles/bamboo_wall (states ×4) |
| `obs_bamboo_wall_states_vertical` | p1 | obstacles/bamboo_wall (states ×4) |
| `gate_purple` | p1 | gates/purple (idle ×6), gates/purple (warn ×4), gates/purple (burst ×6) |
| `gate_blue` | p1 | gates/blue (idle ×6), gates/blue (warn ×4), gates/blue (burst ×6) |
| `gate_orange` | p1 | gates/orange (idle ×6), gates/orange (warn ×4), gates/orange (burst ×6) |
| `item_skill_scroll` | p2 | items/skill_scroll.png |
| `chest` | p2 | items/chest (idle ×4), items/chest (open ×6) |
| `char_a_hon_hit` | p2 | characters/a_hon/frames (hit ×4) |
| `char_a_hon_death` | p2 | characters/a_hon/frames (death ×8) |
| `char_ba_tam_loa_hit` | p2 | characters/ba_tam_loa/frames (hit ×4) |
| `char_ba_tam_loa_death` | p2 | characters/ba_tam_loa/frames (death ×8) |
| `char_ong_ba_vong_hit` | p2 | characters/ong_ba_vong/frames (hit ×4) |
| `char_ong_ba_vong_death` | p2 | characters/ong_ba_vong/frames (death ×8) |
| `char_dai_ca_dep_hit` | p2 | characters/dai_ca_dep/frames (hit ×4) |
| `char_dai_ca_dep_death` | p2 | characters/dai_ca_dep/frames (death ×8) |
| `char_map_tra_sua_hit` | p2 | characters/map_tra_sua/frames (hit ×4) |
| `char_map_tra_sua_death` | p2 | characters/map_tra_sua/frames (death ×8) |
| `char_co_ba_ganh_hit` | p2 | characters/co_ba_ganh/frames (hit ×4) |
| `char_co_ba_ganh_death` | p2 | characters/co_ba_ganh/frames (death ×8) |
| `char_su_phu_mi_hit` | p2 | characters/su_phu_mi/frames (hit ×4) |
| `char_su_phu_mi_death` | p2 | characters/su_phu_mi/frames (death ×8) |
| `char_be_ti_hit` | p2 | characters/be_ti/frames (hit ×4) |
| `char_be_ti_death` | p2 | characters/be_ti/frames (death ×8) |
| `char_shipper_hit` | p2 | characters/shipper/frames (hit ×4) |
| `char_shipper_death` | p2 | characters/shipper/frames (death ×8) |
| `char_cu_rua_hit` | p2 | characters/cu_rua/frames (hit ×4) |
| `char_cu_rua_death` | p2 | characters/cu_rua/frames (death ×8) |
| `vfx_clash_spark` | p2 | vfx/clash_spark (clash_spark ×6) |
| `vfx_sword_shatter` | p2 | vfx/sword_shatter (sword_shatter ×6) |
| `vfx_death_smoke` | p2 | vfx/death_smoke (death_smoke ×6) |
| `tile_rocky_ground` | p2 | tiles/rocky_ground.png |
| `tile_fine_gravel` | p2 | tiles/fine_gravel.png |
| `tile_mossy_stone` | p2 | tiles/mossy_stone.png |
| `tile_soft_grass` | p2 | tiles/soft_grass.png |
| `tile_bamboo_litter` | p2 | tiles/bamboo_litter.png |
| `tile_damp_earth` | p2 | tiles/damp_earth.png |
| `tile_dirt_path` | p2 | tiles/dirt_path.png |
| `tile_village_grass` | p2 | tiles/village_grass.png |
| `tile_worn_cobble` | p2 | tiles/worn_cobble.png |
| `tile_dusty_earth` | p2 | tiles/dusty_earth.png |
| `ui_badges` | p2 | ui (badge ×9), ui (medal ×3) |
| `ui_speech` | p2 | ui/speech_bubble.png |
| `ui_home_bg` | p2 | ui/home_bg.png |
| `icon_skill_scroll` | p3 | ui/icon_skill_scroll.png |
| `icon_skill_a_hon` | p3 | ui/skills/a_hon.png |
| `icon_skill_ba_tam_loa` | p3 | ui/skills/ba_tam_loa.png |
| `icon_skill_ong_ba_vong` | p3 | ui/skills/ong_ba_vong.png |
| `icon_skill_dai_ca_dep` | p3 | ui/skills/dai_ca_dep.png |
| `icon_skill_map_tra_sua` | p3 | ui/skills/map_tra_sua.png |
| `icon_skill_co_ba_ganh` | p3 | ui/skills/co_ba_ganh.png |
| `icon_skill_su_phu_mi` | p3 | ui/skills/su_phu_mi.png |
| `icon_skill_be_ti` | p3 | ui/skills/be_ti.png |
| `icon_skill_shipper` | p3 | ui/skills/shipper.png |
| `icon_skill_cu_rua` | p3 | ui/skills/cu_rua.png |
| `char_dan_em_run` | p3 | characters/dan_em/frames (run ×8) |
| `char_a_hon_chase` | p3 | characters/a_hon/frames (chase ×8) |
| `char_a_hon_flee` | p3 | characters/a_hon/frames (flee ×8) |
| `char_a_hon_skill` | p3 | characters/a_hon/frames (skill ×6) |
| `char_ba_tam_loa_chase` | p3 | characters/ba_tam_loa/frames (chase ×8) |
| `char_ba_tam_loa_flee` | p3 | characters/ba_tam_loa/frames (flee ×8) |
| `char_ba_tam_loa_skill` | p3 | characters/ba_tam_loa/frames (skill ×6) |
| `char_ong_ba_vong_chase` | p3 | characters/ong_ba_vong/frames (chase ×8) |
| `char_ong_ba_vong_flee` | p3 | characters/ong_ba_vong/frames (flee ×8) |
| `char_ong_ba_vong_skill` | p3 | characters/ong_ba_vong/frames (skill ×6) |
| `char_dai_ca_dep_chase` | p3 | characters/dai_ca_dep/frames (chase ×8) |
| `char_dai_ca_dep_flee` | p3 | characters/dai_ca_dep/frames (flee ×8) |
| `char_dai_ca_dep_skill` | p3 | characters/dai_ca_dep/frames (skill ×6) |
| `char_map_tra_sua_chase` | p3 | characters/map_tra_sua/frames (chase ×8) |
| `char_map_tra_sua_flee` | p3 | characters/map_tra_sua/frames (flee ×8) |
| `char_map_tra_sua_skill` | p3 | characters/map_tra_sua/frames (skill ×6) |
| `char_co_ba_ganh_chase` | p3 | characters/co_ba_ganh/frames (chase ×8) |
| `char_co_ba_ganh_flee` | p3 | characters/co_ba_ganh/frames (flee ×8) |
| `char_co_ba_ganh_skill` | p3 | characters/co_ba_ganh/frames (skill ×6) |
| `char_su_phu_mi_chase` | p3 | characters/su_phu_mi/frames (chase ×8) |
| `char_su_phu_mi_flee` | p3 | characters/su_phu_mi/frames (flee ×8) |
| `char_su_phu_mi_skill` | p3 | characters/su_phu_mi/frames (skill ×6) |
| `char_be_ti_chase` | p3 | characters/be_ti/frames (chase ×8) |
| `char_be_ti_flee` | p3 | characters/be_ti/frames (flee ×8) |
| `char_be_ti_skill` | p3 | characters/be_ti/frames (skill ×6) |
| `char_shipper_chase` | p3 | characters/shipper/frames (chase ×8) |
| `char_shipper_flee` | p3 | characters/shipper/frames (flee ×8) |
| `char_shipper_skill` | p3 | characters/shipper/frames (skill ×6) |
| `char_cu_rua_chase` | p3 | characters/cu_rua/frames (chase ×8) |
| `char_cu_rua_flee` | p3 | characters/cu_rua/frames (flee ×8) |
| `char_cu_rua_skill` | p3 | characters/cu_rua/frames (skill ×6) |
| `decor_sheet` | p3 | decor (decor ×10), decor (wall ×1) |
| `zone_wind` | p3 | zones (wind ×4) |
| `vfx_levelup_aura` | p3 | vfx/levelup_aura (levelup_aura ×6) |
| `vfx_lightning` | p3 | vfx/lightning (lightning ×6) |
| `vfx_shield_break` | p3 | vfx/shield_break (shield_break ×6) |
| `vfx_stun_stars` | p3 | vfx/stun_stars (stun_stars ×4) |
| `vfx_crack` | p3 | vfx/crack.png |
| `beast_dragon` | p3 | beasts (parts ×3) |
| `beast_crane` | p3 | beasts (parts ×3) |
| `beast_phoenix` | p3 | beasts (parts ×3) |
| `decal_ground_cracks` | p3 | decals/ground_cracks.png |
| `decal_grass_tuft` | p3 | decals/grass_tuft.png |
| `decal_fallen_leaves` | p3 | decals/fallen_leaves.png |
| `decal_pebbles` | p3 | decals/pebbles.png |
| `decal_puddle` | p3 | decals/puddle.png |
| `decal_tree_roots` | p3 | decals/tree_roots.png |
| `decal_wildflowers` | p3 | decals/wildflowers.png |
| `decal_roof_tiles` | p3 | decals/roof_tiles.png |
| `decal_brick_rubble` | p3 | decals/brick_rubble.png |
| `decal_bamboo_leaves` | p3 | decals/bamboo_leaves.png |
| `decal_moss_patch` | p3 | decals/moss_patch.png |
| `decal_footprints` | p3 | decals/footprints.png |
| `decal_cart_tracks` | p3 | decals/cart_tracks.png |
| `plate_ribbon_top1` | p3 | ui/ribbon_top1.png |
| `plate_ribbon_top2` | p3 | ui/ribbon_top2.png |
| `plate_ribbon_top3` | p3 | ui/ribbon_top3.png |
| `plate_crown_small` | p3 | ui/crown_small.png |
| `plate_realm_pill` | p3 | ui/realm_pill.png |
| `plate_shine_streak` | p3 | ui/shine_streak.png |
| `ui_scroll` | p3 | ui/scroll.png |
| `vfx_boba_1` | p4 | vfx/boba_1.png |
| `vfx_boba_2` | p4 | vfx/boba_2.png |
| `vfx_boba_3` | p4 | vfx/boba_3.png |
| `vfx_boba_4` | p4 | vfx/boba_4.png |
| `vfx_boba_5` | p4 | vfx/boba_5.png |
| `vfx_boba_6` | p4 | vfx/boba_6.png |
| `vfx_boba_drip` | p4 | vfx/boba_drip.png |
| `zone_cliff_wall` | p4 | obstacles/cliff_wall.png |
| `zone_rope_bridge` | p4 | obstacles/rope_bridge.png |
| `zone_bamboo_wall` | p4 | obstacles/bamboo_wall.png |
| `zone_bamboo_corner` | p4 | obstacles/bamboo_corner.png |
| `zone_village_roof` | p4 | obstacles/village_roof.png |
| `zone_lantern_string` | p4 | obstacles/lantern_string.png |
| `zone_ice_island_edge` | p4 | obstacles/ice_island_edge.png |
| `tile_mountain_rock` | p4 | tiles/mountain_rock.png |
| `tile_bamboo_floor` | p4 | tiles/bamboo_floor.png |
| `tile_cobblestone` | p4 | tiles/cobblestone.png |
| `tile_ice` | p4 | tiles/ice.png |
| `tile_snow` | p4 | tiles/snow.png |
| `char_a_hon_levelup` | p4 | characters/a_hon/frames (levelup ×6) |
| `char_a_hon_realm_drop` | p4 | characters/a_hon/frames (realm_drop ×6) |
| `char_a_hon_breakthrough` | p4 | characters/a_hon/frames (breakthrough ×10) |
| `char_a_hon_respawn` | p4 | characters/a_hon/frames (respawn ×6) |
| `char_a_hon_teleport_out` | p4 | characters/a_hon/frames (teleport_out ×4) |
| `char_a_hon_teleport_in` | p4 | characters/a_hon/frames (teleport_in ×4) |
| `char_ba_tam_loa_levelup` | p4 | characters/ba_tam_loa/frames (levelup ×6) |
| `char_ba_tam_loa_realm_drop` | p4 | characters/ba_tam_loa/frames (realm_drop ×6) |
| `char_ba_tam_loa_breakthrough` | p4 | characters/ba_tam_loa/frames (breakthrough ×10) |
| `char_ba_tam_loa_respawn` | p4 | characters/ba_tam_loa/frames (respawn ×6) |
| `char_ba_tam_loa_teleport_out` | p4 | characters/ba_tam_loa/frames (teleport_out ×4) |
| `char_ba_tam_loa_teleport_in` | p4 | characters/ba_tam_loa/frames (teleport_in ×4) |
| `char_ong_ba_vong_levelup` | p4 | characters/ong_ba_vong/frames (levelup ×6) |
| `char_ong_ba_vong_realm_drop` | p4 | characters/ong_ba_vong/frames (realm_drop ×6) |
| `char_ong_ba_vong_breakthrough` | p4 | characters/ong_ba_vong/frames (breakthrough ×10) |
| `char_ong_ba_vong_respawn` | p4 | characters/ong_ba_vong/frames (respawn ×6) |
| `char_ong_ba_vong_teleport_out` | p4 | characters/ong_ba_vong/frames (teleport_out ×4) |
| `char_ong_ba_vong_teleport_in` | p4 | characters/ong_ba_vong/frames (teleport_in ×4) |
| `char_dai_ca_dep_levelup` | p4 | characters/dai_ca_dep/frames (levelup ×6) |
| `char_dai_ca_dep_realm_drop` | p4 | characters/dai_ca_dep/frames (realm_drop ×6) |
| `char_dai_ca_dep_breakthrough` | p4 | characters/dai_ca_dep/frames (breakthrough ×10) |
| `char_dai_ca_dep_respawn` | p4 | characters/dai_ca_dep/frames (respawn ×6) |
| `char_dai_ca_dep_teleport_out` | p4 | characters/dai_ca_dep/frames (teleport_out ×4) |
| `char_dai_ca_dep_teleport_in` | p4 | characters/dai_ca_dep/frames (teleport_in ×4) |
| `char_map_tra_sua_levelup` | p4 | characters/map_tra_sua/frames (levelup ×6) |
| `char_map_tra_sua_realm_drop` | p4 | characters/map_tra_sua/frames (realm_drop ×6) |
| `char_map_tra_sua_breakthrough` | p4 | characters/map_tra_sua/frames (breakthrough ×10) |
| `char_map_tra_sua_respawn` | p4 | characters/map_tra_sua/frames (respawn ×6) |
| `char_map_tra_sua_teleport_out` | p4 | characters/map_tra_sua/frames (teleport_out ×4) |
| `char_map_tra_sua_teleport_in` | p4 | characters/map_tra_sua/frames (teleport_in ×4) |
| `char_co_ba_ganh_levelup` | p4 | characters/co_ba_ganh/frames (levelup ×6) |
| `char_co_ba_ganh_realm_drop` | p4 | characters/co_ba_ganh/frames (realm_drop ×6) |
| `char_co_ba_ganh_breakthrough` | p4 | characters/co_ba_ganh/frames (breakthrough ×10) |
| `char_co_ba_ganh_respawn` | p4 | characters/co_ba_ganh/frames (respawn ×6) |
| `char_co_ba_ganh_teleport_out` | p4 | characters/co_ba_ganh/frames (teleport_out ×4) |
| `char_co_ba_ganh_teleport_in` | p4 | characters/co_ba_ganh/frames (teleport_in ×4) |
| `char_su_phu_mi_levelup` | p4 | characters/su_phu_mi/frames (levelup ×6) |
| `char_su_phu_mi_realm_drop` | p4 | characters/su_phu_mi/frames (realm_drop ×6) |
| `char_su_phu_mi_breakthrough` | p4 | characters/su_phu_mi/frames (breakthrough ×10) |
| `char_su_phu_mi_respawn` | p4 | characters/su_phu_mi/frames (respawn ×6) |
| `char_su_phu_mi_teleport_out` | p4 | characters/su_phu_mi/frames (teleport_out ×4) |
| `char_su_phu_mi_teleport_in` | p4 | characters/su_phu_mi/frames (teleport_in ×4) |
| `char_be_ti_levelup` | p4 | characters/be_ti/frames (levelup ×6) |
| `char_be_ti_realm_drop` | p4 | characters/be_ti/frames (realm_drop ×6) |
| `char_be_ti_breakthrough` | p4 | characters/be_ti/frames (breakthrough ×10) |
| `char_be_ti_respawn` | p4 | characters/be_ti/frames (respawn ×6) |
| `char_be_ti_teleport_out` | p4 | characters/be_ti/frames (teleport_out ×4) |
| `char_be_ti_teleport_in` | p4 | characters/be_ti/frames (teleport_in ×4) |
| `char_shipper_levelup` | p4 | characters/shipper/frames (levelup ×6) |
| `char_shipper_realm_drop` | p4 | characters/shipper/frames (realm_drop ×6) |
| `char_shipper_breakthrough` | p4 | characters/shipper/frames (breakthrough ×10) |
| `char_shipper_respawn` | p4 | characters/shipper/frames (respawn ×6) |
| `char_shipper_teleport_out` | p4 | characters/shipper/frames (teleport_out ×4) |
| `char_shipper_teleport_in` | p4 | characters/shipper/frames (teleport_in ×4) |
| `char_cu_rua_levelup` | p4 | characters/cu_rua/frames (levelup ×6) |
| `char_cu_rua_realm_drop` | p4 | characters/cu_rua/frames (realm_drop ×6) |
| `char_cu_rua_breakthrough` | p4 | characters/cu_rua/frames (breakthrough ×10) |
| `char_cu_rua_respawn` | p4 | characters/cu_rua/frames (respawn ×6) |
| `char_cu_rua_teleport_out` | p4 | characters/cu_rua/frames (teleport_out ×4) |
| `char_cu_rua_teleport_in` | p4 | characters/cu_rua/frames (teleport_in ×4) |
| `bodyA_a_hon` | p4 | characters/a_hon/body (run ×6), characters/a_hon/body (idle ×4) |
| `facesA_a_hon` | p4 | characters/a_hon/faces (face ×8) |
| `bodyA_dai_ca_dep` | p4 | characters/dai_ca_dep/body (run ×6), characters/dai_ca_dep/body (idle ×4) |
| `facesA_dai_ca_dep` | p4 | characters/dai_ca_dep/faces (face ×8) |
| `bodyA_su_phu_mi` | p4 | characters/su_phu_mi/body (run ×6), characters/su_phu_mi/body (idle ×4) |
| `facesA_su_phu_mi` | p4 | characters/su_phu_mi/faces (face ×8) |
| `obs_fence_break` | p4 | obstacles/fence/break (break ×6) |
| `obs_fence_regrow` | p4 | obstacles/fence/regrow (regrow ×6) |
| `obs_stall_break` | p4 | obstacles/stall/break (break ×6) |
| `obs_stall_regrow` | p4 | obstacles/stall/regrow (regrow ×6) |
| `obs_cart_break` | p4 | obstacles/cart/break (break ×6) |
| `obs_cart_regrow` | p4 | obstacles/cart/regrow (regrow ×6) |
| `obs_brick_break` | p4 | obstacles/brick/break (break ×6) |
| `obs_brick_regrow` | p4 | obstacles/brick/regrow (regrow ×6) |
| `obs_boulder_break` | p4 | obstacles/boulder/break (break ×6) |
| `obs_boulder_regrow` | p4 | obstacles/boulder/regrow (regrow ×6) |
| `obs_bamboo_wall_break` | p4 | obstacles/bamboo_wall/break (break ×6) |
| `obs_bamboo_wall_regrow` | p4 | obstacles/bamboo_wall/regrow (regrow ×6) |
| `vfx_run_dust` | p4 | vfx/run_dust (run_dust ×4) |
| `decal_mud_edge` | p4 | decals/mud_edge.png |
| `decal_pond_shore` | p4 | decals/pond_shore.png |
| `decal_stone_path_edge` | p4 | decals/stone_path_edge.png |
| `decal_bush_edge` | p4 | decals/bush_edge.png |
| `decal_ice_skid` | p5 | decals/ice_skid.png |
| `decal_snow_pile` | p5 | decals/snow_pile.png |
| `decal_frozen_reeds` | p5 | decals/frozen_reeds.png |
| `ambient_falling_leaf` | p5 | ambient/falling_leaf.png |
| `ambient_firefly` | p5 | ambient/firefly.png |
| `ambient_butterfly` | p5 | ambient/butterfly.png |
| `ambient_chicken` | p5 | ambient/chicken.png |
| `ambient_lantern` | p5 | ambient/lantern.png |
| `ambient_bird` | p5 | ambient/bird.png |
| `ambient_cloud_shadow` | p5 | ambient/cloud_shadow.png |
| `beast_dragon_leg` | p5 | beasts/dragon_leg.png |
| `beast_phoenix_feather` | p5 | beasts/phoenix_feather.png |
| `beast_mote_cloud` | p5 | beasts/mote_cloud.png |
| `beast_mote_spark` | p5 | beasts/mote_spark.png |
| `beast_mote_glow` | p5 | beasts/mote_glow.png |
| `beast_mote_feather` | p5 | beasts/mote_feather.png |
| `portrait_a_hon_neutral` | p5 | characters/a_hon/portrait_neutral.png |
| `portrait_a_hon_evil` | p5 | characters/a_hon/portrait_evil.png |
| `portrait_a_hon_panic` | p5 | characters/a_hon/portrait_panic.png |
| `portrait_a_hon_cry` | p5 | characters/a_hon/portrait_cry.png |
| `portrait_ba_tam_loa_neutral` | p5 | characters/ba_tam_loa/portrait_neutral.png |
| `portrait_ba_tam_loa_evil` | p5 | characters/ba_tam_loa/portrait_evil.png |
| `portrait_ba_tam_loa_panic` | p5 | characters/ba_tam_loa/portrait_panic.png |
| `portrait_ba_tam_loa_cry` | p5 | characters/ba_tam_loa/portrait_cry.png |
| `portrait_ong_ba_vong_neutral` | p5 | characters/ong_ba_vong/portrait_neutral.png |
| `portrait_ong_ba_vong_evil` | p5 | characters/ong_ba_vong/portrait_evil.png |
| `portrait_ong_ba_vong_panic` | p5 | characters/ong_ba_vong/portrait_panic.png |
| `portrait_ong_ba_vong_cry` | p5 | characters/ong_ba_vong/portrait_cry.png |
| `portrait_dai_ca_dep_neutral` | p5 | characters/dai_ca_dep/portrait_neutral.png |
| `portrait_dai_ca_dep_evil` | p5 | characters/dai_ca_dep/portrait_evil.png |
| `portrait_dai_ca_dep_panic` | p5 | characters/dai_ca_dep/portrait_panic.png |
| `portrait_dai_ca_dep_cry` | p5 | characters/dai_ca_dep/portrait_cry.png |
| `portrait_map_tra_sua_neutral` | p5 | characters/map_tra_sua/portrait_neutral.png |
| `portrait_map_tra_sua_evil` | p5 | characters/map_tra_sua/portrait_evil.png |
| `portrait_map_tra_sua_panic` | p5 | characters/map_tra_sua/portrait_panic.png |
| `portrait_map_tra_sua_cry` | p5 | characters/map_tra_sua/portrait_cry.png |
| `portrait_co_ba_ganh_neutral` | p5 | characters/co_ba_ganh/portrait_neutral.png |
| `portrait_co_ba_ganh_evil` | p5 | characters/co_ba_ganh/portrait_evil.png |
| `portrait_co_ba_ganh_panic` | p5 | characters/co_ba_ganh/portrait_panic.png |
| `portrait_co_ba_ganh_cry` | p5 | characters/co_ba_ganh/portrait_cry.png |
| `portrait_su_phu_mi_neutral` | p5 | characters/su_phu_mi/portrait_neutral.png |
| `portrait_su_phu_mi_evil` | p5 | characters/su_phu_mi/portrait_evil.png |
| `portrait_su_phu_mi_panic` | p5 | characters/su_phu_mi/portrait_panic.png |
| `portrait_su_phu_mi_cry` | p5 | characters/su_phu_mi/portrait_cry.png |
| `portrait_be_ti_neutral` | p5 | characters/be_ti/portrait_neutral.png |
| `portrait_be_ti_evil` | p5 | characters/be_ti/portrait_evil.png |
| `portrait_be_ti_panic` | p5 | characters/be_ti/portrait_panic.png |
| `portrait_be_ti_cry` | p5 | characters/be_ti/portrait_cry.png |
| `portrait_shipper_neutral` | p5 | characters/shipper/portrait_neutral.png |
| `portrait_shipper_evil` | p5 | characters/shipper/portrait_evil.png |
| `portrait_shipper_panic` | p5 | characters/shipper/portrait_panic.png |
| `portrait_shipper_cry` | p5 | characters/shipper/portrait_cry.png |
| `portrait_cu_rua_neutral` | p5 | characters/cu_rua/portrait_neutral.png |
| `portrait_cu_rua_evil` | p5 | characters/cu_rua/portrait_evil.png |
| `portrait_cu_rua_panic` | p5 | characters/cu_rua/portrait_panic.png |
| `portrait_cu_rua_cry` | p5 | characters/cu_rua/portrait_cry.png |
| `char_a_hon_ref` | p9 | characters/a_hon/ref_sheet.png |
| `char_ba_tam_loa_ref` | p9 | characters/ba_tam_loa/ref_sheet.png |
| `char_ong_ba_vong_ref` | p9 | characters/ong_ba_vong/ref_sheet.png |
| `char_dai_ca_dep_ref` | p9 | characters/dai_ca_dep/ref_sheet.png |
| `char_map_tra_sua_ref` | p9 | characters/map_tra_sua/ref_sheet.png |
| `char_co_ba_ganh_ref` | p9 | characters/co_ba_ganh/ref_sheet.png |
| `char_su_phu_mi_ref` | p9 | characters/su_phu_mi/ref_sheet.png |
| `char_be_ti_ref` | p9 | characters/be_ti/ref_sheet.png |
| `char_shipper_ref` | p9 | characters/shipper/ref_sheet.png |
| `char_cu_rua_ref` | p9 | characters/cu_rua/ref_sheet.png |
