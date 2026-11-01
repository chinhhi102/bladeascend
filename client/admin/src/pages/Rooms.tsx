// Phòng & Khu — ADMIN_SPEC v1.4 §4: thẻ cấu hình phòng, công tắc bot toàn server, bảng phòng
// (bấm mở 5 khu) kèm chế độ bot từng phòng và nút "Rút bot ngay".
import { Alert, Button, Card, Col, Empty, Form, InputNumber, Row, Segmented, Select, Space, Table, Tag, Typography } from 'antd';
import { useEffect, useMemo, type ReactNode } from 'react';

import { post } from '../api';
import { BOT_LABEL, BOT_MODES, f1, LEVEL_COLOR, tickLevel } from '../format';
import { go, useFetch, useStats } from '../hooks';
import type { BotRoom, BotsResp, RoomSettings, RoomSettingsView, RoomStat, ZoneStat } from '../types';
import { tableBase, useAsk, useRun } from '../ui';

const FIELDS: { key: keyof RoomSettings; label: string; hint: string }[] = [
  { key: 'max_humans', label: 'Trần người thật / phòng', hint: 'Bot không tính. Hạ trần không đuổi ai — phòng đang vượt chỉ ngừng nhận người mới.' },
  { key: 'zone_min_chars', label: 'Sàn nhân vật mỗi khu', hint: 'Bot lấp mỗi khu cho đủ số này khi được phép sinh.' },
  { key: 'bot_auto_off_at', label: 'Tự tắt sinh bot khi người thật ≥', hint: 'Bật lại khi dưới ngưỡng − 5. 0 = không tự tắt.' },
];

function overLoadtest(v: RoomSettingsView, maxHumans: number): boolean {
  return v.loadtested_humans > 0 && maxHumans > v.loadtested_humans;
}

