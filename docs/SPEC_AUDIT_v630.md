# Rà soát GAME_SPEC v6.30 + ADMIN_SPEC v1.4 so với code — 01/11/2026

Đối chiếu `docs/_raw/GAME_SPEC.md` (v6.30) và `docs/_raw/ADMIN_SPEC.md` (v1.4) với code ở nhánh `main`
(HEAD `7f922ff`, có thêm thay đổi i18n chưa commit của agent khác trong cây làm việc).
`go test ./...` pass toàn bộ lúc rà. `npm run assets:budget` đạt (chặn vào trận 4,45 MB / 8 MB, toàn bộ 7,83 MB / 80 MB).
`docs/V6_AUDIT.md` đã cũ: nhiều mục ghi CHƯA ở đó giờ đã XONG (FlagBot, bot đứng yên 5%, bot dùng kỹ năng theo v6.9,
mệnh giá Tiên khí trong config, bỏ khói khi chém vật cản, bảng xếp hạng phòng §11.5).

Trạng thái: **XONG** · **MỘT PHẦN** · **CHƯA** · **ĐANG LÀM** (agent khác đang làm trên nhánh riêng: admin React+antd,
góp ý 3 ảnh, i18n/SEO đa ngôn ngữ — không chấm theo code `main`).

## Tóm tắt

| Trạng thái | GAME_SPEC | ADMIN_SPEC | Tổng |
|---|---|---|---|
| XONG | 100 | 10 | 110 |
| MỘT PHẦN | 11 | 4 | 15 |
| CHƯA | 3 | 0 | 3 |
| ĐANG LÀM | 2 | 3 | 5 |

§17 (81 dòng): **PASS 73 · FAIL 3 · CẦN MÁY THẬT 2** · 3 dòng **lỗi thời** (mâu thuẫn với luật hiện hành, không chấm).

---

## 1. Bảng theo mục GAME_SPEC

### §1–§2 Tổng quan, luồng màn hình
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Trang chủ: tên + nhân vật + Vào chơi + Bảng xếp hạng | XONG | `client/src/ui/home.ts` | |
| Tên 2–16 ký tự, có dấu, lọc tục, cho trùng | XONG | `config/game.json:126-127`; `TestNameFilter` | |
| Nhớ tên + nhân vật (localStorage) | XONG | `client/src/ui/home.ts:14-22` | |
| Điều khiển WASD/mũi tên/chuột, joystick | XONG | `client/src/input.ts` | |

### §3 Tu luyện
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| scale 1→5 theo tầng; vòng k = 10k kiếm; max_swords | XONG | `TestScaleRange`, `TestMaxSwords`, `TestRingCapacityGrowsByTen` | |
| exp_needed = 100·2^r·(1+0.08s), bảng tổng | XONG | `TestExpNeededMatchesSpecTable`, `TestRealmTotalsMatchSpec` | |
| Đại Thừa 10 đầy thanh thì bỏ Linh khí, chỉ nhặt Tiên khí | XONG | `TestTienEntryNeedsFullBar`, `TestTienKhiAndLinhKhiDoNotMix` | |
| Tiên giới theo tổng Tiên khí, trần 1000/3000/6000/10000/15000 | XONG | `config/game.json:27-33`; `TestTienKhiPromotesAtThresholds` | |
| Viên không nhặt được vẽ mờ 30% | XONG | `client/src/render/orblayer.ts:17` | |
| Chết ở Tiên giới: người chém nhận 30% (trần +1 cảnh giới Tiên), luôn rơi 30%, mệnh giá 500/100/10/2/1 ≤ 40 viên | XONG | `config/game.json:35-40`; `TestDaoToKilledDropsAndCapsTien`, `TestMortalKillerStillDropsTien`, `TestSplitTienOrbsCapsAt40` | |
| Đạo Tổ nạp đầy 15 000 rồi 10 Tiên khí = 1 Uy Danh; Thiêu Đốt kéo xuống | XONG | `TestDaoToFillsThenConverts`, `TestDaoToCanBurn`; HUD `client/src/ui/hud.ts:315` | |
| Thiên Kiếp + kill feed "vượt Thiên Kiếp" | XONG | `client/src/render/breakthrough.ts`; `strings.ts:80` `tien_ascend` | |
| Hào quang Tiên 5 màu sau lưng | XONG | `client/src/render/tienaura.ts`; `config/game.json:42-48` | |
| **Đạo Tổ hiện trên minimap cả phòng** | **CHƯA** | `client/src/ui/minimap.ts:248-294` chỉ vẽ Top 1, Đài, rương | Server không gửi vị trí Đạo Tổ, client không vẽ |
| §3.5 Thần thú: đột phá xoắn 1.5 vòng, lượn thường trực, số đốt theo chất lượng 16/10/6 | XONG | `client/src/render/beast.ts`; `quality.ts:25` | |
| §3.6 Blend cộng + bloom riêng `vfxLayer` + Shockwave + hit-stop 80ms | XONG | `client/src/render/vfxlayer.ts:38`; `breakthrough.ts:7,61` | |
| §3.6 GlowFilter / GodrayFilter, particle `revolt-fx` + preset JSON, texture CC0 + `assets/CREDITS.md` | MỘT PHẦN | `revolt-fx` chỉ có trong `client/package.json`, không file nào import | Không Glow/Godray, không `vfx/presets/*.json`, không `assets/CREDITS.md` |

### §4 Kiếm, tốc độ, Thiêu Đốt
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Bán kính vòng chỉ theo k + scale; kiếm thứ 11 mở vòng 2 | XONG | `TestRingRadiusIndependentOfSwordCount`, `TestRingsForSwordCount` | |
| So le, nghiêng 35°, lật khi đổi chiều, màu từng vòng | XONG | `TestSlotAngleStagger`, `TestTiltFlipsWithDirection`; `config/game.json:83-92` | |
| 1 vòng/giây, Tốc Kiếm ×1.6 không đổi tầm | XONG | `TestOneRevolutionPerRevSecs`, `TestSpinBuffDoesNotChangeReach` | |
| Va chạm theo cung quét | XONG | `TestSweptHitCatchesFastBlades` | |
| Tốc độ CỘNG trên base, kẹp ×2 (Gom Đơn ×3), sàn 0.4 | XONG | `config/game.json:129-141`; `TestSpeedAdditiveCapped`, `TestSpeedFloorAndFrozen` | |
| Camera = max(900·scale^0.85, 2.6·reach) | XONG | `client/src/config.ts:202` | |
| Thiêu Đốt: Shift / nháy đúp giữ / 🔥; 8% EXP tầng/s; không rớt cảnh giới; Tiên giới 0.8% | XONG | `client/src/input.ts:15-105`; `TestBurnStopsAtRealmFloor`, `TestTienBurnRatePerRealm` | |

