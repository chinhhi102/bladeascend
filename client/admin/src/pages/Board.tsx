// Bảng xếp hạng — rà soát và ẩn (ADMIN_SPEC §6.2). Chỉ ẨN, không xoá dữ liệu, và sẽ không
// có nút xoá (xem `internal/store/hidden.go`). Bỏ ẩn được bất cứ lúc nào.
import { Alert, Button, Descriptions, Drawer, Empty, Input, Space, Switch, Table, Tag, Typography } from 'antd';
import { useState, type ReactNode } from 'react';

import { post } from '../api';
import { f1, fullTime, includesCI, int } from '../format';
import { useFetch } from '../hooks';
import type { Life } from '../types';
import { clickRow, RelTime, tableBase, useAsk, useRun } from '../ui';

/**
 * SIM_SECS: thời gian chơi THẬT để đạt mốc, đo bằng `cmd/sim`. Cân bằng đổi thì chạy lại sim
 * rồi sửa hai con số này.
 */
const SIM_SECS: Record<number, number> = { 80: 429, 90: 479 };

/** suspectWhy: lý do ngờ một mạng chơi, '' nếu không có gì bất thường. */
export function suspectWhy(r: Life): string {
  const s = Math.max(r.survive_secs, 0.1);
  if (r.peak_level >= 90 && s < SIM_SECS[90] / 2) {
    return `Tiên giới (tầng ${r.peak_level}) trong ${Math.round(s)}s — chơi thật cần ~${SIM_SECS[90]}s`;
  }
  if (r.peak_level >= 80 && s < SIM_SECS[80] / 2) {
    return `Đại Thừa (tầng ${r.peak_level}) trong ${Math.round(s)}s — chơi thật cần ~${SIM_SECS[80]}s`;
  }
  return '';
}

const lifeKey = (r: Life): string => `${r.guest_id}|${r.ended_at}`;

function HideButton({ r, onDone }: { r: Life; onDone: () => void }): ReactNode {
  const ask = useAsk();
  const run = useRun();
  const toggle = async (): Promise<void> => {
    const name = r.name || '(không tên)';
    const reason = r.hidden
      ? await ask({ title: `Bỏ ẩn mạng chơi của "${name}"?`, content: 'Nó sẽ hiện lại trên bảng.', okText: 'Bỏ ẩn' })
      : await ask({ title: `Ẩn mạng chơi của "${name}" khỏi bảng?`, content: 'Chỉ ẩn, không xoá dữ liệu.', okText: 'Ẩn khỏi bảng', danger: true, required: true, initial: suspectWhy(r) });
    if (reason === null) {
      return;
    }
    const ok = await run(() => post('/api/admin/hide', { guest_id: r.guest_id, ended_at: r.ended_at, reason, unhide: r.hidden }), r.hidden ? 'Đã bỏ ẩn.' : 'Đã ẩn khỏi bảng.');
    if (ok) {
      onDone();
    }
  };
  return (
    <Button
      size="small"
      danger={!r.hidden}
      onClick={(e) => {
        e.stopPropagation();
        void toggle();
      }}
    >
      {r.hidden ? 'Bỏ ẩn' : 'Ẩn khỏi bảng'}
    </Button>
  );
}

function why(r: Life): ReactNode {
  if (r.hidden) {
    return <Tag color="default">{`ĐANG ẨN — ${r.hide_reason}`}</Tag>;
  }
  const w = suspectWhy(r);
  return w ? <Tag color="warning">{w}</Tag> : '';
}

function filterLives(rows: Life[], onlySus: boolean, onlyHid: boolean, q: string): Life[] {
  return rows.filter((r) => {
    if (onlySus && !suspectWhy(r) && !r.hidden) {
      return false;
    }
    if (onlyHid && !r.hidden) {
      return false;
    }
    return !q || includesCI(r.name, q) || includesCI(r.guest_id, q);
  });
}

