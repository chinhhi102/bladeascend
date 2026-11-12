// Trang Hiệu năng — ADMIN_SPEC §12.3.
//
// Biểu đồ vẽ tay bằng SVG, không thêm thư viện: cả trang cần đúng một kiểu đồ thị (dải
// min–max + đường avg + đường p95 + vạch dọc deploy), mà gói biểu đồ nhỏ nhất cũng kéo theo
// vài trăm KB vào gói admin.
import { Alert, Button, Card, Col, Row, Segmented, Select, Space, Statistic, Table, Tag, Typography } from 'antd';
import { useMemo, useState, type ReactNode } from 'react';

import { f1, fullTime, int } from '../format';
import { useFetch } from '../hooks';
import { tableBase } from '../ui';

interface Point {
  ts: number;
  count: number;
  min: number;
  max: number;
  avg: number;
  p95: number;
  build: string;
}
interface DeployMark {
  TS: number;
  Build: string;
  Note: string;
}
interface SeriesResp {
  points: Point[] | null;
  deploys: DeployMark[] | null;
  budget: number;
}
interface CatalogResp {
  metrics: Record<string, string[]> | null;
  droppedMinutes: number;
  zones: Record<string, string>;
}
interface Alert12 {
  metric: string;
  scope: string;
  message: string;
}
interface SummaryResp {
  alerts: Alert12[] | null;
  cards:
    | {
        metric: string;
        scope: string;
        ts: number;
        p95: number;
        avg: number;
        humans: number;
        load1: number;
      }[]
    | null;
  from: number;
  to: number;
}

/** Tên tiếng Việt + đơn vị của từng chỉ số. Chỉ số lạ thì hiện nguyên tên khoá. */
const LABEL: Record<string, string> = {
  tick_ms: 'Tick (ms)',
  sys_input_ms: 'Hệ thống: input (ms)',
  sys_bots_ms: 'Hệ thống: bot AI (ms)',
  sys_move_ms: 'Hệ thống: di chuyển (ms)',
  sys_skills_ms: 'Hệ thống: kỹ năng (ms)',
  sys_collide_ms: 'Hệ thống: va chạm (ms)',
  sys_world_ms: 'Hệ thống: thế giới (ms)',
  sys_deaths_ms: 'Hệ thống: xử lý chết (ms)',
  sys_send_ms: 'Hệ thống: AOI + gửi gói (ms)',
  humans: 'Người thật',
  bots: 'Bot',
  monsters: 'Quái',
  items: 'Viên trên đất',
  obstacles_broken: 'Vật cản chờ mọc lại',
  swords: 'Kiếm trong khu',
  skills_per_min: 'Lượt kỹ năng/phút',
  deaths_per_min: 'Lượt chết/phút',
  kbps_out: 'KB/s gửi mỗi client',
  heap_mb: 'Heap (MB)',
  goroutines: 'Goroutine',
  gc_pause_ms: 'GC dừng (ms/giây)',
  cpu_pct: 'CPU (%)',
  load1: 'Load máy (1 phút)',
  mem_free_mb: 'RAM còn trống (MB)',
};

const RANGES = [
  { label: '24 giờ', hours: 24, res: 'minute' },
  { label: '7 ngày', hours: 24 * 7, res: 'hour' },
  { label: '30 ngày', hours: 24 * 30, res: 'hour' },
  { label: '180 ngày', hours: 24 * 180, res: 'day' },
];

const W = 1000;
const H = 320;
const PAD = { l: 58, r: 58, t: 14, b: 26 };

interface Scale {
  lo: number;
  hi: number;
}

function scaleOf(pts: Point[], withBand: boolean, budget: number): Scale {
  let lo = Infinity;
  let hi = -Infinity;
  for (const p of pts) {
    lo = Math.min(lo, withBand ? p.min : p.avg);
    hi = Math.max(hi, withBand ? p.max : p.p95);
  }
  if (budget > 0) {
    hi = Math.max(hi, budget);
  }
  if (!Number.isFinite(lo) || !Number.isFinite(hi)) {
    return { lo: 0, hi: 1 };
  }
  if (hi - lo < 1e-9) {
    hi = lo + 1;
  }
  return { lo: Math.min(0, lo), hi: hi * 1.05 };
}