### §5 Chiến đấu
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Chém trúng thân = chết ở mọi cảnh giới (còn ≤ 1 vòng) | XONG | `TestBodyHitKillsAnyRealm`, `TestBodyHitThresholdComesFromConfig` | |
| Người chém nhận 30% (trần +2 đại cảnh giới), bot 15%, rơi 30% + 50% kiếm, mất 40% | XONG | `config/game.json:112-122`; `TestKillRewardCappedAtTwoRealms`, `TestStealShareHumanVsBot`, `TestKillRewardAndDeathSwordDrop` | |
| Khiên Kim Cang: vỡ, không chết, bất tử 1s | XONG | `TestShieldSavesFromDeath` | |
| 1 hit/nạn nhân/tick | XONG | `TestOneHitPerVictimPerTick` | |
| Bể kiếm mỗi tick, đổi chiều cooldown 0.3s, chỉ cùng cảnh giới mới đổi | XONG | `TestClashGrindsOneOuterRingPerSecond`, `TestRingFlipKeepsItsOwnCooldown`, `TestClashAcrossRealmsDoesNotFlip` | |
| Độ bền 1.2^r, Cương Kiếm ×1.5, crack | XONG | `TestDurabilityTwelveClashesTradeTwelveForTen`, `TestHardBuffMakesLowRealmTheHarderSide` | |
| Đi xuyên vòng kiếm nhau; số kiếm về 0 không chết | XONG | `TestPlayersDoNotBlockEachOther`, `TestRingCanBeStrippedToZero` | |
| Quái: viên EXP chỉ người hạ nhặt trong 1.5s | XONG | `TestMonsterOrbsReservedForKiller` | |
| Vật cản đủ bậc → biến mất ngay, kiếm không mất | XONG | `TestCutObstacleWhenRealmEnough`; commit `4625455` | |

### §6 EXP, vật phẩm, kỹ năng
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| §6.0 viên × 1.5^r, hút × √scale, quái × 1.5^r người chém, bot 15% | XONG | `TestAmbientOrbScalesWithPickerRealm`, `TestMonsterExpScalesWithKillerRealmOnce`, `TestPickupRadiusSqrtScaleAndMagnet` | |
| §6.0 CI `sim --farm` và `--magnet-always` (±25%, Luyện Hư→Chân Tiên ≥ 15 phút) | XONG | `.github/workflows/ci.yml:49-52` | |
| §6.0 hệ số cuối do sim chỉnh "rồi ghi lại vào đây" | MỘT PHẦN | `docs/EXP_AUDIT.md:141-142` (Nam Châm 1.2, cell 160) | Spec vẫn ghi 1.5 và 140 (xem Số liệu lệch) |
| §6.1 viên mọc lại 5s trong ô 1000² | XONG | `TestAmbientOrbsRegrowInSameRegion` | Mật độ 160 chứ không 140 |
| §6.2 9 vật phẩm, weight, buff làm mới thời gian, 200/phòng, cách người ≥ 300 | XONG | `config/game.json:179-198`; `TestBuffExpiresAndRefreshes`, `TestItemPoolRespawnsAfterDelay` | |
| Ẩn Thân 80%, 5s, chém trúng là lộ | XONG | `TestStealthCappedAtMaxSecs`, `TestStealthEndsWhenOwnerHitsSomeone` | |
| Rương Báu mỗi 60s, báo trước 10s, nổ 8 món | XONG | `TestChestAnnouncedThenBurstsIntoItems` | |
| §6.3 Tự hồi + tích lượt đúng bảng, Tiên +1, vào trận 1 lượt, khoá 3s, chết mất lượt | XONG | `config/game.json:512-526`; `TestSkillCooldownTableAndTien`, `TestSkillLockout`, `TestDeathClearsCharges` | |
| Mọi phạm vi theo U = 90×scale, đẩy theo U người dùng | XONG | `TestNoHardcodedSkillRanges`, `TestPushDistanceUsesCasterU` | |
| Bà Tám toả tròn 360° 6 U, đổi chiều vòng | XONG | `TestLoaPhuongPushesAllAround` | |
| Ông Ba: hát 0.5s, ngủ 2.5s trong 4 U, Trộm Kiếm 4s | XONG | `TestRuNguSleepsAfterChannel`, `TestRuNguCancelledByBodyHit`, `TestTromKiemStealsClashedSwords`, `TestTromKiemOverflowDropsBundles` | |
| **Người ngủ gục: không kỹ năng / không Thiêu Đốt** | **MỘT PHẦN** | Đứng im đúng: `server/internal/game/movement.go:23-25`. Nhưng `CanCastSkill` chỉ chặn choáng `skills.go:133-135`; `Burning` chỉ chặn choáng `movement.go:115-121` | Người đang ngủ vẫn dùng được kỹ năng; giữ Shift khi ngủ vẫn bị trừ Linh khí dù đứng im |
| Đại Ca 2 đàn em, 30% kiếm, 8s | XONG | `TestMinionSwordsHalfOfBoss` (đọc `dan_em_sword_share`), `TestMinionsCannotBodyHit` | |
| Thằng Mập bịt mắt 5s: không nhặt, kiếm không giết | XONG | `TestTranChauBlinds`, `TestBlindedSwordCannotKill`, `TestBlindBlocksEveryPickup`; `client/src/render/boba.ts` | |
| Cô Ba sạp bậc +1, chủ đi xuyên, vỏ chuối trượt + choáng 1s, 6s | XONG | `stall_test.go` (8 test) | |
| Sư Phụ Mì cướp 25% (trần 20% tối đa mình), hụt mất lượt | XONG | `TestHutMiStealsQuarterOfTargetSwords`, `TestHutMiCappedByOwnMax`, `TestHutMiMissStillSpendsCharge` | |
| Bé Tí gồng 4s độ bền ×2, miễn khống chế | XONG | `TestGongCoBlocksControl`, `TestBraceImmuneToStun` | |
| **Shipper Gom Đơn: dùng được khi đang trói/ngủ/choáng và thoát ngay** | **MỘT PHẦN** | Xoá khống chế `skills.go:398-404`; xuyên vật cản `TestGhostPassesEverythingButBorder` | `CanCastSkill` trả false khi **choáng** (`skills.go:135`) → dẫm vỏ chuối là không Gom Đơn được. Không có test cho vế này |
| Cụ Rùa rụt cổ 5s, sóng 2 U đẩy 3 U | XONG | `TestRutCoProtectsThenBursts` | |
| A Hớn hồi 20% kiếm tối đa | XONG | `TestBanhBaoRestoresShareOfMax` | |
| Bot dùng kỹ năng theo luật từng nhân vật | XONG | `server/internal/bot/skill_test.go` (6 test) | |
| Cân bằng theo telemetry người thật (không theo tay đôi) | XONG | commit `300c69f`, `1476271` (bảng Nhân vật trong admin) | |

### §7 Chết & Tổng kết
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Chết = hết ván, không tự vào lại, tải lại vẫn ở Tổng kết | XONG | `TestDeadHumanNeverAutoRejoins`; `client/src/ui/runend.ts:10`; `main.ts:229` | |
| Màn Tổng kết: ai chém, câu tấu hài, cảnh giới đỉnh, chém, sống, hạng, bảng Uy Danh theo nguồn, 3 nút | XONG | `client/src/ui/summary.ts` | |
| Vào lại Luyện Khí 1, bất tử 5s, chỗ an toàn 800 | XONG | `config/game.json:111,403` | |

