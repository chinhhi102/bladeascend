# Runtime image only: the server binary (linux/arm64) and the built client are made on the
# workstation by deploy/staging.sh and rsynced next to this file.
# Telemetry nằm trên volume /data cùng bảng xếp hạng: mặc định server ghi vào
# /app/server/data/telemetry.db — trong container, nên MẤT SẠCH mỗi lần deploy (phát hiện 1/11).
FROM alpine:3.20
WORKDIR /app
COPY bin/server /app/bin/server
COPY config /app/config
COPY maps /app/maps
COPY assets /app/assets
COPY client/dist /app/client/dist
EXPOSE 8080
ENTRYPOINT ["/app/bin/server", "-addr", ":8080", "-config", "/app/config/game.json", "-maps", "/app/maps", \
  "-assets", "/app/assets", "-client", "/app/client/dist", "-db", "/data/leaderboard.db", \
  "-telemetry-db", "/data/telemetry.db"]
