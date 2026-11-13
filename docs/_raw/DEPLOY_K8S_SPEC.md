# DEPLOY_K8S_SPEC — Triển khai BladeAscend bằng Kubernetes (v1.0)

> Đặt tại `docs/_raw/DEPLOY_K8S_SPEC.md`. Mục tiêu: chuyển cách chạy server từ `tools/deploy.sh` + Docker Compose trên máy dev sang **Kubernetes trên máy chủ riêng**, có CI/CD, sao lưu, giám sát — **không đổi luật chơi, không đổi giao thức**.

---

## 0. Sự thật về kiến trúc hiện tại (phải giữ trong đầu khi thiết kế)

| Đặc điểm | Hệ quả với Kubernetes |
|---|---|
| Phòng chơi nằm **trong bộ nhớ** của 1 tiến trình Go | Thay pod = mọi trận đang chơi kết thúc. Không "rolling update không gián đoạn" được nếu chưa có cơ chế chuyển phòng |
| Dữ liệu ở **SQLite** (1 tệp, 1 bên ghi) + thư mục ảnh góp ý | Chỉ chạy **1 bản sao (replica)**, cần **ổ lưu bền (PVC)**, chiến lược `Recreate` |
| Mã phiên vào lại (§7.1) nằm trong bộ nhớ | Pod khởi động lại → người chơi không vào lại được trận cũ (chấp nhận được, đã ghi trong §13.8) |
| WebSocket dài | Ingress phải bật WebSocket + timeout dài; nếu sau này nhiều pod thì cần sticky theo phòng |
| Job theo giờ (janitor 04:30, rò rỉ 03:00, prune Docker) đang chạy bằng cron của máy | Chuyển thành `CronJob` hoặc bộ hẹn giờ trong ứng dụng |

**Kết luận thiết kế**: giai đoạn đầu là **1 node, 1 replica**, dùng Kubernetes để có triển khai chuẩn hoá, GitOps, TLS tự động, giám sát, sao lưu — **không** để chia tải. Chia tải nhiều pod là giai đoạn 4, chỉ làm khi số người thật cần tới.

---

## 1. Chuẩn bị trong ứng dụng (làm TRƯỚC khi viết manifest)

1. **Cấu hình qua biến môi trường / ConfigMap**: mọi đường dẫn (`DATA_DIR`, `ASSETS_DIR`), cổng, `APP_ENV`, `GOMEMLIMIT`, mức log. `config/game.json` đóng vào image, có thể ghi đè bằng ConfigMap.
2. **Health**: `GET /healthz` (sống: tiến trình trả lời) và `GET /readyz` (sẵn sàng: DB mở được, mọi khu đang tick, không ở trạng thái "đang tắt").
3. **Tắt êm khi nhận SIGTERM** (`terminationGracePeriodSeconds: 60`):
   - Ngừng nhận người vào phòng mới (`/readyz` trả 503).
   - Gửi mọi người chơi thông báo i18n "Máy chủ đang cập nhật, vào lại sau ít giây" (+ màn Tổng kết như hết ván, ghi điểm vào bảng xếp hạng).
   - Ghi xong SQLite (checkpoint WAL), đóng kết nối, thoát trong ≤ 45 s.
4. **Log ra stdout dạng JSON** (thời gian, mức, khu, build, sự kiện) — bỏ ghi tệp log trong container.
5. **Metrics định dạng Prometheus** ở `GET /metrics` (cổng nội bộ, không public): tick theo khu, người/bot, KB/s, heap, goroutine, GC pause, lỗi — dùng lại số của bộ thu gom ADMIN §12 (không đo hai lần).
6. **Image**: multi-stage, chạy **non-root**, `readOnlyRootFilesystem` (chỉ `DATA_DIR` ghi được), build **đa kiến trúc amd64 + arm64** (staging là ARM). Gắn tag = build id (`20261113_1200_<sha>`), không dùng `latest`.
7. **GOMEMLIMIT = 70% `resources.limits.memory`** (đọc từ cgroup hoặc biến môi trường).

---

## 2. Cụm và công cụ (giai đoạn 1 — 1 máy chủ riêng)