### §8 Bản đồ
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| 5 khu 12 000², chỉ qua bằng cổng | XONG | `TestRoomHasFiveZones`, `TestPlayerCannotLeavePlayArea` | |
| Băng trơn có quán tính | XONG | `TestIceKeepsSlidingAfterInputStops` | |
| Spawn khu ít người nhất; banner tên khu 1.5s; M phóng to + 5 khu + số người | XONG | `TestNewPlayersSpreadAcrossZones`; `main.ts:8,502`; `minimap.ts:161-170` | |
| Kill feed tiền tố `[khu]` | XONG | `client/src/ui/hud.ts:548` | |
| §8.1 mapgen Thiên Đình + `--png` | XONG | `server/cmd/mapgen` | |
| §8.1b `mapgen --check` 5 khu × 14 cảnh giới trong CI | XONG | `ci.yml:43`; `maps/check.json` `"bad": 0` | |
| §8.2 bậc 0–13, Cương Kiếm +1, "Cần {cảnh giới}", mọc lại 30s, sạp 30% rơi đồ | XONG | `config/game.json:281-311`; `TestHardBuffCutsOneTierHigher`, `TestTienRealmsStillBlockedByHighTiers`, `TestObstacleRegrowsWhenClear` | |
| Không vật cản ∞ trong khu, mỗi khu có bậc ≥ 9 | XONG | `TestEveryZoneHasImmortalTierButNoInfiniteWall` | |
| Chân đế 30% đáy + y-sort | XONG | `TestFootprintBlocksOnlyTheBase` | |
| §8.2a Chung mã va chạm client↔server, trượt dọc, hiệu chỉnh mượt < 30 | XONG | `cmd/collisionfixture` + `npm run check:collision`; `config/game.json:504-505` | |
| §8.3 bụi / bùn / gió / hẻm | XONG | `TestStealthAndBushVisibility`, `TestWindAngles`, `TestNarrowAlleyBlocksBigPlayers` | |
| §8.4 cổng ngẫu nhiên, không ra lại cổng vào, lóe 0.5s, không bất tử, cooldown 8s | XONG | `TestGateExitIsRandom`, `TestGateTeleportsWithoutInvulnerability` | |
| **§8.4 Thời gian dịch chuyển 0.8s** | **MỘT PHẦN** | `config/game.json:324` `transit_secs: 2.0`; `server/internal/game/gates.go:45-46` | Người chơi chờ 2s thay vì 0.8s (cả Thiên Môn) |
| §8.4a Thiên Môn 4 khu → bến cố định ở Thiên Đình; icon + mũi tên minimap | XONG | `TestHeavenGateAlwaysLeadsToHeaven`; `minimap.ts:118-123,258-263` | Khoảng cách bến 1400 hardcode `server/cmd/mapgen/zones.go:86` |
| §8.5 Đài nhỏ 5s/30s, Thiên Tiên Đài 150s báo trước 10s, viên sống 60s | XONG | `TestSmallAltarSpraysOnSchedule`, `TestCenterAltarGiantOrbAndWarning`, `TestTienKhiOrbExpires` | |
| §8.6 vùng hữu cơ, server dùng đúng đa giác | XONG | `TestPolygonZoneUsesTheDrawnShape`, `TestGeneratedMapsHaveOrganicZones` | |

### §9 Phòng, bot
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Trần người thật chỉnh ở admin, mặc định theo APP_ENV, sống qua deploy | XONG | `config/game.json:405-408`; `TestRoomCapFollowsAdminSetting`, `TestSettingRoundTrip` | |
| Sàn 10 nhân vật/khu; tự tắt bot ≥ ngưỡng, bật lại < ngưỡng−5; auto/on/off toàn server + từng phòng | XONG | `TestZoneFloorPerZone`, `TestBotAutoHysteresis`, `TestBotModeOverrides`, `TestBotsOffNoSpawnNoRespawn` | |
| Matchmaking phòng đông nhất; phòng hủy sau 60s trống | XONG | `TestMatchmakingPrefersFullestRoom`, `TestEmptyRoomCloses` | |
| Bot sinh như người mới; "Rút bot ngay" | XONG | `TestBotSpawnsAsNewPlayer`, `TestWithdrawSkipsBotsInCombat` | |
| §9.4 1 000 tên, không chữ "bot", không trùng, nhớ 10 phút, 70/30, 5% đứng yên | XONG | `config/bot_names.txt` (1000 dòng); `TestBotNamesUniqueInRoom`, `TestRecentNameNotReused`, `TestBotIdlesAfterJoin` | |
| Phân bố nhân vật bot theo thống kê **7 ngày** | MỘT PHẦN | `server/internal/room/botnames.go:117` | Lấy theo người thật đang ở trong phòng, không phải 7 ngày |
| Tên có ký tự ngoài BitmapFont vẽ bằng font dự phòng | MỘT PHẦN | `client/src/render/fonts.ts` | Không có nhánh font dự phòng |
| Bot không lộ (không cờ bot trong gói) | XONG | `TestSnapshotNeverFlagsBots` | |

### §10 Bot AI
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Farm/Hunt/Flee/Ambush/Wander, quyết định 0.3–0.6s, A* lưới 100 | XONG | `server/internal/bot/bot.go`; `config/game.json:428-438`; `TestFindPathCutting*` | |
| Tôn trọng bụi, Ẩn Thân thấy 50%; dùng cổng/Thiên Môn | XONG | `TestBotSightHalvedAgainstStealth`, `TestBotsEventuallyChangeZone`; `bot.go:369-376` | |

### §11 Bảng xếp hạng
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Uy Danh: 100·(r+1)·lật kèo + 50, đột phá 20r, Top 1 +1/s, bot 50%, chống cày 60s | XONG | `TestScoreUpsetExample` (850), `TestScoreSameVictimCooldown`, `TestScoreBotVictimHalved`, `TestScoreTop1PerSecond` | |
| 3 bảng toàn cục × 3 kỳ, API, bot không ghi, debug không ghi | XONG | `server/cmd/server/http.go:146-160`; `TestTopByBoardAndPeriod`, `TestSpeedBoardOnlyReachedDaiThua`; `room.go:522` | |
| Danh hiệu Top 1–10 tính cả phòng, giữ 3s | XONG | `TestTop1IsRoomWideNotPerZone`, `TestTitleNeedsHold` | |
| §11.5 bục Top 1/2/3, dòng 4–10, ghim dòng mình, khu + cảnh giới, Ẩn Thân "???", thu gọn điện thoại | XONG | `client/src/ui/roomboard.ts`; `TestLeaderboardRowsCarryZoneRealmAndHideStealth` | Ảnh duyệt chờ chủ dự án |
| §11.4 tin loan toàn server Top 5 | XONG | `TestTopFiveKillAnnouncedServerWide` | |