function Chart({
  a,
  b,
  from,
  to,
  budget,
  deploys,
  metricA,
  metricB,
}: {
  a: Point[];
  b: Point[];
  from: number;
  to: number;
  budget: number;
  deploys: DeployMark[];
  metricA: string;
  metricB: string;
}): ReactNode {
  const [hover, setHover] = useState<number | null>(null);
  const sa = scaleOf(a, true, budget);
  const sb = scaleOf(b, false, 0);
  const x = (ts: number): number => PAD.l + ((ts - from) / Math.max(1, to - from)) * (W - PAD.l - PAD.r);
  const y = (v: number, s: Scale): number => H - PAD.b - ((v - s.lo) / (s.hi - s.lo)) * (H - PAD.t - PAD.b);
  const line = (pts: Point[], pick: (p: Point) => number, s: Scale): string =>
    pts.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(p.ts).toFixed(1)},${y(pick(p), s).toFixed(1)}`).join(' ');
  const band =
    a.length > 0
      ? `${a.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(p.ts).toFixed(1)},${y(p.max, sa).toFixed(1)}`).join(' ')} ` +
        `${a
          .slice()
          .reverse()
          .map((p) => `L${x(p.ts).toFixed(1)},${y(p.min, sa).toFixed(1)}`)
          .join(' ')} Z`
      : '';
  const ticksY = [0, 0.25, 0.5, 0.75, 1].map((f) => sa.lo + (sa.hi - sa.lo) * f);
  const ticksX = [0, 0.25, 0.5, 0.75, 1].map((f) => from + (to - from) * f);
  const hv = hover !== null ? a[hover] : null;
  const hb = hover !== null ? b[hover] : null;

  const onMove = (e: React.MouseEvent<SVGSVGElement>): void => {
    if (a.length === 0) {
      return;
    }
    const r = e.currentTarget.getBoundingClientRect();
    const px = ((e.clientX - r.left) / r.width) * W;
    let best = 0;
    for (let i = 1; i < a.length; i++) {
      if (Math.abs(x(a[i].ts) - px) < Math.abs(x(a[best].ts) - px)) {
        best = i;
      }
    }
    setHover(best);
  };

  return (
    <svg viewBox={`0 0 ${W} ${H}`} style={{ width: '100%', height: 'auto' }} onMouseMove={onMove} onMouseLeave={() => setHover(null)}>
      {ticksY.map((v, i) => (
        <g key={i}>
          <line x1={PAD.l} x2={W - PAD.r} y1={y(v, sa)} y2={y(v, sa)} stroke="currentColor" strokeOpacity={0.12} />
          <text x={PAD.l - 6} y={y(v, sa) + 4} textAnchor="end" fontSize={11} fill="currentColor" fillOpacity={0.65}>
            {f1(v)}
          </text>
        </g>
      ))}
      {b.length > 0
        ? ticksY.map((_, i) => {
            const v = sb.lo + ((sb.hi - sb.lo) * i) / 4;
            return (
              <text key={`r${i}`} x={W - PAD.r + 6} y={y(sa.lo + ((sa.hi - sa.lo) * i) / 4, sa) + 4} fontSize={11} fill="#d48806">
                {f1(v)}
              </text>
            );
          })
        : null}
      {ticksX.map((ts, i) => (
        <text key={`x${i}`} x={x(ts)} y={H - 8} textAnchor="middle" fontSize={11} fill="currentColor" fillOpacity={0.65}>
          {new Date(ts * 1000).toLocaleString('vi-VN', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })}
        </text>
      ))}
      {band ? <path d={band} fill="#1677ff" fillOpacity={0.12} /> : null}
      <path d={line(a, (p) => p.avg, sa)} fill="none" stroke="#1677ff" strokeWidth={1.6} />
      <path d={line(a, (p) => p.p95, sa)} fill="none" stroke="#722ed1" strokeWidth={1.4} strokeDasharray="5 3" />
      {b.length > 0 ? <path d={line(b, (p) => p.avg, sb)} fill="none" stroke="#d48806" strokeWidth={1.4} /> : null}
      {budget > 0 ? (
        <>
          <line x1={PAD.l} x2={W - PAD.r} y1={y(budget, sa)} y2={y(budget, sa)} stroke="#ff4d4f" strokeDasharray="6 4" />
          <text x={W - PAD.r} y={y(budget, sa) - 4} textAnchor="end" fontSize={11} fill="#ff4d4f">
            ngân sách {f1(budget)}
          </text>
        </>
      ) : null}
      {deploys.map((d) => (
        <g key={d.TS}>
          <line x1={x(d.TS)} x2={x(d.TS)} y1={PAD.t} y2={H - PAD.b} stroke="#52c41a" strokeOpacity={0.7} strokeDasharray="3 3" />
          <text x={x(d.TS) + 3} y={PAD.t + 10} fontSize={10} fill="#52c41a">
            deploy
          </text>
        </g>
      ))}
      {hv ? (
        <g>
          <line x1={x(hv.ts)} x2={x(hv.ts)} y1={PAD.t} y2={H - PAD.b} stroke="currentColor" strokeOpacity={0.35} />
          <circle cx={x(hv.ts)} cy={y(hv.avg, sa)} r={3} fill="#1677ff" />
          <foreignObject x={Math.min(x(hv.ts) + 8, W - 250)} y={PAD.t} width={240} height={110}>
            <div
              style={{
                font: '11px/1.5 system-ui, sans-serif',
                background: 'rgba(0,0,0,.78)',
                color: '#fff',
                padding: '6px 8px',
                borderRadius: 6,
              }}
            >
              <div>{fullTime(hv.ts)}</div>
              <div>
                {LABEL[metricA] ?? metricA}: TB {f1(hv.avg)} · p95 {f1(hv.p95)} · min {f1(hv.min)} · max {f1(hv.max)}
              </div>
              <div>mẫu: {int(hv.count)}</div>
              {hb ? (
                <div style={{ color: '#ffd666' }}>
                  {LABEL[metricB] ?? metricB}: TB {f1(hb.avg)} · max {f1(hb.max)}
                </div>
              ) : null}
            </div>
          </foreignObject>
        </g>
      ) : null}
    </svg>
  );
}

