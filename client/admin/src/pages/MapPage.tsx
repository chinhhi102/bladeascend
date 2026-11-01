// Xem bản đồ toàn cảnh — ADMIN_SPEC §5. Chọn phòng → khu (bản đồ khu đó đang chạy) hoặc chọn
// thẳng một bản đồ; lớp phủ mặc định TẮT để nhìn như người chơi; số liệu khu bên cạnh (2 s) —
// chỉ số, không chấm vị trí.
import { Alert, App, Button, Card, Col, Descriptions, Row, Select, Slider, Space, Switch, Typography } from 'antd';
import { useEffect, useMemo, useRef, useState, type ReactNode, type RefObject } from 'react';

import { api } from '../api';
import { f1, realmName } from '../format';
import { go, useFetch, useStats } from '../hooks';
import { MapStage, NO_OVERLAYS, type Overlays } from '../map/mapStage';
import type { GameMap } from '../../../src/net/protocol';
import type { CheckReport, CheckRow, Stats, ZoneStat } from '../types';

interface Choice {
  value: string;
  label: string;
  map: string;
}

/** choices: "phòng · khu" từ gói stats, rồi mọi bản đồ có trong manifest asset. */
function choices(s: Stats | null, manifestZones: string[]): Choice[] {
  const out: Choice[] = [];
  for (const r of s?.rooms ?? []) {
    for (const z of r.zones ?? []) {
      out.push({ value: `z:${r.id}:${z.index}`, label: `${r.id} · ${z.name}`, map: z.map });
    }
  }
  for (const id of manifestZones) {
    out.push({ value: `m:${id}`, label: `Bản đồ ${id}`, map: id });
  }
  return out;
}

function zoneOf(s: Stats | null, value: string): ZoneStat | null {
  const [kind, room, idx] = value.split(':');
  if (kind !== 'z') {
    return null;
  }
  return s?.rooms?.find((r) => r.id === room)?.zones?.find((z) => z.index === Number(idx)) ?? null;
}

function useManifestZones(): string[] {
  const [zones, setZones] = useState<string[]>([]);
  useEffect(() => {
    fetch('/assets/manifest.json', { cache: 'no-cache' })
      .then((r) => r.json() as Promise<{ zones?: string[] }>)
      .then((m) => setZones(m.zones ?? []))
      .catch(() => setZones([]));
  }, []);
  return zones;
}

function useGateRadius(): number {
  const [r, setR] = useState(0);
  useEffect(() => {
    fetch('/api/config')
      .then((x) => x.json() as Promise<{ gates?: { radius?: number } }>)
      .then((c) => setR(c.gates?.radius ?? 260))
      .catch(() => setR(260));
  }, []);
  return r;
}

function ZonePanel({ z }: { z: ZoneStat | null }): ReactNode {
  if (!z) {
    return <Typography.Text type="secondary">Đang xem bản đồ tĩnh — không gắn với khu nào đang chạy.</Typography.Text>;
  }
  return (
    <Descriptions size="small" column={1} items={[
      { key: 'n', label: 'Người thật / bot', children: `${z.humans} / ${z.bots}` },
      { key: 't', label: 'Tick TB / p95', children: `${f1(z.tick_avg)} / ${f1(z.tick_p95)} ms` },
      { key: 'r', label: 'Cảnh giới cao nhất', children: z.top_level >= 0 ? `${z.top_realm} (tầng ${z.top_level})` : '—' },
      { key: 'm', label: 'Quái', children: z.monsters },
      { key: 'i', label: 'Vật phẩm trên đất', children: z.items },
      { key: 'b', label: 'Vật cản đang chém', children: z.broken_obstacles },
    ]} />
  );
}

function CheckLine({ rep, mapId, realm }: { rep: CheckReport | null; mapId: string; realm: number }): ReactNode {
  const r: CheckRow | undefined = rep?.maps?.[mapId]?.[realm];
  if (!r) {
    return <Typography.Text type="secondary">Chưa có báo cáo --check cho bản đồ này (cd server &amp;&amp; go run ./cmd/mapgen --check).</Typography.Text>;
  }
  const want = rep?.targets?.[mapId];
  const pct = (v: number): string => `${(v * 100).toFixed(1)}%`;
  const bad = (cond: boolean): 'danger' | undefined => (cond ? 'danger' : undefined);
  return (
    <Space direction="vertical" size={0}>
      <Typography.Text>{`${r.Name} (r=${r.Realm}), thân ${Math.round(r.BodyR)}`}</Typography.Text>
      <Typography.Text type={bad(!!want && (r.Density > want[1] || r.Density < want[0]))}>{`Mật độ hình vẽ ${pct(r.Density)}${want ? ` (mục tiêu ${pct(want[0])}–${pct(want[1])})` : ''}`}</Typography.Text>
      <Typography.Text type={bad(r.Breathing < 0.2)}>{`Khoảng thở ${pct(r.Breathing)}`}</Typography.Text>
      <Typography.Text type={bad(r.OpenFrac < 0.6)}>{`Đi được ${pct(r.OpenFrac)}`}</Typography.Text>
      <Typography.Text type={bad(r.BiggestTrap >= 4)}>{`Túi kín to nhất ${r.BiggestTrap} ô`}</Typography.Text>
      {r.CutOff?.length ? <Typography.Text type="danger">{`KHÔNG TỚI ĐƯỢC: ${r.CutOff.join(', ')}`}</Typography.Text> : null}
    </Space>
  );
}