### §12 Anim, bảng tên, HUD, minimap
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Mood server gửi, giữ 0.5s | XONG | `config/game.json:484-492`; `server/internal/game/mood.go` | |
| §12.3 bảng tên: ruy băng hạng, huy hiệu cảnh giới, ▲●▼, tên mình vàng ▼, cache texture | XONG | `client/src/render/nameplate.ts:118,178` | Font dự phòng (xem §9.4) |
| §12.4 chân dung HUD, viền màu cảnh giới, 4 biểu cảm | XONG | `client/src/ui/hud.ts:128,362-368` | |
| §12.5 B chuẩn hoá frame + điểm giật | MỘT PHẦN | `docs/JITTER_REPORT.md:8-11` | Còn **1** anim lặp đỏ (≥ 45) |
| §12.5 C bộ điều khiển anim (crossfade, giữ pha, lật có trễ…) | XONG | `client/src/render/animctl.ts` | |
| **§12.5 A một thân nhiều mặt** | **CHƯA** | Không file client nào đọc `faces/` | Art mặt có cho 3 nhân vật nhưng không ghép |
| §12.2 thanh máu quái | XONG | `client/src/render/entities.ts:76` | |
| §12.1 minimap: mình, khung camera, cổng, Top 1, rương, Đài; không người khác | XONG | `client/src/ui/minimap.ts:248-294` | Cỡ điện thoại 96 px (spec 120) |

### §13 Kỹ thuật
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Server authoritative, tick 30 / snapshot 20 | XONG | `config/game.json:493-497` | |
| Delta snapshot, lượng tử hoá, ≤ 150 entity, xa 10 Hz, bỏ snapshot cũ khi nghẽn | XONG | `server/internal/room/broadcast.go:22-26,156`; `client.go:365` | |
| **§13.6 #1 mỗi vòng kiếm = 1 RenderTexture** | MỘT PHẦN | `client/src/render/swordring.ts:82-83` dùng ParticleContainer | Khác cách spec bắt buộc (§4.1b lại cho dùng ParticleContainer) |
| **§13.6 #6 giải mã MessagePack trong Web Worker** | **CHƯA** | không có `new Worker` trong `client/src` | |
| §13.6 #8 chất lượng mặc định Vừa/Thấp, tự hạ/nâng 10s/20s, khoá 5 phút, chỉ lưu tay | XONG | `client/src/render/quality.ts:19-40,165-185`; `npm run check:quality` | |
| F3 bảng đo, `cmd/loadtest` | XONG | `main.ts:168`; `server/cmd/loadtest` | |
| §13.5 rate limit tin nhắn + kết nối/IP | XONG | `config/game.json:501-502` | |
| §13.7 pipeline WebP + hash + nhóm tải + ngân sách CI + màn tải theo byte + Thử lại | XONG | `tools/asset-pipeline`; `client/scripts/assets-budget.ts`; `client/src/ui/loading.ts:26,74,150` | |
| §13.7-3b thư viện art không build | XONG | `assets/library/INDEX.md` | |

### §14–§16
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| Quái 4 loại HP/EXP/drop/weight, boss ≤ 1 | XONG | `config/game.json:231-274` | ~150 quái là **mỗi khu** (`items.go:270`), tức 750/phòng — spec ghi "/phòng" |
| §16 mọi số trong `config/game.json` | MỘT PHẦN | | Lệch giá trị — xem bảng Số liệu lệch |

### §18 Debug
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| `?debug=true` chỉ có tác dụng khi có phiên admin; server từ chối + audit; dev localhost ngoại lệ | XONG | `TestDebugDeniedWithoutAdminSession`, `TestDebugDevGateNeedsLocalhost`, `TestDebugTokenOnlyInDev`; `cmd/server/admin.go:280,298` | |
| Phím E/R/K/1–7/X/B/hitbox | XONG | `client/src/main.ts:55`; `server/internal/room/debug.go:27-71` | |
| Tắt hẳn debug production bằng biến môi trường | MỘT PHẦN | `server/internal/admin/debuggate.go:40-41` | Không có biến tắt hẳn |
| §18.1 bảng hướng dẫn lúc vào trận (có dòng Kỹ năng) | XONG | `client/src/ui/hud.ts` | |

### §19–§20
| Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|
| §19 i18n 6 ngôn ngữ, `i18n/{lang}.json`, chọn ngôn ngữ, font Thai | ĐANG LÀM | `client/src/i18n/` (chưa commit) | |
| §19 SEO đa ngôn ngữ `/en/…`, hreflang, sitemap | ĐANG LÀM | chỉ có trang vi `client/index.html` | |
| §20 tên BladeAscend / Đạo Hữu Xin Dừng Bước, title, description, khẩu hiệu, đoạn giới thiệu, `config.brand` | XONG | `config/game.json:2-7`; `client/index.html:17-38,147` | |
| Bỏ "Kiếm Xoay Tu Tiên" trong code, CI kiểm | XONG | `npm run check:brand` | Spec §13.2 vẫn còn `kiem-xoay/` |

---

## 2. Bảng theo mục ADMIN_SPEC v1.4

| Mục | Luật | TT | Bằng chứng | Thiếu |
|---|---|---|---|---|
| §1 | bcrypt từ env, `cmd/adminpass`, phiên 32 byte 12h, khoá 5 lần/15 phút, chặn HTTP thường, IP allowlist, audit đăng nhập | XONG | `TestLoginLocksAfterFiveFailures`, `TestSessionExpires`, `TestRejectsPlainHTTPFromRemote`, `TestIPAllowlist`, `TestAuditRecordsLogins` | |
| §1 | HTML/JS admin chỉ phục vụ sau đăng nhập; CSRF `X-Admin-Token` | XONG | `server/cmd/server/admin.go:18-35` | |
| §1.1 | debug gắn phiên, kiểm lại 60s, mạng debug không lên bảng | XONG | `server/internal/room/client.go:158`; `room.go:522` | |
| §2 / §2.1 | Bố cục + React + antd v5 | ĐANG LÀM | | Agent khác |
| §2 | Dữ liệu live qua `/api/admin/ws` | MỘT PHẦN | Không có route WS; `GET /api/admin/stats` (`admin.go:37`) | Đang hỏi định kỳ qua HTTP |
| §3 | Chỉ số dashboard | MỘT PHẦN | `server/cmd/server/adminstats.go:46-75` có người/bot/phòng, tick TB/p95 theo khu, goroutine, heap, phân bố cảnh giới, góp ý mới, cờ, lỗi | Thiếu CPU %, người vào/rời mỗi phút, thiết bị, KB/s, snapshot, ping p50/p95, FPS, mất WebGL, chém/chết mỗi phút, Tiên khí phun/nhặt |
| §4 | Cấu hình phòng (trần, sàn, ngưỡng) + cảnh báo loadtest + audit | XONG | `admin.go:55-56`; `adminroom.go:43-48` | |
| §4 | Điều khiển bot toàn server / từng phòng / Rút bot ngay | XONG | `admin.go:51-54`; `adminbots.go` | |
| §4 | Cột từng khu | MỘT PHẦN | `adminstats.go:22-34` | Thiếu số viên Linh/Tiên khí, đếm ngược Đài, lượt qua cổng/Thiên Môn mỗi phút, seed + phiên bản mapgen |
| §5 | Bản đồ toàn cảnh bằng bộ vẽ game | ĐANG LÀM | Luồng 5 Hz đã bỏ (`400c21f`) | Phần giao diện thuộc bản admin mới |
| §6 | Danh sách người chơi (tên, thiết bị, IP che, phòng/khu, cảnh giới, kiếm, Uy Danh, ping, FPS, hiệu chỉnh, cờ) | XONG | `server/cmd/server/players.go:22-46` | Không có cột quốc gia, không có API chi tiết `/players/{guest_id}` |
| §6.1 | 6 cờ nghi hack, không tự kick, bỏ cờ | XONG | `server/internal/room/flags.go:21-33`; `admin.go:45` | |
| §6.1 | Cờ lưu bằng chứng (bảng `cheat_flags`) | MỘT PHẦN | `flags.go` giữ trong RAM | Khởi động lại server là mất cờ |
| §6.2 | Kick / kick + cấm (guest_id + IP, thời hạn) / gỡ cấm, màn "bị mời ra" có lý do + giờ hết hạn | XONG | `players.go:176-255`; `cmd/server/ws.go:224-229`; `client/src/ui/kicked.ts` | |
| §7 | Góp ý: 3 ảnh, kéo-thả, dán, xoá EXIF, PreviewGroup | ĐANG LÀM | | Agent khác |
| §7 | Góp ý chữ: loại, 5–1000 ký tự, rate limit, người bị cấm không gửi, trạng thái + ghi chú, CSV | XONG | `server/cmd/server/feedback.go:86,117`; `admin.go:38-40,59` | |
| §8 | Bảng SQLite `admin_sessions/audit/login_attempts/bans/feedback` | XONG | `server/internal/admin` | Thiếu `cheat_flags` |