| Thành phần | Chọn | Lý do |
|---|---|---|
| Kubernetes | **k3s** (1 node) | Nhẹ (~500 MB RAM), cài 1 lệnh, có sẵn Traefik + local-path storage |
| Ingress | **Traefik** (đi kèm k3s) | Hỗ trợ WebSocket, middleware giới hạn tốc độ |
| TLS | **cert-manager** + Let's Encrypt | Tự cấp/gia hạn chứng chỉ `bladeascend.io`, `staging.…` |
| Lưu trữ | **local-path** PVC (k3s) | Đủ cho 1 node; sao lưu ra ngoài máy (mục 5) |
| Bí mật | **SOPS + age** (mã hoá trong git) hoặc **Sealed Secrets** | `ADMIN_PASSWORD_HASH`, khoá minigen… không bao giờ ở dạng rõ trong repo (luật 14) |
| Đóng gói manifest | **Kustomize**: `deploy/k8s/base` + `overlays/staging`, `overlays/production` | Đơn giản hơn Helm cho 1 ứng dụng |
| GitOps | **Argo CD** (hoặc Flux) | Trạng thái cụm = trạng thái trong git; quay lui = revert commit |
| Registry | **GHCR** (`ghcr.io/chinhhi102/bladeascend`) | Miễn phí với repo công khai, gắn với GitHub Actions |

**Máy chủ tối thiểu** (production): 4 vCPU x86, 8 GB RAM, 80 GB SSD, đặt ở Singapore/Việt Nam. Ngân sách RAM: k3s ~0,6 GB · game ~0,5–1 GB (limit 1,5 GB) · giám sát ~1 GB · còn dư cho hệ điều hành.

---

## 3. Tài nguyên Kubernetes (base)

- `Namespace`: `bladeascend-staging`, `bladeascend-prod`.
- `Deployment` game-server: `replicas: 1`, `strategy: Recreate`; requests `cpu: 1, memory: 512Mi`, limits `cpu: 3, memory: 1536Mi`; probes `/healthz` (liveness, chờ 10 s) + `/readyz` (readiness); `securityContext` non-root, drop ALL capabilities; `preStop` không cần (đã xử lý SIGTERM).
- `PersistentVolumeClaim` `game-data` 10 Gi (SQLite + ảnh góp ý).
- `Service` ClusterIP (cổng game + cổng metrics riêng).
- `Ingress` (Traefik): host theo overlay, TLS từ cert-manager; route `/ws` WebSocket timeout ≥ 1 giờ; `/admin` + `/api/admin` thêm middleware giới hạn tốc độ (lớp phòng thủ thứ hai, sau §1 ADMIN).
- `ConfigMap` `game-config` (ghi đè `game.json` theo môi trường: production 50/30/20, staging 20/15/10 — §9.5).
- `Secret` (mã hoá SOPS/Sealed): `ADMIN_PASSWORD_HASH`, `DEBUG_TOKEN` (chỉ staging), khoá minigen (không đưa vào production).
- `NetworkPolicy`: chỉ Ingress vào cổng game; chỉ Prometheus vào cổng metrics.
- `PodDisruptionBudget` không cần với 1 replica (ghi rõ để người sau không thêm nhầm).

**CronJob** (thay cron của máy):
| Job | Lịch | Ghi chú |
|---|---|---|
| Sao lưu SQLite | 04:00 hằng ngày | `sqlite3 .backup` (hoặc endpoint backup của app) → nén → đẩy ra **ngoài máy** (S3-compatible: Cloudflare R2 / Backblaze B2), giữ 14 bản ngày + 8 bản tuần; kiểm mở được bản sao |
| Janitor (ADMIN §13) | 04:30 | gọi endpoint nội bộ của app (app tự dọn, giữ 1 nơi viết luật) |
| Bài rò rỉ hằng đêm (§13.9) | 03:00 | pod riêng image Playwright, chạy vào **staging**, không vào production; kết quả đẩy về admin |

---

## 4. CI/CD (GitHub Actions)

