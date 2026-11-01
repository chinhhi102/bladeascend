// Dashboard live — ADMIN_SPEC §3: lưới Card + Statistic, màu theo ngưỡng, mũi tên so với
// ~1 phút trước. Số liệu từ `/api/admin/stats` (2 s, chung cả trang) + chế độ bot toàn server.
import { Alert, Card, Col, Empty, Row, Space, Statistic, Table, Tag, Typography } from 'antd';
import type { ReactNode } from 'react';

import { BOT_LABEL, f1, LEVEL_COLOR, countLevel, realmName, tickLevel, uptime, type Level } from '../format';
import { useFetch, useStats } from '../hooks';
import type { BotsResp, Stats, ZoneStat } from '../types';
import { tableBase } from '../ui';

interface Metric {
  title: string;
  value: number | string;
  /** pick: lấy cùng chỉ số từ gói cũ để so (bỏ trống = không vẽ mũi tên). */
  pick?: (s: Stats) => number;
  precision?: number;
  suffix?: string;
  level?: Level;
}

function Trend({ now, ago }: { now: number; ago: number | undefined }): ReactNode {
  if (ago === undefined || Math.abs(now - ago) < 1e-9) {
    return null;
  }
  const up = now > ago;
  return <span style={{ fontSize: 13, marginLeft: 6, opacity: 0.7 }}>{up ? '▲' : '▼'}</span>;
}

function MetricCard({ m, now, ago }: { m: Metric; now: Stats; ago: Stats | null }): ReactNode {
  const color = m.level && m.level !== 'ok' ? LEVEL_COLOR[m.level] : undefined;
  const cur = m.pick?.(now);
  return (
    <Card size="small" style={color ? { borderColor: color } : undefined}>
      <Statistic
        title={m.title}
        value={m.value}
        precision={m.precision}
        suffix={
          <>
            {m.suffix}
            {cur !== undefined && m.pick ? <Trend now={cur} ago={ago ? m.pick(ago) : undefined} /> : null}
          </>
        }
        valueStyle={color ? { color } : undefined}
      />
    </Card>
  );
}

function Group({ title, items, now, ago }: { title: string; items: Metric[]; now: Stats; ago: Stats | null }): ReactNode {
  return (
    <>
      <Typography.Title level={5} style={{ marginTop: 16 }}>{title}</Typography.Title>
      <Row gutter={[12, 12]}>
        {items.map((m) => (
          <Col key={m.title} xs={12} sm={8} md={6} xl={4}>
            <MetricCard m={m} now={now} ago={ago} />
          </Col>
        ))}
      </Row>
    </>
  );
}

function playerMetrics(s: Stats, botMode: string): Metric[] {
  return [
    { title: 'Người thật online', value: s.players.humans, pick: (x) => x.players.humans },
    { title: 'Bot', value: s.players.bots, pick: (x) => x.players.bots },
    { title: 'Phòng', value: s.players.rooms, pick: (x) => x.players.rooms },
    { title: 'Bot toàn server', value: BOT_LABEL[botMode] ?? (botMode || '—') },
  ];
}

function serverMetrics(s: Stats): Metric[] {
  return [
    { title: 'Tick TB cao nhất', value: s.server.tick_avg_max, precision: 1, suffix: 'ms', level: tickLevel(s.server.tick_avg_max), pick: (x) => x.server.tick_avg_max },
    { title: 'Tick p95 cao nhất', value: s.server.tick_p95_max, precision: 1, suffix: 'ms', level: tickLevel(s.server.tick_p95_max), pick: (x) => x.server.tick_p95_max },
    { title: 'Goroutine', value: s.server.goroutines, pick: (x) => x.server.goroutines },
    { title: 'Heap', value: s.server.heap_mb, precision: 1, suffix: 'MB', pick: (x) => x.server.heap_mb },
    { title: 'Uptime', value: uptime(s.uptime_secs) },
  ];
}