---

## 3. Số liệu lệch (spec ↔ `config/game.json`)

| Khoá | Spec | Config | Vị trí | Ghi chú |
|---|---|---|---|---|
| `buffs.magnet.mult` | 1.5 (§6.0 mục 3, §6.2, §16) | **1.2** | `config/game.json:156-157` | Chỉnh theo sim (`docs/EXP_AUDIT.md:141`); spec chưa ghi lại như §6.0 yêu cầu |
| `exp_orbs.cell` | 140 (§6.1 "1 viên / 140×140", §16) | **160** | `config/game.json:217` | Như trên (`EXP_AUDIT.md:142`) |
| `gates.transit_secs` | 0.8 (§8.4, §8.4a, §16) | **2.0** | `config/game.json:324` | Còn từ 22/10 (`docs/PROGRESS.md:621`), không có quyết định nào đổi |
| `minimap.size_mobile` | 120 (§12.1, §16) | **96** | `config/game.json:344` | Chưa có câu hỏi/quyết định trong QUESTIONS_DONE |
| `ice.accel` | 0.25 (§16) | **468.0** | `config/game.json:509` | Đổi đơn vị (đơn vị/s² thay hệ số/tick); spec §16 chưa cập nhật |
| `cultivation.tien_rob_share` | 0.5 (§16) — nhưng §3.4 nói 30% | 0.3 | `config/game.json:35` | Config đúng §3.4; §16 sai |
| `player.hit_drops_realm` | `true` (§16) | (không có) | — | Luật bỏ từ v6.26; §16 còn sót |
| `buffs.haste.mult` / `player.burn_speed_mult` | 1.4 / 1.8 nhân (§16) | không có; `speed.haste 0.4`, `speed.burn 0.8` cộng | `config/game.json:129-141` | Đúng v6.25; §16 còn số cũ |
| `room.monster_count` | "~150 quái/phòng" (§14) | 150 **mỗi khu** | `config/game.json:401`; `items.go:270` | 750 quái/phòng |
| Khoảng cách bến Thiên Môn | 1 400 (`heaven_gate.landing_r_from_center`, §16) | không có trong config | `server/cmd/mapgen/zones.go:86` (hằng số) | Vi phạm luật 2 (nhỏ) |
| `mapgen.min_open_ratio` | 0.65 (§8.1) vs ≥ 60% (§8.1b) | 0.65 | `config/game.json:393` | Spec tự mâu thuẫn |

---

## 4. Checklist §17 — từng dòng