function LifeDrawer({ r, onClose }: { r: Life | null; onClose: () => void }): ReactNode {
  return (
    <Drawer open={!!r} onClose={onClose} title={r ? r.name || '(không tên)' : ''} width={Math.min(520, window.innerWidth)}>
      {r ? (
        <Descriptions size="small" column={1} bordered items={[
          { key: 'g', label: 'guest_id', children: r.guest_id },
          { key: 'c', label: 'Nhân vật', children: r.character },
          { key: 'e', label: 'Kết thúc', children: fullTime(r.ended_at) },
          { key: 's', label: 'Uy Danh', children: int(r.score) },
          { key: 'l', label: 'Tầng cao nhất', children: r.peak_level },
          { key: 'h', label: 'Chém trúng / hạ', children: `${r.hits} / ${r.kills}` },
          { key: 'v', label: 'Sống', children: `${f1(r.survive_secs)} s` },
          { key: 'm', label: 'Bản đồ', children: r.map },
          { key: 'w', label: 'Ghi chú', children: r.hidden ? `ĐANG ẨN — ${r.hide_reason}` : suspectWhy(r) || '—' },
        ]} />
      ) : null}
    </Drawer>
  );
}

export function BoardPage(): ReactNode {
  // 1000 = trần của `RecentLives` — để thấp hơn thì bộ lọc "đang ẩn" lặng lẽ thiếu dòng.
  const { data, error, loading, reload } = useFetch<Life[]>('/api/admin/lives?limit=1000');
  const [onlySus, setOnlySus] = useState(false);
  const [onlyHid, setOnlyHid] = useState(false);
  const [q, setQ] = useState('');
  const [openKey, setOpenKey] = useState('');
  const all = data ?? [];
  const rows = filterLives(all, onlySus, onlyHid, q.trim());
  return (
    <div>
      <Typography.Paragraph type="secondary">
        Chỉ <b>ẩn</b> khỏi bảng, không xoá dữ liệu. Mốc so sánh lấy từ <code>cmd/sim</code>: chơi thật đạt Đại Thừa ~429 giây, Chân Tiên ~479 giây. Danh sách đã rà: <code>docs/LEADERBOARD_AUDIT.md</code>.
      </Typography.Paragraph>
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Space wrap style={{ marginBottom: 12 }}>
        <Switch checked={onlySus} onChange={setOnlySus} /> <span>Chỉ mạng đáng ngờ</span>
        <Switch checked={onlyHid} onChange={setOnlyHid} /> <span>Chỉ mạng đang ẩn</span>
        <Input.Search allowClear placeholder="Tìm tên hoặc guest_id..." value={q} onChange={(e) => setQ(e.target.value)} style={{ width: 240 }} />
        <Typography.Text type="secondary">{`${all.length} mạng gần nhất · ${all.filter((r) => suspectWhy(r)).length} đáng ngờ · ${all.filter((r) => r.hidden).length} đang ẩn · hiện ${rows.length}`}</Typography.Text>
      </Space>
      <div id="bdtbl">
        <Table<Life>
          {...tableBase<Life>()}
          rowKey={lifeKey}
          loading={loading && !data}
          dataSource={rows}
          rowClassName={(r) => [r.hidden ? 'hid' : '', suspectWhy(r) ? 'sus' : ''].filter(Boolean).join(' ')}
          onRow={clickRow((r) => setOpenKey(lifeKey(r)))}
          locale={{ emptyText: <Empty description="Không có dòng nào khớp." /> }}
          columns={[
            { title: 'Kết thúc', dataIndex: 'ended_at', render: (v: number) => <RelTime at={v} />, sorter: (a, b) => a.ended_at - b.ended_at, defaultSortOrder: 'descend' },
            { title: 'Tên', dataIndex: 'name', render: (_, r) => <span>{r.name || '(không tên)'} <Typography.Text type="secondary" style={{ fontSize: 12 }}>{r.guest_id.slice(0, 8)}</Typography.Text></span> },
            { title: 'Uy Danh', dataIndex: 'score', align: 'right', render: int, sorter: (a, b) => a.score - b.score },
            { title: 'Tầng', dataIndex: 'peak_level', align: 'right', sorter: (a, b) => a.peak_level - b.peak_level },
            { title: 'Sống', dataIndex: 'survive_secs', align: 'right', render: (v: number) => `${f1(v)}s`, sorter: (a, b) => a.survive_secs - b.survive_secs },
            { title: 'Vì sao ngờ', key: 'why', render: (_, r) => why(r) },
            { title: '', key: 'x', render: (_, r) => <HideButton r={r} onDone={reload} /> },
          ]}
        />
      </div>
      <LifeDrawer r={all.find((r) => lifeKey(r) === openKey) ?? null} onClose={() => setOpenKey('')} />
    </div>
  );
}