/** RoomSettingsCard: nạp MỘT lần (không làm mới theo nhịp — đang gõ số thì bị ghi đè mất). */
function RoomSettingsCard(): ReactNode {
  const { data, error, reload } = useFetch<RoomSettingsView>('/api/admin/room-settings');
  const [form] = Form.useForm<RoomSettings>();
  const maxHumans = Form.useWatch('max_humans', form) ?? 0;
  const ask = useAsk();
  const run = useRun();
  useEffect(() => {
    if (data) {
      form.setFieldsValue(data.settings);
    }
  }, [data, form]);
  if (!data) {
    return <Card title="Cấu hình phòng" loading={!error}>{error ? <Alert type="error" message={error} /> : null}</Card>;
  }
  const save = async (v: RoomSettings): Promise<void> => {
    const body = { max_humans: Math.round(v.max_humans), zone_min_chars: Math.round(v.zone_min_chars), bot_auto_off_at: Math.round(v.bot_auto_off_at) };
    const reason = await ask({ title: 'Áp dụng cấu hình phòng ngay?', content: `Trần ${body.max_humans} · sàn ${body.zone_min_chars}/khu · tắt bot ở ${body.bot_auto_off_at}.`, okText: 'Lưu và áp dụng' });
    if (reason !== null && (await run(() => post('/api/admin/room-settings', { ...body, reason }), 'Đã lưu cấu hình phòng.'))) {
      reload();
    }
  };
  const d = data.defaults;
  const warn = data.warn || overLoadtest(data, maxHumans);
  return (
    <Card title="Cấu hình phòng" size="small">
      <Typography.Paragraph type="secondary">
        {`Môi trường: ${data.env} · mặc định ${d.max_humans} / ${d.zone_min_chars} / ${d.bot_auto_off_at} · đã qua loadtest: ${data.loadtested_humans} người/phòng`}
      </Typography.Paragraph>
      <Form form={form} layout="vertical" onFinish={(v) => void save(v)}>
        <Row gutter={16}>
          {FIELDS.map((f) => (
            <Col key={f.key} xs={24} md={8}>
              <Form.Item name={f.key} label={f.label} extra={f.hint} rules={[{ required: true, message: 'Nhập một số' }]}>
                <InputNumber min={0} precision={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          ))}
        </Row>
        {warn ? <Alert type="warning" showIcon style={{ marginBottom: 12 }} message={`Cao hơn mức đã qua loadtest của máy này (${data.loadtested_humans} người) — tick có thể vượt 4 ms khi đông. Vẫn lưu được.`} /> : null}
        <Button type="primary" htmlType="submit">Lưu</Button>
      </Form>
    </Card>
  );
}

function botEffective(r: BotRoom): string {
  if (!r.ok) {
    return '(phòng không trả lời)';
  }
  const base = BOT_LABEL[r.effective] ?? r.effective;
  if (r.effective !== 'auto') {
    return base;
  }
  return r.auto_off ? `${base} — đang tắt vì ${r.humans} người` : `${base} — đang bật`;
}

function GlobalBots({ mode, onDone }: { mode: string; onDone: () => void }): ReactNode {
  const ask = useAsk();
  const run = useRun();
  const pick = async (m: string): Promise<void> => {
    const label = BOT_LABEL[m] ?? m;
    const reason = await ask({ title: `Đổi chế độ bot TOÀN SERVER sang "${label}"?`, okText: 'Đổi' });
    if (reason !== null && (await run(() => post('/api/admin/bots/mode', { room: '', mode: m, reason }), `Bot toàn server: ${label}.`))) {
      onDone();
    }
  };
  return (
    <Space wrap>
      <Typography.Text strong>Bot toàn server:</Typography.Text>
      <Segmented value={mode} options={BOT_MODES} onChange={(v) => void pick(String(v))} />
    </Space>
  );
}

interface RoomRow extends RoomStat {
  bot?: BotRoom;
}

function RoomBotControls({ r, onDone }: { r: RoomRow; onDone: () => void }): ReactNode {
  const ask = useAsk();
  const run = useRun();
  const setMode = async (m: string): Promise<void> => {
    const label = BOT_LABEL[m] ?? m;
    const reason = await ask({ title: `Đổi chế độ bot phòng ${r.id} sang "${label}"?`, okText: 'Đổi' });
    if (reason !== null && (await run(() => post('/api/admin/bots/mode', { room: r.id, mode: m, reason }), `Phòng ${r.id}: ${label}.`))) {
      onDone();
    }
  };
  const withdraw = async (): Promise<void> => {
    const reason = await ask({ title: `Rút bot phòng ${r.id}?`, content: 'Bot không giao tranh trong 5 giây sẽ rời ngay, bot đang đánh rút sau.', okText: 'Rút bot ngay', danger: true });
    if (reason !== null && (await run(() => post('/api/admin/bots/withdraw', { room: r.id, reason }), `Đang rút bot phòng ${r.id}.`))) {
      onDone();
    }
  };
  return (
    <Space wrap onClick={(e) => e.stopPropagation()}>
      <Select size="small" style={{ width: 160 }} value={r.bot?.override ?? ''} onChange={(v) => void setMode(v)} options={[{ value: '', label: 'Theo toàn server' }, ...BOT_MODES]} />
      <Button size="small" danger onClick={() => void withdraw()}>Rút bot ngay</Button>
    </Space>
  );
}

function ZoneTable({ room }: { room: RoomRow }): ReactNode {
  const tick = (v: number): ReactNode => <span style={{ color: tickLevel(v) === 'ok' ? undefined : LEVEL_COLOR[tickLevel(v)] }}>{f1(v)}</span>;
  return (
    <Table<ZoneStat>
      {...tableBase<ZoneStat>()}
      pagination={false}
      rowKey="index"
      dataSource={room.zones ?? []}
      columns={[
        { title: 'Khu', dataIndex: 'name' },
        { title: 'Người / bot / tổng', key: 'n', align: 'right', render: (_, z) => `${z.humans} / ${z.bots} / ${z.humans + z.bots}` },
        { title: 'Tick TB', dataIndex: 'tick_avg', align: 'right', render: tick },
        { title: 'p95', dataIndex: 'tick_p95', align: 'right', render: tick },
        { title: 'Cảnh giới cao nhất', key: 'top', render: (_, z) => (z.top_level >= 0 ? `${z.top_realm} (tầng ${z.top_level})` : '—') },
        { title: 'Quái', dataIndex: 'monsters', align: 'right' },
        { title: 'Vật phẩm trên đất', dataIndex: 'items', align: 'right' },
        { title: 'Vật cản đang chém', dataIndex: 'broken_obstacles', align: 'right' },
        { title: 'Bản đồ', dataIndex: 'map' },
        { title: '', key: 'go', render: (_, z) => <Button size="small" onClick={() => go('map', room.id, z.index)}>Xem bản đồ</Button> },
      ]}
    />
  );
}

function RoomTable({ rows, onDone }: { rows: RoomRow[]; onDone: () => void }): ReactNode {
  return (
    <Table<RoomRow>
      {...tableBase<RoomRow>()}
      rowKey="id"
      dataSource={rows}
      locale={{ emptyText: <Empty description="Chưa có phòng nào" /> }}
      expandable={{ expandedRowRender: (r) => <ZoneTable room={r} />, expandRowByClick: true }}
      columns={[
        { title: 'Phòng', dataIndex: 'id', sorter: (a, b) => a.id.localeCompare(b.id) },
        { title: 'Người / bot / tổng', key: 'n', align: 'right', sorter: (a, b) => a.humans - b.humans, render: (_, r) => `${r.humans} / ${r.bots} / ${r.humans + r.bots}` },
        { title: 'Trần', key: 'max', align: 'right', render: (_, r) => (r.bot ? `giới hạn ${r.bot.max_humans}` : '—') },
        { title: 'Bot đang áp dụng', key: 'eff', render: (_, r) => (r.bot ? <Tag color={r.bot.spawning ? 'green' : 'default'}>{botEffective(r.bot)}</Tag> : '—') },
        { title: 'Chờ rút', key: 'wd', align: 'right', render: (_, r) => r.bot?.withdrawing || '' },
        { title: 'Đổi chế độ bot', key: 'ctl', render: (_, r) => <RoomBotControls r={r} onDone={onDone} /> },
      ]}
    />
  );
}

export function RoomsPage(): ReactNode {
  const { now } = useStats();
  const bots = useFetch<BotsResp>('/api/admin/bots', 2000);
  const botMap = useMemo(() => new Map((bots.data?.rooms ?? []).map((b) => [b.id, b])), [bots.data]);
  const rows: RoomRow[] = (now?.rooms ?? []).map((r) => ({ ...r, bot: botMap.get(r.id) }));
  return (
    <Space direction="vertical" size="middle" style={{ width: '100%' }}>
      <RoomSettingsCard />
      <Card size="small" title="Điều khiển bot">
        <GlobalBots mode={bots.data?.mode ?? ''} onDone={bots.reload} />
        {bots.error ? <Alert type="error" message={bots.error} style={{ marginTop: 8 }} /> : null}
      </Card>
      <Card size="small" title="Phòng (bấm để mở các khu)">
        <RoomTable rows={rows} onDone={bots.reload} />
      </Card>
    </Space>
  );
}