function OverlayControls({ ov, setOv }: { ov: Overlays; setOv: (o: Overlays) => void }): ReactNode {
  const sw = (key: 'tiers' | 'grid' | 'zones' | 'motion', label: string): ReactNode => (
    <Space key={key}>
      <Switch size="small" checked={ov[key]} onChange={(v) => setOv({ ...ov, [key]: v })} />
      <span>{label}</span>
    </Space>
  );
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      {sw('tiers', 'Bậc vật cản (tô màu)')}
      {sw('zones', 'Vùng (bụi / bùn / băng / nước)')}
      {sw('grid', 'Lưới 1 000')}
      {sw('motion', 'Chuyển động (nước, mây, đom đóm)')}
      <Space>
        <Switch size="small" checked={ov.realm >= 0} onChange={(v) => setOv({ ...ov, realm: v ? 0 : -1 })} />
        <span>Kết quả --check theo cảnh giới</span>
      </Space>
      {ov.realm >= 0 ? <Slider min={0} max={13} value={ov.realm} onChange={(v) => setOv({ ...ov, realm: v })} tooltip={{ formatter: (v) => `r=${v} ${realmName(v ?? 0)}` }} /> : null}
    </Space>
  );
}

function useStage(host: RefObject<HTMLDivElement>): MapStage | null {
  const [stage, setStage] = useState<MapStage | null>(null);
  useEffect(() => {
    if (!host.current) {
      return undefined;
    }
    const st = new MapStage(host.current);
    let live = true;
    st.init().then(() => live && setStage(st)).catch((e: unknown) => console.warn('[admin] không khởi tạo được bản đồ', e));
    return () => {
      live = false;
      st.destroy();
    };
  }, [host]);
  return stage;
}

function download(blob: Blob, name: string): void {
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = name;
  a.click();
  window.setTimeout(() => URL.revokeObjectURL(a.href), 5000);
}

export function MapPage({ args }: { args: string[] }): ReactNode {
  const { now } = useStats();
  const { message } = App.useApp();
  const manifestZones = useManifestZones();
  const gateR = useGateRadius();
  const opts = useMemo(() => choices(now, manifestZones), [now, manifestZones]);
  const fromArgs = args[0] ? `z:${args[0]}:${args[1] ?? 0}` : '';
  const [picked, setPicked] = useState('');
  const value = picked || fromArgs || opts[0]?.value || '';
  const mapId = opts.find((o) => o.value === value)?.map ?? '';
  const host = useRef<HTMLDivElement>(null);
  const stage = useStage(host);
  const [ov, setOv] = useState<Overlays>(NO_OVERLAYS);
  const [err, setErr] = useState('');
  const [busy, setBusy] = useState(false);
  const [px, setPx] = useState(4096);
  const check = useFetch<CheckReport>(ov.realm >= 0 ? '/api/admin/check' : null);

  useEffect(() => {
    if (!stage || !mapId || !gateR) {
      return;
    }
    setErr('');
    setBusy(true);
    api<GameMap>(`/api/map/${encodeURIComponent(mapId)}`)
      .then((m) => stage.show(m, gateR))
      .catch((e: unknown) => setErr(e instanceof Error ? e.message : String(e)))
      .finally(() => setBusy(false));
  }, [stage, mapId, gateR]);

  useEffect(() => stage?.setOverlays(ov), [stage, ov]);

  const exportPng = async (): Promise<void> => {
    if (!stage) {
      return;
    }
    setBusy(true);
    const size = Math.min(px, stage.maxTextureSize());
    const blob = await stage.exportPng(size).catch(() => null);
    setBusy(false);
    if (!blob) {
      message.error('Không xuất được ảnh (thiếu bộ nhớ đồ hoạ?) — thử cỡ nhỏ hơn.');
      return;
    }
    download(blob, `ban-do-${mapId}-${size}px.png`);
  };

  return (
    <Row gutter={[12, 12]}>
      <Col xs={24} lg={17}>
        <Card size="small" styles={{ body: { padding: 0 } }}>
          <div ref={host} style={{ width: '100%', height: 'min(75vh, 900px)', minHeight: 320, position: 'relative' }} />
        </Card>
        <Typography.Text type="secondary">Kéo để di chuyển · lăn chuột / chụm hai ngón để phóng to.</Typography.Text>
      </Col>
      <Col xs={24} lg={7}>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Card size="small" title="Chọn khu">
            <Select style={{ width: '100%' }} value={value || undefined} placeholder="Chưa có khu nào" options={opts} onChange={(v: string) => { setPicked(v); go('map'); }} showSearch optionFilterProp="label" />
            <Space wrap style={{ marginTop: 12 }}>
              <Button onClick={() => stage?.fit()} disabled={!stage}>Vừa khung</Button>
              <Select value={px} onChange={setPx} style={{ width: 110 }} options={[2048, 4096, 8192].map((v) => ({ value: v, label: `${v} px` }))} />
              <Button type="primary" loading={busy} disabled={!stage || !mapId} onClick={() => void exportPng()}>Xuất PNG</Button>
            </Space>
            {err ? <Alert type="error" message={err} style={{ marginTop: 12 }} /> : null}
          </Card>
          <Card size="small" title="Lớp phủ">
            <OverlayControls ov={ov} setOv={setOv} />
            {ov.realm >= 0 ? <div style={{ marginTop: 8 }}><CheckLine rep={check.data} mapId={mapId} realm={ov.realm} /></div> : null}
          </Card>
          <Card size="small" title="Số liệu khu (2 giây)">
            <ZonePanel z={zoneOf(now, value)} />
          </Card>
          <Typography.Text type="secondary">Xem bố cục theo seed bất kỳ chưa có API phía server — hiện chỉ xem được các bản đồ đã sinh.</Typography.Text>
        </Space>
      </Col>
    </Row>
  );
}
