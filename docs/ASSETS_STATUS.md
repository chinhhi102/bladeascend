# Tình trạng art

Tạo lại file này: `cd server && go run ./cmd/artgen status`.
Asset thiếu → game tự vẽ placeholder (hình tròn + chữ). Sinh thêm: `go run ./cmd/artgen run -p 2 -n 30` rồi `go run ./cmd/artgen prep`.
Nếu tự tải ảnh từ Gemini: lưu thành `assets/_raw/<tên job>.png` rồi chạy `prep`.

## Còn thiếu (130 ảnh)

| Job | Ưu tiên | Ra file |
|---|---|---|
| `char_a_hon_hit` | p2 | characters/a_hon/frames (hit ×4) |
| `char_a_hon_death` | p2 | characters/a_hon/frames (death ×8) |
| `char_ba_tam_loa_hit` | p2 | characters/ba_tam_loa/frames (hit ×4) |
| `char_ba_tam_loa_death` | p2 | characters/ba_tam_loa/frames (death ×8) |
| `char_dai_ca_dep_hit` | p2 | characters/dai_ca_dep/frames (hit ×4) |
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
| `ui_badges` | p2 | ui (badge ×9), ui (medal ×3) |
| `ui_home_bg` | p2 | ui/home_bg.png |
| `char_a_hon_chase` | p3 | characters/a_hon/frames (chase ×8) |
| `char_a_hon_flee` | p3 | characters/a_hon/frames (flee ×8) |
| `char_ba_tam_loa_chase` | p3 | characters/ba_tam_loa/frames (chase ×8) |
| `char_ba_tam_loa_flee` | p3 | characters/ba_tam_loa/frames (flee ×8) |
| `char_ong_ba_vong_chase` | p3 | characters/ong_ba_vong/frames (chase ×8) |
| `char_ong_ba_vong_flee` | p3 | characters/ong_ba_vong/frames (flee ×8) |
| `char_dai_ca_dep_chase` | p3 | characters/dai_ca_dep/frames (chase ×8) |
| `char_dai_ca_dep_flee` | p3 | characters/dai_ca_dep/frames (flee ×8) |
| `char_map_tra_sua_chase` | p3 | characters/map_tra_sua/frames (chase ×8) |
| `char_map_tra_sua_flee` | p3 | characters/map_tra_sua/frames (flee ×8) |
| `char_co_ba_ganh_chase` | p3 | characters/co_ba_ganh/frames (chase ×8) |
| `char_co_ba_ganh_flee` | p3 | characters/co_ba_ganh/frames (flee ×8) |
| `char_su_phu_mi_chase` | p3 | characters/su_phu_mi/frames (chase ×8) |
| `char_su_phu_mi_flee` | p3 | characters/su_phu_mi/frames (flee ×8) |
| `char_be_ti_chase` | p3 | characters/be_ti/frames (chase ×8) |
| `char_be_ti_flee` | p3 | characters/be_ti/frames (flee ×8) |
| `char_shipper_chase` | p3 | characters/shipper/frames (chase ×8) |
| `char_shipper_flee` | p3 | characters/shipper/frames (flee ×8) |
| `char_cu_rua_chase` | p3 | characters/cu_rua/frames (chase ×8) |
| `char_cu_rua_flee` | p3 | characters/cu_rua/frames (flee ×8) |
| `decor_sheet` | p3 | decor (decor ×10), decor (wall ×1) |
| `zone_wind` | p3 | zones (wind ×4) |
| `vfx_levelup_aura` | p3 | vfx/levelup_aura (levelup_aura ×6) |
| `vfx_lightning` | p3 | vfx/lightning (lightning ×6) |
| `vfx_shield_break` | p3 | vfx/shield_break (shield_break ×6) |
| `vfx_stun_stars` | p3 | vfx/stun_stars (stun_stars ×4) |
| `vfx_crack` | p3 | vfx/crack.png |
| `ui_scroll` | p3 | ui/scroll.png |
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
| `vfx_run_dust` | p4 | vfx/run_dust (run_dust ×4) |
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

## Đã có (66 ảnh)

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
| `item_orb_small` | p0 | items/orb_small.png |
| `item_orb_medium` | p0 | items/orb_medium.png |
| `item_orb_large` | p0 | items/orb_large.png |
| `buff_spin` | p0 | buffs/spin.png |
| `buff_haste` | p0 | buffs/haste.png |
| `buff_magnet` | p0 | buffs/magnet.png |
| `buff_long` | p0 | buffs/long.png |
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
| `obs_stall_states` | p1 | obstacles/stall (states ×4) |
| `obs_cart_states` | p1 | obstacles/cart (states ×4) |
| `obs_brick_states` | p1 | obstacles/brick (states ×4) |
| `obs_boulder_states` | p1 | obstacles/boulder (states ×4) |
| `obs_stone_wall_states` | p1 | obstacles/stone_wall (states ×4) |
| `gate_purple` | p1 | gates/purple (idle ×6), gates/purple (warn ×4), gates/purple (burst ×6) |
| `gate_blue` | p1 | gates/blue (idle ×6), gates/blue (warn ×4), gates/blue (burst ×6) |
| `gate_orange` | p1 | gates/orange (idle ×6), gates/orange (warn ×4), gates/orange (burst ×6) |
| `chest` | p2 | items/chest (idle ×4), items/chest (open ×6) |
| `char_ong_ba_vong_hit` | p2 | characters/ong_ba_vong/frames (hit ×4) |
| `char_ong_ba_vong_death` | p2 | characters/ong_ba_vong/frames (death ×8) |
| `char_dai_ca_dep_death` | p2 | characters/dai_ca_dep/frames (death ×8) |
| `char_map_tra_sua_hit` | p2 | characters/map_tra_sua/frames (hit ×4) |
| `ui_speech` | p2 | ui/speech_bubble.png |