function inboxMetrics(s: Stats): Metric[] {
  return [
    { title: 'Góp ý chưa đọc', value: s.feedback_new, level: countLevel(s.feedback_new), pick: (x) => x.feedback_new },
    { title: 'Người có cờ nghi hack', value: s.flagged, level: countLevel(s.flagged), pick: (x) => x.flagged },
    { title: 'Loại lỗi 24 giờ', value: s.errors_today, level: countLevel(s.errors_today), pick: (x) => x.errors_today },
    { title: 'Người Tiên giới', value: s.tien_count, pick: (x) => x.tien_count },
  ];
}

function Realms({ s }: { s: Stats }): ReactNode {
  const keys = Object.keys(s.realms ?? {}).map(Number).sort((a, b) => a - b);
  if (!keys.length) {
    return <Typography.Text type="secondary">Chưa có người thật nào.</Typography.Text>;
  }
  return (
    <Space wrap>
      {keys.map((r) => (
        <Tag key={r} color={r >= 9 ? 'gold' : 'blue'}>{`${realmName(r)}: ${s.realms[r]}`}</Tag>
      ))}
    </Space>
  );
}

interface ZoneRow extends ZoneStat {
  room: string;
}

function ZoneTicks({ s }: { s: Stats }): ReactNode {
  const rows: ZoneRow[] = (s.rooms ?? []).flatMap((r) => (r.zones ?? []).map((z) => ({ ...z, room: r.id })));
  const tickCell = (v: number): ReactNode => <span style={{ color: tickLevel(v) === 'ok' ? undefined : LEVEL_COLOR[tickLevel(v)] }}>{f1(v)}</span>;
  return (
    <Table<ZoneRow>
      {...tableBase<ZoneRow>()}
      pagination={false}
      rowKey={(z) => `${z.room}/${z.index}`}
      dataSource={rows}
      locale={{ emptyText: <Empty description="Chưa có phòng nào" /> }}
      columns={[
        { title: 'Phòng', dataIndex: 'room' },
        { title: 'Khu', dataIndex: 'name' },
        { title: 'Người', dataIndex: 'humans', align: 'right', sorter: (a, b) => a.humans - b.humans },
        { title: 'Bot', dataIndex: 'bots', align: 'right' },
        { title: 'Tick TB (ms)', dataIndex: 'tick_avg', align: 'right', render: tickCell, sorter: (a, b) => a.tick_avg - b.tick_avg },
        { title: 'Tick p95 (ms)', dataIndex: 'tick_p95', align: 'right', render: tickCell, sorter: (a, b) => a.tick_p95 - b.tick_p95, defaultSortOrder: 'descend' },
      ]}
    />
  );
}

export function DashboardPage(): ReactNode {
  const { now, ago, error } = useStats();
  const bots = useFetch<BotsResp>('/api/admin/bots', 2000);
  if (!now) {
    return error ? <Alert type="error" message={error} /> : <Card loading />;
  }
  return (
    <div>
      {error ? <Alert type="warning" showIcon message={`Mất kết nối một nhịp: ${error}`} style={{ marginBottom: 12 }} /> : null}
      <Typography.Text type="secondary">{`Cập nhật ${new Date(now.at * 1000).toLocaleTimeString('vi-VN')} · mũi tên so với ~1 phút trước`}</Typography.Text>
      <Group title="Người chơi" items={playerMetrics(now, bots.data?.mode ?? '')} now={now} ago={ago} />
      <Group title="Server" items={serverMetrics(now)} now={now} ago={ago} />
      <Group title="Hộp thư & trò chơi" items={inboxMetrics(now)} now={now} ago={ago} />
      <Typography.Title level={5} style={{ marginTop: 16 }}>Phân bố cảnh giới (người thật)</Typography.Title>
      <Realms s={now} />
      <Typography.Title level={5} style={{ marginTop: 16 }}>Tick theo từng khu (đỏ nếu &gt; 4 ms)</Typography.Title>
      <ZoneTicks s={now} />
    </div>
  );
}
