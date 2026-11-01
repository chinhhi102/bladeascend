// Người chơi thật đang online — ADMIN_SPEC §6. Cờ nghi hack chỉ là DẤU HIỆU; server không
// bao giờ tự kick. Mọi thao tác ở đây do admin bấm, có hộp xác nhận và lý do.
import { Alert, Button, Descriptions, Drawer, Empty, Input, List, Select, Space, Switch, Table, Tag, Typography } from 'antd';
import { useState, type ReactNode } from 'react';

import { post } from '../api';
import { BAN_SPANS, duration, includesCI, int } from '../format';
import { go, useFetch } from '../hooks';
import type { Flag, Player } from '../types';
import { clickRow, RelTime, tableBase, useAsk, useRun } from '../ui';

const flagsOf = (p: Player): Flag[] => p.flags ?? [];

function KickBox({ p, onDone }: { p: Player; onDone: () => void }): ReactNode {
  const [span, setSpan] = useState(0);
  const ask = useAsk();
  const run = useRun();
  const what = BAN_SPANS.find((s) => s.secs === span)?.label ?? '';
  const kick = async (): Promise<void> => {
    const reason = await ask({
      title: what,
      content: `Người chơi: ${p.name || '(chưa vào trận)'} (${p.guest_id.slice(0, 8)})`,
      okText: span === 0 ? 'Mời ra' : 'Mời ra + cấm',
      danger: true,
      required: true,
      hint: 'Lý do (bắt buộc — NGƯỜI CHƠI SẼ ĐỌC ĐƯỢC câu này):',
    });
    if (reason !== null && (await run(() => post('/api/admin/kick', { guest_id: p.guest_id, reason, ban_secs: span }), span === 0 ? 'Đã mời ra.' : 'Đã mời ra và cấm.'))) {
      onDone();
    }
  };
  return (
    <Space wrap>
      <Select value={span} onChange={setSpan} style={{ width: 240 }} options={BAN_SPANS.map((s) => ({ value: s.secs, label: s.label }))} />
      <Button danger type="primary" onClick={() => void kick()}>{span === 0 ? 'Mời ra' : 'Mời ra + cấm'}</Button>
    </Space>
  );
}

function FlagList({ p, onDone }: { p: Player; onDone: () => void }): ReactNode {
  const ask = useAsk();
  const run = useRun();
  const clear = async (f: Flag): Promise<void> => {
    const reason = await ask({ title: `Bỏ cờ "${f.name}"?`, content: 'Đánh dấu nhầm — bỏ cờ này đi.', okText: 'Bỏ cờ' });
    if (reason !== null && (await run(() => post('/api/admin/clearflag', { guest_id: p.guest_id, kind: f.kind, reason }), 'Đã bỏ cờ.'))) {
      onDone();
    }
  };
  if (!flagsOf(p).length) {
    return <Typography.Text type="secondary">Không có cờ nào.</Typography.Text>;
  }
  return (
    <List
      size="small"
      dataSource={flagsOf(p)}
      renderItem={(f) => (
        <List.Item actions={[<Button key="x" size="small" onClick={() => void clear(f)}>Bỏ cờ</Button>]}>
          <List.Item.Meta title={<Tag color="warning">{`⚠ ${f.name}${f.count > 1 ? ` ×${f.count}` : ''}`}</Tag>} description={<><span>{f.evidence}</span> · <RelTime at={f.at} /></>} />
        </List.Item>
      )}
    />
  );
}

function PlayerDrawer({ p, onClose, onDone }: { p: Player | null; onClose: () => void; onDone: () => void }): ReactNode {
  return (
    <Drawer open={!!p} onClose={onClose} width={Math.min(600, window.innerWidth)} title={p ? p.name || '(chưa vào trận)' : ''}>
      {p ? (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          {p.debug_source ? <Alert type="info" showIcon message={`Kết nối có quyền lệnh test: ${p.debug_source}${p.debug_used ? ' (đã dùng)' : ''}`} /> : null}
          <Descriptions size="small" column={1} bordered items={[
            { key: 'gid', label: 'guest_id', children: p.guest_id },
            { key: 'char', label: 'Nhân vật', children: p.character || '?' },
            { key: 'dev', label: 'Thiết bị', children: p.device },
            { key: 'ip', label: 'IP', children: p.ip_mask },
            { key: 'zone', label: 'Phòng / khu', children: `${p.room} / ${p.zone || '—'}` },
            { key: 'realm', label: 'Cảnh giới', children: `${p.realm || '—'} (tầng ${p.level})` },
            { key: 'sw', label: 'Kiếm', children: p.swords },
            { key: 'score', label: 'Uy Danh', children: int(p.score) },
            { key: 'hits', label: 'Chém trúng', children: p.hits },
            { key: 'alive', label: 'Sống', children: duration(p.alive_secs) },
            { key: 'ping', label: 'Ping', children: `${Math.round(p.ping_ms)} ms` },
            { key: 'fps', label: 'FPS', children: Math.round(p.fps) },
            { key: 'corr', label: 'Hiệu chỉnh/phút', children: p.corrections },
          ]} />
          <Typography.Title level={5} style={{ margin: 0 }}>Cờ nghi hack</Typography.Title>
          <FlagList p={p} onDone={onDone} />
          <Typography.Title level={5} style={{ margin: 0 }}>Thao tác</Typography.Title>
          <KickBox p={p} onDone={onDone} />
          {p.room ? <Button onClick={() => go('map', p.room)}>Xem bản đồ phòng này</Button> : null}
        </Space>
      ) : null}
    </Drawer>
  );
}