| # | Dòng | Kết luận | Bằng chứng |
|---|---|---|---|
| 1 | Bà Tám đẩy 6 U mọi hướng + đổi chiều | PASS | `TestLoaPhuongPushesAllAround` |
| 2 | Ông Ba Trộm Kiếm 4s | PASS | `TestTromKiemStealsClashedSwords`, `TestTromKiemOverflowDropsBundles` |
| 3 | Chết → Tổng kết, không tự vào, tải lại vẫn ở Tổng kết | PASS | `TestDeadHumanNeverAutoRejoins`; `client/src/ui/runend.ts` |
| 4 | Admin trần 50→30, không đuổi ai, sống qua deploy | PASS | `TestRoomCapFollowsAdminSetting`, `TestSettingRoundTrip` |
| 5 | Production 40/35, staging 20, mỗi khu đủ 10 | PASS | `TestBotAutoHysteresis`, `TestZoneFloorPerZone`; `config/game.json:405-408` |
| 6 | Ru Ngủ / choáng: kiếm vẫn đánh trả | PASS | `TestSleepDoesNotDisableRing`, `TestSlickStuns` |
| 7 | Bịt mắt: không giết, vẫn bể kiếm | PASS | `TestBlindedSwordCannotKill` |
| 8 | Kim Đan bị chém chết, 30% trần +2, rơi 30% + 50% kiếm | PASS | `TestBodyHitKillsAnyRealm`, `TestKillRewardCappedAtTwoRealms`, `TestKillRewardAndDeathSwordDrop` |
| 9 | Kim Cang: không chết, khiên vỡ | PASS | `TestShieldSavesFromDeath` |
| 10 | Đạo Tổ chết rơi 30%, người chém trần +1 Tiên | PASS | `TestDaoToKilledDropsAndCapsTien` |
| 11 | Tốc ×2.0 / Gom Đơn ≤ ×3 / bùn ×1.4 | PASS | `TestSpeedAdditiveCapped` |
| 12 | Không Bí Kíp; hồi đúng bảng, Tiên +1, vào 1 lượt | PASS | `TestSkillCooldownTableAndTien`, `TestSkillStartsWithOneAndRecharges`; weights không có Bí Kíp |
| 13 | Bot mới sinh Luyện Khí 1, 1 kiếm | PASS | `TestBotSpawnsAsNewPlayer` |
| 14 | loadtest 50 người: tick p95 < 4 ms | PASS | `docs/PERF_LOG.md` lượt #24: i5-12400F 2,01 ms ✅ (staging ARM 5,18 ms ❌ → staging giữ 20) |
| 15 | 2 khu, 1 Top 1 cả phòng, bảng đúng khu + cảnh giới | PASS | `TestTop1IsRoomWideNotPerZone`, `TestLeaderboardRowsCarryZoneRealmAndHideStealth` |
| 16 | Bục Top 1/2/3, dòng 4–10, ghim dòng mình | PASS | `client/src/ui/roomboard.ts` |
| 17 | Không tràn 1366×768 và iPhone dọc, điện thoại thu gọn | CẦN MÁY THẬT | Code thu gọn `roomboard.ts:97`; cần iPhone thật để duyệt |
| 18 | Cô Ba sạp bậc 3, chủ xuyên, vỏ chuối, 6s, Shipper xuyên | PASS | `stall_test.go` (`TestStallTierIsOwnerRealmPlusOne`, `TestStallGhostPassesThrough`…) |
| 19 | Hút Mì 25, trần 20%, hụt mất lượt | PASS | `TestHutMi*` |
| 20 | Gom Đơn hút 3 U; **đang trói/ngủ/choáng dùng được** | **FAIL** | `skills.go:135` chặn khi choáng → không dùng được sau khi dẫm vỏ chuối |
| 21 | Luyện Khí chém rào biến mất, gạch "Cần Kim Đan", mọc lại 30s | PASS | `TestCutObstacleWhenRealmEnough`, `TestCannotCutObstacleBelowTier`, `TestObstacleRegrowsWhenClear` |
| 22 | `mapgen --check` 5 × 14 | PASS | `ci.yml:43`; `maps/check.json` bad = 0 |
| 23 | Không vật cản ∞, có bậc ≥ 9, Kim Tiên không chém cột ngọc, Đạo Tổ chém thạch bi | PASS | `TestEveryZoneHasImmortalTierButNoInfiniteWall`, `TestTienRealmsStillBlockedByHighTiers` |
| 24 | Loa đẩy 540 / 3 240 theo U người dùng, không xuyên tường | PASS | `TestPushDistanceUsesCasterU`, `TestKnockbackCannotPushOutOfPlayArea` |
| 25 | Không số phạm vi cố định | PASS | `TestNoHardcodedSkillRanges` |
| 26 | boot ≤ 8 MB, tổng ≤ 80 MB, lần 2 từ cache | PASS | `npm run assets:budget` 4,45 / 7,83 MB; cache immutable `server/cmd/server/http.go` |
| 27 | Màn tải theo byte, %, MB, Thử lại | PASS | `client/src/ui/loading.ts:74,150` |
| 28 | Qua cổng khu chưa tải chờ ≤ 2s | PASS | `client/src/render/assets.ts` |
| 29 | Tên bot không "bot", không trùng, 70% quay lại 4–12s | PASS | `botnames_test.go:19`, `TestBotNamesUniqueInRoom`, `TestBotDeathBehaviourConfig` |
| 30 | run→chase→flee chỉ đổi mặt + tư thế | **FAIL** | Không có lớp mặt ở client (§12.5 A) |
| 31 | `assets:report` không frame nào vượt ngưỡng | **FAIL** | `docs/JITTER_REPORT.md:10` còn 1 anim đỏ |
| 32 | 5 khu, 200 lượt cổng đều, không ra lại cổng vào | PASS | `TestRoomHasFiveZones`, `TestGateExitIsRandom`, `TestGateEventuallyReachesEveryZone` |
| 33 | Spawn khu ít người nhất; "tổng 20/phòng" | PASS (vế 1) | `TestNewPlayersSpreadAcrossZones`; vế "20/phòng" lỗi thời (xem §5) |
| 34 | Băng có quán tính | PASS | `TestIceKeepsSlidingAfterInputStops` |
| 35 | 4 Thiên Môn, ra đúng bến, minimap icon + mũi tên | PASS | `TestHeavenGateAlwaysLeadsToHeaven`; `minimap.ts:118,258` |
| 36 | Đi sát vật cản không giật, F3 hiệu chỉnh ≈ 0 | PASS | `npm run check:collision` + `cmd/collisionfixture` |
| 37 | Trân châu che ~60% 5s, không nhặt, HUD không bị che | PASS | `client/src/render/boba.ts`; `TestBlindBlocksEveryPickup` |
| 38 | Shipper xuyên vách, không ra biên, kẹt thì đẩy ra | PASS | `TestGhostPassesEverythingButBorder`, `TestGhostUnsticksSmoothly` |
| 39 | Hai Kim Đan mài 1s còn ~30; đi xuyên; chỉ 1 vòng mới bị chém | PASS | `TestClashGrindsOneOuterRingPerSecond`, `TestPlayersDoNotBlockEachOther` |
| 40 | Tốc Kiếm không đổi tầm | PASS | `TestSpinBuffDoesNotChangeReach` |
| 41 | Luyện Khí + Cương Kiếm vs Trúc Cơ | PASS | `TestHardBuffMakesLowRealmTheHarderSide` (vế "phá rào nhanh gấp đôi" lỗi thời) |
| 42 | loadtest 40 nv, 10 Đạo Tổ: tick, snapshot, FPS ≥ 60 PC tầm trung, ping | CẦN MÁY THẬT | tick/snapshot đạt (PERF_LOG S1); FPS cần PC tầm trung thật (mới đo RTX 3060) |
| 43 | Không còn Trường Kiếm | PASS | `config/game.json:180-190`; `TestNoBuffExtendsTheBlade` |
| 44 | Shift ×1.8, 8%/s, không rớt cảnh giới | PASS | `TestBurnSpeedsUpAndCostsExp`, `TestBurnStopsAtRealmFloor` |
| 45 | Đài nhỏ 5s/30s, Thiên Tiên Đài 10s / 150 / 150s | PASS | `TestSmallAltarSpraysOnSchedule`, `TestCenterAltarGiantOrbAndWarning` |
| 46 | mapgen 10 seed liên thông, 5 Đài, 4 sân, đường trống | PASS | `TestNoObjectOutsideMapOverTenSeeds`, `TestEveryMapHasFiveAltars` |
| 47 | `?debug=true` + `B` thả bot | PASS | `TestDebugAllowedWithAdminSession`; `main.ts:55` |
| 48 | Luyện Khí 1 < 10; Đại Thừa 10 scale 5 | PASS | `TestScaleRange` |
| 49 | Kiếm dài + vòng rộng theo scale | PASS | `TestReachGrowsWithRingsAndOuterTipsAreFarther` |
| 50 | 1→10 kiếm cùng bán kính, kiếm 11 mở vòng 2 | PASS | `TestRingRadiusIndependentOfSwordCount` |
| 51 | 10/30/60 kiếm = 1/2/3 vòng, so le, màu, nghiêng lật | PASS | `TestRingsForSwordCount`, `TestSlotAngleStagger`, `TestTiltFlipsWithDirection` |
| 52 | Vòng 3 xa hơn; hitbox khớp hình | PASS | `TestReachGrowsWithRingsAndOuterTipsAreFarther`; debug capsule `swordring.ts:104` |
| 53 | 1 vòng/s, 0.625s với Tốc Kiếm | PASS | `TestOneRevolutionPerRevSecs` |
| 54 | Đi sát mép trên đá, bị che phía sau | PASS | `TestFootprintBlocksOnlyTheBase` |
| 55 | Farm Luyện Khí → Trúc Cơ 30–60s | LỖI THỜI | §6.0 nay là ~1,3 phút (CI `sim --farm`) |
| 56 | Cùng cảnh giới: mỗi bên −1, cả hai đổi chiều | PASS | `TestClashFlipsBothRings` |
| 57 | Trúc Cơ vs Luyện Khí 12:10, không đổi chiều | PASS | `TestDurabilityTwelveClashesTradeTwelveForTen`, `TestClashAcrossRealmsDoesNotFlip` |
| 58 | Nguyên Anh vs Luyện Khí ≈ 1.73 | PASS | `TestResolveClashTable`, `TestDurabilityRatio` |
| 59 | 0 kiếm vẫn sống | PASS | `TestRingCanBeStrippedToZero` (vế "rớt cảnh giới" lỗi thời) |
| 60 | Viên quái chỉ người chém nhặt 1.5s | PASS | `TestMonsterOrbsReservedForKiller` |
| 61 | Gió 90° không đổi tốc | PASS | `TestWindAngles` |
| 62 | Vật cản vỡ → đi qua → 30s mọc lại, không kẹt | PASS | `TestObstacleRegrowsWhenClear` |
| 63 | Chém vật cản không mất kiếm | PASS | `TestCutObstacleWhenRealmEnough` |
| 64 | Rào tre ≈ 15s, đá tảng ≈ 9s | LỖI THỜI | v6.7 bỏ HP, chém là mất ngay |
| 65 | Cổng: ra cổng khác, lóe trước, không bất tử | PASS | `TestGateTeleportsWithoutInvulnerability`; `config/game.json:325` |
| 66 | Người to không lọt hẻm 150 | PASS | `TestNarrowAlleyBlocksBigPlayers` |
| 67 | Buff hết hạn đúng, nhặt lại làm mới | PASS | `TestBuffExpiresAndRefreshes` |
| 68 | Ẩn Thân 20%, không tên, 5s, chém là lộ | PASS | `TestSnapshotSendsStealthFlaggedWithoutName`, `TestStealthCappedAtMaxSecs`, `TestStealthEndsWhenOwnerHitsSomeone` |
| 69 | Trúc Cơ chém Nguyên Anh +850; lại trong 60s +0 | PASS | `TestScoreUpsetExample`, `TestScoreSameVictimCooldown` |
| 70 | Thần Tốc chỉ mạng đã Đại Thừa | PASS | `TestSpeedBoardOnlyReachedDaiThua` |
| 71 | Ranh giới Linh/Tiên khí, ngưỡng 1 / 1 001 / 10 001, mờ 30% | PASS | `TestTienEntryNeedsFullBar`, `TestTienKhiPromotesAtThresholds`, `TestOrbsOnlyGoToTheRightCultivator` |
| 72 | Đột phá Kim Tiên: rồng xoắn 1.5 vòng, lượn thường trực | PASS | `client/src/render/beast.ts` (xem được bằng debug `R` trên PC) |
| 73 | Chân Tiên 800 bị chém rơi 240 | PASS | `TestSplitTienOrbsChecklistExample`, `TestMortalKillerStillDropsTien` |
| 74 | Đạo Tổ nạp / đổi / đốt 40/s | PASS | `TestDaoToFillsThenConverts`, `TestTienBurnRatePerRealm` |
| 75 | Không vùng chữ nhật/tròn, mép bùn khớp đa giác | PASS | `TestPolygonZoneUsesTheDrawnShape`, `TestGeneratedMapsHaveOrganicZones` |
| 76 | HUD chân dung, viền màu, biểu cảm | PASS | `client/src/ui/hud.ts:128,362` |
| 77 | Bảng tên rõ, tên mình vàng ▼, ▲ đỏ, Top 1 ruy băng | PASS | `client/src/render/nameplate.ts:118,178` |
| 78 | Danh hiệu Top 1/2/3 trên đầu + bảng, giữ 3s | PASS | `TestTitleNeedsHold`, `TestBoardOrderAndTitlesNeverDisagree` |
| 79 | Minimap: mình, camera, cổng, Top 1, rương; Top 1 vào bụi thì ẩn | PASS | `client/src/ui/minimap.ts:248-294` |
| 80 | Phòng 5 người → 15 bot; người thứ 21 sang phòng mới | LỖI THỜI | §9.1 nay: trần 50/20 + 10 nhân vật/khu |
| 81 | Leaderboard 3 bảng sau khi chết, không bot | PASS | `TestDeathRecordsHumanOnlyAndFreesSlot`, `TestTopByBoardAndPeriod` |