1. **PR**: test Go (-race), vet, build client, mọi `check:*`, mapgen --check, test đơn vị netcode — **không** cần secret (repo công khai, PR từ fork chạy được).
2. **Merge vào `main`**: build image đa kiến trúc → đẩy GHCR với tag build id → cập nhật tag trong `deploy/k8s/overlays/staging` (commit tự động) → Argo CD sync staging → chạy bài nhanh vào staging (crashtest, rejoin, gomdon-smooth) → đạt thì mở PR "promote to production".
3. **Production**: chỉ đổi khi chủ dự án merge PR promote. Argo CD sync; app tắt êm (§1.3). Ưu tiên deploy giờ vắng — Argo CD **sync window** chỉ cho tự động sync production 02:00–06:00, ngoài khung đó phải bấm tay.
4. **Quay lui**: revert commit tag trong overlay → Argo CD sync bản cũ. Thời gian mục tiêu < 2 phút.
5. `tools/deploy.sh` giữ lại cho máy dev (`:9999`) tới khi production chạy ổn trên k8s ≥ 7 ngày.

---

## 5. Giám sát & log

- **Nhẹ (mặc định)**: **VictoriaMetrics single** + **Grafana** + **Loki** (log) — tổng ~0,7–1 GB RAM. Dashboard Grafana: tick p95 theo khu, người/bot, KB/s, heap/GC, lỗi theo chữ ký, CPU/RAM node, load máy.
- `kube-prometheus-stack` đầy đủ nặng (~1,5–2 GB) — chỉ dùng khi máy ≥ 16 GB.
- **Cảnh báo** (Alertmanager/Grafana → Telegram hoặc email): pod khởi động lại > 2 lần/giờ; tick p95 > 8 ms 5 phút (§ADMIN cảnh báo, kèm nhãn máy bận); RAM pod > 85% limit; ổ PVC > 80%; sao lưu thất bại; chứng chỉ còn < 14 ngày.
- Trang admin **giữ nguyên** (dữ liệu trò chơi, góp ý, người chơi); Grafana lo hạ tầng. Không xây trùng.

---

## 6. Chuyển từ máy dev sang k8s (từng bước, có đường lùi)

1. Dựng k3s trên máy chủ riêng → cài cert-manager, Argo CD, Sealed Secrets/SOPS, giám sát.
2. Deploy **staging** trên k8s; chạy đủ bài kiểm + loadtest 110 nhân vật (luật 16 đo trên máy mới — đây là lần đầu có số tick sạch).
3. Sao lưu SQLite máy dev → khôi phục vào PVC production → đếm số dòng các bảng khớp (như RENAME_STAGING).
4. Đổi DNS `bladeascend.io` sang máy mới (TTL thấp trước 24 giờ). Máy dev giữ chạy song song 7 ngày, chỉ đọc.
5. Ghi `docs/K8S_RUNBOOK.md`: xem log, vào shell pod, khôi phục từ bản sao lưu, quay lui, gia hạn chứng chỉ thủ công, xoay vòng bí mật.

---

## 7. Giai đoạn 4 — nhiều pod (CHƯA làm, chỉ ghi hướng)

Khi một pod không đủ (vd > 200 người thật cùng lúc, hoặc cần không gián đoạn khi deploy):
- Dữ liệu bền chuyển **SQLite → PostgreSQL** (bảng xếp hạng, lượt thích, góp ý, admin, telemetry), mã phiên vào lại → **Redis**.
- Mỗi pod giữ một số phòng; **dịch vụ ghép phòng** (lobby) chọn phòng/pod, client nối thẳng tới pod đó (sticky theo phòng).
- Cân nhắc **Agones** (framework chạy game server trên Kubernetes: cấp phát server, không giết pod đang có trận) thay vì tự viết vòng đời.
- Deploy không gián đoạn: pod cũ ngừng nhận phòng mới, chạy tới khi phòng trống rồi mới tắt.

---

## 8. Checklist

- [ ] `kubectl rollout` production: người chơi thấy thông báo "đang cập nhật", điểm được ghi, vào lại được sau < 60 s.
- [ ] Xoá pod bằng tay → pod mới lên, dữ liệu SQLite còn nguyên, `/readyz` đạt < 30 s.
- [ ] Không bí mật nào ở dạng rõ trong git (CI quét).
- [ ] Sao lưu đêm có mặt ở kho ngoài máy, khôi phục thử thành công trên staging mỗi tháng.
- [ ] Quay lui bằng revert tag < 2 phút.
- [ ] Loadtest 110 nhân vật trên máy mới: tick p95 mỗi khu < 4 ms (lần đo sạch đầu tiên).
