// Danh sách cấm — ADMIN_SPEC §6.2. Cấm theo `guest_id` và IP cùng lúc; gỡ là gỡ cả hai.
import { Alert, Button, Descriptions, Drawer, Empty, Input, Space, Table, Tag } from 'antd';
import { useState, type ReactNode } from 'react';

import { post } from '../api';
import { fullTime, includesCI } from '../format';
import { useFetch } from '../hooks';
import type { Ban } from '../types';
import { clickRow, RelTime, tableBase, useAsk, useRun } from '../ui';

function UnbanButton({ b, onDone }: { b: Ban; onDone: () => void }): ReactNode {
  const ask = useAsk();
  const run = useRun();
  const unban = async (): Promise<void> => {
    const reason = await ask({ title: `Gỡ cấm ${b.guest_id.slice(0, 12)}?`, content: `Lý do cấm: ${b.reason}`, okText: 'Gỡ cấm', danger: true });
    if (reason !== null && (await run(() => post('/api/admin/unban', { guest_id: b.guest_id, reason: reason || 'gỡ thủ công' }), 'Đã gỡ cấm.'))) {
      onDone();
    }
  };
  return (
    <Button
      size="small"
      onClick={(e) => {
        e.stopPropagation();
        void unban();
      }}
    >
      Gỡ cấm
    </Button>
  );
}

const until = (b: Ban): ReactNode => (b.until ? <RelTime at={b.until} /> : <Tag color="red">vĩnh viễn</Tag>);

export function BansPage(): ReactNode {
  const { data, error, loading, reload } = useFetch<Ban[]>('/api/admin/bans');
  const [q, setQ] = useState('');
  const [open, setOpen] = useState<Ban | null>(null);
  const rows = (data ?? []).filter((b) => !q || includesCI(b.guest_id, q) || includesCI(b.ip, q) || includesCI(b.reason, q));
  return (
    <div>
      <Alert type="info" showIcon style={{ marginBottom: 12 }} message="Cấm theo guest_id và IP cùng lúc. Gỡ cấm thì gỡ cả hai." />
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Space style={{ marginBottom: 12 }}>
        <Input.Search allowClear placeholder="Tìm guest_id, IP, lý do..." value={q} onChange={(e) => setQ(e.target.value)} style={{ width: 260 }} />
      </Space>
      <Table<Ban>
        {...tableBase<Ban>()}
        rowKey="id"
        loading={loading && !data}
        dataSource={rows}
        onRow={clickRow(setOpen)}
        locale={{ emptyText: <Empty description="Chưa cấm ai." /> }}
        columns={[
          { title: 'Từ', dataIndex: 'from', render: (v: number) => <RelTime at={v} />, sorter: (a, b) => a.from - b.from, defaultSortOrder: 'descend' },
          { title: 'guest_id', dataIndex: 'guest_id', render: (v: string) => v.slice(0, 12) || '—' },
          { title: 'IP', dataIndex: 'ip', render: (v: string) => v || '—' },
          { title: 'Hết hạn', key: 'until', render: (_, b) => until(b), sorter: (a, b) => (a.until ?? Infinity) - (b.until ?? Infinity) },
          { title: 'Lý do', dataIndex: 'reason' },
          { title: '', key: 'x', render: (_, b) => <UnbanButton b={b} onDone={reload} /> },
        ]}
      />
      <Drawer open={!!open} onClose={() => setOpen(null)} title="Chi tiết lệnh cấm" width={Math.min(520, window.innerWidth)}>
        {open ? (
          <Descriptions size="small" column={1} bordered items={[
            { key: 'g', label: 'guest_id', children: open.guest_id || '—' },
            { key: 'ip', label: 'IP', children: open.ip || '—' },
            { key: 'f', label: 'Từ', children: fullTime(open.from) },
            { key: 'u', label: 'Hết hạn', children: open.until ? fullTime(open.until) : 'vĩnh viễn' },
            { key: 'r', label: 'Lý do', children: open.reason },
          ]} />
        ) : null}
      </Drawer>
    </div>
  );
}