---

## 5. Mâu thuẫn trong spec (chủ dự án sửa)

GAME_SPEC:
1. Dòng 1–5: tiêu đề còn "(v2)", "đặt tại `docs/GAME_SPEC.md`", nhắc `SETUP_GUIDE.md`. Nội dung thật là v6.30.
2. **Luật "rớt cảnh giới" đã bỏ từ v6.26 nhưng còn ở**: §1 dòng 137 ("bị chém thì rớt tầng"), §3.2 dòng 200 ("nhận 50%… mất nguyên 1 cảnh giới"), §3.4 dòng 262 và §3.5 dòng 291 (hào quang/thần thú khi rớt cảnh giới), §5.2 dòng 430 và 452 ("Trúc Cơ trở lên rớt 1 cảnh giới"), §6.3 dòng 545/549, §11.4 dòng 1038 ("rớt cảnh giới hoặc chết", "rớt xuống {realm}"), §12 dòng 1054 (anim `realm_drop`), §15 M2, §16 `hit_drops_realm: true`, §17 dòng 1637.
3. §16 còn số cũ: `tien_rob_share 0.5` (§3.4: 30%), `burn_speed_mult 1.8`, `haste.mult 1.4` (v6.25 đã chuyển sang cộng), `magnet 1.5`, `cell 140`.
4. §6.1 dòng 518 "nhịp farm chỉ tham khảo, không phải mục tiêu cân bằng" ↔ §6.0 dòng 484 "bảng nhịp là CHUẨN" + CI fail. §17 dòng 1633 "30–60s" ↔ §6.0 "~1,3 phút".
5. §8.0 dòng 635–636 ("vách đá / rặng tre không phá được") và §8.3 dòng 781 ("hẻm giữa vật cản không phá được") ↔ §8.2 dòng 704 ("không còn vật cản nào không phá được").
6. §8.1 dòng 694 "trống ≥ 65%" ↔ §8.1b dòng 759 "≥ 60%".
7. §9.3 dòng 879 "bot chết → hồi sinh sau 3s cùng khu" ↔ §9.4 dòng 920 "70% hồi sinh sau 4–12s, vào khu ít người".
8. §6.3 dòng 605 "Bà Tám dùng Loa khi địch trong nón" ↔ v6.30 Loa toả tròn 360°.
9. §4.1b dòng 384 "dùng ParticleContainer cho kiếm" ↔ §13.6 #1 dòng 1278 "mỗi vòng 1 RenderTexture".
10. §13.6 dòng 1263 mục tiêu "phòng đầy 40 nhân vật" ↔ §9.1 tới 50 người thật + 10 nhân vật/khu.
11. §14 dòng 1365 "hiện số HP trên người" ↔ §12.2 thanh máu; dòng 1374 "150 quái/phòng" viết từ thời 1 bản đồ.
12. §13.2 dòng 1215 cây thư mục `kiem-xoay/`, bản đồ `lang_tre.tmj…` ↔ §20 "không còn chỗ nào dùng kiem-xoay".
13. §15 M3 "1 tab = 1 phòng có 19 bot"; §17 dòng 1609 "tổng 20/phòng", dòng 1658 "5 người → 15 bot, người 21 sang phòng mới" ↔ §9.1 v6.28.
14. §17 dòng 1618 "phá rào tre nhanh gấp đôi", dòng 1642 "rào ≈ 15s, đá ≈ 9s" ↔ v6.7 bỏ HP.