function filterPlayers(rows: Player[], onlyFlag: boolean, q: string): Player[] {
  const needle = q.trim();
  return rows.filter((p) => (!onlyFlag || flagsOf(p).length > 0) && (!needle || includesCI(p.name, needle) || includesCI(p.guest_id, needle)));
}

export function PlayersPage(): ReactNode {
  // Làm mới 2 s CHỈ khi trang này đang mở (trang tự gỡ interval khi rời đi).
  const { data, error, reload } = useFetch<Player[]>('/api/admin/players', 2000);
  const [onlyFlag, setOnlyFlag] = useState(false);
  const [q, setQ] = useState('');
  const [openId, setOpenId] = useState<string | null>(null);
  const all = data ?? [];
  const rows = filterPlayers(all, onlyFlag, q);
  const open = all.find((p) => p.guest_id === openId) ?? null;
  return (
    <div>
      <Alert type="info" showIcon style={{ marginBottom: 12 }} message="Cờ nghi hack chỉ là dấu hiệu, không phải bằng chứng — mạng chập chờn, hai anh em chung nhà, máy yếu nhấp liên tục đều gắn cờ được. Server không bao giờ tự kick ai." />
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Space wrap style={{ marginBottom: 12 }}>
        <Switch checked={onlyFlag} onChange={setOnlyFlag} /> <span>Chỉ người có cờ</span>
        <Input.Search allowClear placeholder="Tìm tên hoặc guest_id..." value={q} onChange={(e) => setQ(e.target.value)} style={{ width: 240 }} />
        <Typography.Text type="secondary">{`${all.length} người online · ${all.filter((p) => flagsOf(p).length).length} có cờ · hiện ${rows.length}`}</Typography.Text>
      </Space>
      <div id="pllist">
        <Table<Player>
          {...tableBase<Player>()}
          rowKey="guest_id"
          dataSource={rows}
          onRow={clickRow((p) => setOpenId(p.guest_id))}
          locale={{ emptyText: <Empty description={all.length ? 'Không ai khớp bộ lọc.' : 'Chưa có người thật nào online.'} /> }}
          columns={[
            { title: 'Tên', dataIndex: 'name', render: (v: string) => v || '(chưa vào trận)', sorter: (a, b) => a.name.localeCompare(b.name) },
            { title: 'Nhân vật', dataIndex: 'character' },
            { title: 'guest_id', dataIndex: 'guest_id', render: (v: string) => v.slice(0, 8) },
            { title: 'Thiết bị', dataIndex: 'device' },
            { title: 'IP', dataIndex: 'ip_mask' },
            { title: 'Phòng / khu', key: 'z', render: (_, p) => `${p.room} / ${p.zone || '—'}` },
            { title: 'Cảnh giới', key: 'lv', sorter: (a, b) => a.level - b.level, render: (_, p) => `${p.realm || '—'} (${p.level})` },
            { title: 'Kiếm', dataIndex: 'swords', align: 'right', sorter: (a, b) => a.swords - b.swords },
            { title: 'Uy Danh', dataIndex: 'score', align: 'right', render: int, sorter: (a, b) => a.score - b.score },
            { title: 'Chém trúng', dataIndex: 'hits', align: 'right', sorter: (a, b) => a.hits - b.hits },
            { title: 'Sống', dataIndex: 'alive_secs', align: 'right', render: duration, sorter: (a, b) => a.alive_secs - b.alive_secs },
            { title: 'Ping', dataIndex: 'ping_ms', align: 'right', render: (v: number) => `${Math.round(v)} ms`, sorter: (a, b) => a.ping_ms - b.ping_ms },
            { title: 'FPS', dataIndex: 'fps', align: 'right', render: (v: number) => Math.round(v), sorter: (a, b) => a.fps - b.fps },
            { title: 'Hiệu chỉnh/ph', dataIndex: 'corrections', align: 'right', sorter: (a, b) => a.corrections - b.corrections },
            { title: 'Cờ', key: 'f', align: 'right', sorter: (a, b) => flagsOf(a).length - flagsOf(b).length, render: (_, p) => (flagsOf(p).length ? <Tag color="warning">{`⚠ ${flagsOf(p).length}`}</Tag> : '') },
          ]}
        />
      </div>
      <PlayerDrawer p={open} onClose={() => setOpenId(null)} onDone={reload} />
    </div>
  );
}