export function PerfPage(): ReactNode {
  const [rangeIdx, setRangeIdx] = useState(0);
  const [metric, setMetric] = useState('tick_ms');
  const [scope, setScope] = useState('zone:0');
  const [metric2, setMetric2] = useState('');
  const [scope2, setScope2] = useState('machine');
  const [res, setRes] = useState('');

  const range = RANGES[rangeIdx];
  const resolution = res || range.res;
  // Chốt mốc thời gian theo phút: để đổi ô chọn không sinh URL mới mỗi lần render.
  const now = Math.floor(Date.now() / 60000) * 60;
  const from = now - range.hours * 3600;

  const cat = useFetch<CatalogResp>('/api/admin/metrics/catalog');
  const sum = useFetch<SummaryResp>('/api/admin/metrics/summary', 60000);
  const qs = (m: string, s: string): string =>
    `/api/admin/metrics?metric=${encodeURIComponent(m)}&scope=${encodeURIComponent(s)}&res=${resolution}&from=${from}&to=${now}`;
  const a = useFetch<SeriesResp>(qs(metric, scope), 60000);
  const b = useFetch<SeriesResp>(metric2 ? qs(metric2, scope2) : null, 60000);

  const metrics = cat.data?.metrics ?? {};
  const zoneName = (s: string): string => cat.data?.zones?.[s] ?? s;
  const metricOpts = Object.keys(metrics)
    .sort()
    .map((m) => ({ value: m, label: LABEL[m] ?? m }));
  const scopeOpts = (m: string): { value: string; label: string }[] =>
    (metrics[m] ?? []).map((s) => ({ value: s, label: `${zoneName(s)} (${s})` }));

  const pts = a.data?.points ?? [];
  const rows = useMemo<Point[]>(() => pts.map((p) => ({ ...p })), [pts]);
  const csv = `/api/admin/metrics.csv?metric=${encodeURIComponent(metric)}&scope=${encodeURIComponent(scope)}&res=${resolution}&from=${from}&to=${now}`;

  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <Typography.Title level={4} style={{ margin: 0 }}>
        Hiệu năng
      </Typography.Title>

      {(sum.data?.alerts ?? []).map((a) => (
        <Alert key={`${a.metric}:${a.scope}`} type="warning" showIcon message={a.message} />
      ))}

      <Row gutter={[12, 12]}>
        {(sum.data?.cards ?? []).map((c) => (
          <Col key={`${c.metric}:${c.scope}`} xs={12} md={8} lg={6}>
            <Card size="small">
              <Statistic
                title={`${LABEL[c.metric] ?? c.metric} — giờ tệ nhất`}
                value={f1(c.p95)}
                suffix="p95"
                valueStyle={{ color: c.metric === 'tick_ms' && c.p95 > 4 ? '#ff4d4f' : undefined }}
              />
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                {zoneName(c.scope)} · {fullTime(c.ts)} · {int(c.humans)} người · load {f1(c.load1)}
              </Typography.Text>
            </Card>
          </Col>
        ))}
      </Row>

      <Card size="small">
        <Space wrap>
          <Segmented options={RANGES.map((r, i) => ({ label: r.label, value: i }))} value={rangeIdx} onChange={(v) => setRangeIdx(Number(v))} />
          <Segmented
            options={[
              { label: 'phút', value: 'minute' },
              { label: 'giờ', value: 'hour' },
              { label: 'ngày', value: 'day' },
            ]}
            value={resolution}
            onChange={(v) => setRes(String(v))}
          />
          <Select style={{ width: 240 }} options={metricOpts} value={metric} onChange={setMetric} showSearch optionFilterProp="label" />
          <Select style={{ width: 220 }} options={scopeOpts(metric)} value={scope} onChange={setScope} />
          <Select
            style={{ width: 240 }}
            allowClear
            placeholder="chồng thêm chỉ số thứ hai…"
            options={metricOpts}
            value={metric2 || undefined}
            onChange={(v) => setMetric2(v ?? '')}
            showSearch
            optionFilterProp="label"
          />
          {metric2 ? <Select style={{ width: 200 }} options={scopeOpts(metric2)} value={scope2} onChange={setScope2} /> : null}
          <Button href={csv} target="_blank">
            Xuất CSV
          </Button>
        </Space>
      </Card>

      {a.error ? <Alert type="error" message={a.error} /> : null}
      {!a.error && pts.length === 0 ? (
        <Alert
          type="info"
          showIcon
          message="Chưa có số liệu trong khoảng này"
          description={`Bộ thu gom CHỈ ghi những phút có ít nhất một người thật online (ADMIN_SPEC §12.2). Đã bỏ ${int(cat.data?.droppedMinutes ?? 0)} phút không có ai kể từ lần khởi động gần nhất.`}
        />
      ) : null}

      <Card size="small" styles={{ body: { paddingTop: 8 } }}>
        <Space size="large" style={{ marginBottom: 4 }}>
          <Tag color="blue">dải min–max + đường TB</Tag>
          <Tag color="purple">p95</Tag>
          {metric2 ? <Tag color="orange">{LABEL[metric2] ?? metric2} (trục phải)</Tag> : null}
          <Tag color="green">vạch deploy</Tag>
        </Space>
        <Chart
          a={pts}
          b={b.data?.points ?? []}
          from={from}
          to={now}
          budget={a.data?.budget ?? 0}
          deploys={a.data?.deploys ?? []}
          metricA={metric}
          metricB={metric2}
        />
      </Card>

      <Table
        {...tableBase<Point>()}
        rowKey="ts"
        dataSource={rows}
        columns={[
          { title: 'Khoảng', dataIndex: 'ts', render: (v: number) => fullTime(v), sorter: (x: Point, y: Point) => x.ts - y.ts, defaultSortOrder: 'descend' },
          { title: 'Mẫu', dataIndex: 'count', render: int, sorter: (x: Point, y: Point) => x.count - y.count },
          { title: 'Min', dataIndex: 'min', render: f1, sorter: (x: Point, y: Point) => x.min - y.min },
          { title: 'TB', dataIndex: 'avg', render: f1, sorter: (x: Point, y: Point) => x.avg - y.avg },
          { title: 'p95', dataIndex: 'p95', render: f1, sorter: (x: Point, y: Point) => x.p95 - y.p95 },
          { title: 'Max', dataIndex: 'max', render: f1, sorter: (x: Point, y: Point) => x.max - y.max },
          { title: 'Build', dataIndex: 'build', ellipsis: true },
        ]}
      />
    </Space>
  );
}