ADMIN_SPEC:
15. Dòng 11 "không thêm framework UI" ↔ §2.1 "React + Ant Design (antd v5) — đã duyệt".
16. §0 mục 1 "xem trực tiếp người đang chơi trong khu" ↔ §5 v1.2 "không vẽ người chơi trực tiếp".

---

## 6b. Cập nhật 02/11/2026 — rà lại Top 10 so với code hiện tại

Bản rà này chấm theo spec **v6.30**; spec nay đã là **v6.36**. Rà lại từng mục của Top 10 bên
dưới so với `main` (HEAD `7c230f9`):

| # | Việc | Nay |
|---|---|---|
| 1 | Shipper bị choáng không Gom Đơn được | **XONG** — `skills.go:141` miễn trừ `p.Char == "shipper"` |
| 2 | Ngủ gục vẫn dùng được kỹ năng / Thiêu Đốt | **XONG** — `actionLocked()` gồm cả `StSleep`; `movement.go:24-25` chặn Thiêu Đốt |
| 3 | Qua cổng 2s thay vì 0.8s | **CHẤM SAI** — xem dưới |
| 4 | Lớp mặt "một thân nhiều mặt" (§12.5 A) | **CHƯA** — không file client nào đọc `faces/`. Hoãn có chủ ý (Q42): cần art mặt cho 7 nhân vật còn lại |
| 5 | Đạo Tổ trên minimap cả phòng | **XONG** — `minimap.ts:283` chấm tím ☯ |
| 6 | 1 anim lặp vượt ngưỡng giật | **CHƯA** — `JITTER_REPORT.md` vẫn "Còn đỏ: 1" |
| 7 | Không có font dự phòng cho ký tự ngoài bộ | **XONG** — `fonts.ts:65-69`, Pixi 8 `DynamicBitmapFont` tự sinh glyph, rơi về font hệ thống |
| 8 | Web Worker giải mã + vòng kiếm RenderTexture | **CHƯA** — không có `new Worker`, không có `RenderTexture` trong `render/`. Netcode bước 2/4/5/7, số đo hiện vẫn trong ngân sách |
| 9 | Minimap 96 px · VFX thiếu filter | **nửa CHẤM SAI, nửa XONG** — xem dưới |
| 10 | Admin dashboard thiếu nửa chỉ số | **PHẦN LỚN XONG** — 12 endpoint `/api/admin/*`, 10 trang React; còn thiếu `/api/admin/ws` và CPU/KB-s theo thiết bị |

**Hai mục chấm sai, sửa ở đây để lần rà sau không lặp lại:**

- **Mục 3 — không phải khoảng thiếu.** GAME_SPEC dòng 533 ghi `gates.transit_secs` = **2.0**,
  đúng bằng `config/game.json`. Con số 0.8 trong bản rà là `cultivation.breakthrough_secs`
  (thời gian đột phá cảnh giới), một khoá khác hẳn. Hai số nằm gần nhau trong tệp config và
  bị đọc nhầm sang nhau.
- **Mục 9 (phần minimap) — code không sai, SPEC tự mâu thuẫn.** Dòng 534 ghi 96 px, dòng 1585
  ghi 120. Code theo 96. Đã ghi thành **Q43** trong `docs/QUESTIONS.md` để chủ dự án chốt một
  số rồi sửa dòng còn lại.
- **Mục 9 (phần VFX) — XONG.** `pixi-filters` và `revolt-fx` đã cài và đang dùng trong
  `render/breakthrough.ts` và `render/vfxlayer.ts`.

Còn lại thật sự: **mục 4, 6, 8** và phần đuôi của mục 10. Cả bốn đều đã nằm trong
"Lỗi / hạn chế đã biết" của `docs/PROGRESS.md`, không có việc nào bị bỏ quên.

Yêu cầu MỚI của v6.31–v6.36 (ngoài phạm vi bản rà này) đều đã làm: luật cộng `Stat()` dùng
chung (v6.31–33), tên bot đa quốc gia (v6.34), sàn bot theo khu (v6.35), đủ 5 lớp hiển thị
kỹ năng + icon trạng thái (v6.36).

## 6. Top 10 việc FAIL / CHƯA quan trọng nhất (theo mức người chơi thấy)

1. **Shipper bị choáng không Gom Đơn được** — `server/internal/game/skills.go:135` chặn mọi kỹ năng khi choáng; spec §6.3/§17 nói đang trói/ngủ/choáng vẫn dùng được và thoát ngay. Thiếu cả test.
2. **Người đang ngủ (Ru Ngủ) vẫn dùng được kỹ năng và vẫn bị trừ Linh khí khi giữ Thiêu Đốt** — `CanCastSkill` (`skills.go:133-135`) và `Burning` (`movement.go:115-121`) chỉ kiểm choáng, không kiểm `StSleep`.
3. **Qua cổng mất 2s thay vì 0.8s** (`config/game.json:324`) — mỗi lần qua cổng/Thiên Môn người chơi đứng chờ lâu gấp 2,5 lần.
4. **Lớp mặt "một thân nhiều mặt" chưa ghép ở client** (§12.5 A, §17 #30) — đổi mood vẫn đổi cả bộ frame.
5. **Đạo Tổ không hiện trên minimap cả phòng** (§3.4) — mất cơ chế "cả phòng hợp sức săn".
6. **Còn 1 anim lặp vượt ngưỡng giật** (`docs/JITTER_REPORT.md`, §17 #31).
7. **Không có font dự phòng cho tên ngoài bộ BitmapFont** (§9.4, §12.3) — tên có ký tự lạ có thể ra ô vuông.
8. **Chưa giải mã gói trong Web Worker; vòng kiếm chưa vẽ thành RenderTexture** (§13.6 #1, #6) — ảnh hưởng FPS máy yếu/điện thoại.
9. **Minimap điện thoại 96 px thay vì 120** (`config/game.json:344`) và VFX §3.6 thiếu GlowFilter/GodrayFilter/revolt-fx.
10. **Admin: dashboard thiếu nửa số chỉ số §3 (CPU, KB/s, ping, FPS, thiết bị…), không có `/api/admin/ws`, cờ nghi hack không lưu DB** — chủ dự án thấy, người chơi không thấy.
