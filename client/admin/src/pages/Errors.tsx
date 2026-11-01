// Lỗi của người chơi thật — ADMIN_SPEC §3, TELEMETRY_SPEC §4. Gom theo nội dung, xếp theo
// SỐ NGƯỜI DÍNH trước rồi mới tới số lần.
import { Alert, Descriptions, Drawer, Empty, Select, Space, Table, Tag, Typography } from 'antd';
import { useState, type ReactNode } from 'react';

import { ERR_LABEL, fullTime } from '../format';
import { useFetch } from '../hooks';
import type { ErrorGroup, ErrorsResp } from '../types';
import { clickRow, Pre, RelTime, tableBase } from '../ui';

/** shown: `gpu_pressure` chở JSON số liệu — hiện khoá=giá trị; lỗi thường giữ nguyên vết. */
function shown(e: ErrorGroup): string {
  if (e.kind !== 'gpu_pressure') {
    return e.sample || e.sig;
  }
  try {
    return Object.entries(JSON.parse(e.sample) as Record<string, unknown>).map(([k, v]) => `${k}=${String(v)}`).join('\n');
  } catch {
    return e.sample;
  }
}

function ErrorDrawer({ e, onClose }: { e: ErrorGroup | null; onClose: () => void }): ReactNode {
  return (
    <Drawer open={!!e} onClose={onClose} width={Math.min(720, window.innerWidth)} title={e ? ERR_LABEL[e.kind] ?? e.kind : ''}>
      {e ? (
        <>
          <Descriptions size="small" column={1} bordered items={[
            { key: 's', label: 'Số người dính', children: e.sessions },
            { key: 'c', label: 'Số lần', children: e.count },
            { key: 'm', label: 'Trên điện thoại', children: e.mobile },
            { key: 'f', label: 'Lần đầu', children: fullTime(e.first_at) },
            { key: 'l', label: 'Lần cuối', children: fullTime(e.last_at) },
            { key: 'd', label: 'Thiết bị', children: (e.devices ?? []).join(', ') || '—' },
            { key: 'b', label: 'Bản build', children: (e.builds ?? []).join(', ') || '—' },
          ]} />
          <Typography.Title level={5} style={{ marginTop: 16 }}>Nội dung</Typography.Title>
          <Pre>{shown(e)}</Pre>
        </>
      ) : null}
    </Drawer>
  );
}

export function ErrorsPage(): ReactNode {
  const [days, setDays] = useState(7);
  const { data, error, loading } = useFetch<ErrorsResp>(`/api/admin/errors?days=${days}`);
  const [open, setOpen] = useState<ErrorGroup | null>(null);
  const g = data?.groups ?? [];
  const kinds = [...new Set(g.map((x) => x.kind))];
  return (
    <div>
      <Space wrap style={{ marginBottom: 12 }}>
        <Select value={days} onChange={setDays} style={{ width: 150 }} options={[{ value: 1, label: '24 giờ qua' }, { value: 7, label: '7 ngày qua' }, { value: 30, label: '30 ngày qua' }]} />
        <Typography.Text type="secondary">
          {g.length ? `${g.length} loại lỗi · ${g.reduce((a, x) => a + x.count, 0)} lần · ${g.reduce((a, x) => a + x.sessions, 0)} lượt người dính` : ''}
        </Typography.Text>
      </Space>
      {data?.off ? <Alert type="info" showIcon message={data.note} style={{ marginBottom: 12 }} /> : null}
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Table<ErrorGroup>
        {...tableBase<ErrorGroup>()}
        rowKey={(e) => `${e.kind}|${e.sig}`}
        loading={loading && !data}
        dataSource={g}
        onRow={clickRow(setOpen)}
        locale={{ emptyText: <Empty description="Không có lỗi nào. Tốt." /> }}
        columns={[
          { title: 'Loại', dataIndex: 'kind', filters: kinds.map((k) => ({ text: ERR_LABEL[k] ?? k, value: k })), onFilter: (v, e) => e.kind === v, render: (v: string) => <Tag color={v === 'error' ? 'red' : 'orange'}>{ERR_LABEL[v] ?? v}</Tag> },
          { title: 'Người dính', dataIndex: 'sessions', align: 'right', sorter: (a, b) => a.sessions - b.sessions, defaultSortOrder: 'descend' },
          { title: 'Số lần', dataIndex: 'count', align: 'right', sorter: (a, b) => a.count - b.count },
          { title: 'Điện thoại', dataIndex: 'mobile', align: 'right', sorter: (a, b) => a.mobile - b.mobile },
          { title: 'Nội dung', dataIndex: 'sig', render: (v: string) => <Typography.Text ellipsis style={{ maxWidth: 460, fontFamily: 'ui-monospace, monospace', fontSize: 12 }}>{v}</Typography.Text> },
          { title: 'Lần cuối', dataIndex: 'last_at', render: (v: number) => <RelTime at={v} />, sorter: (a, b) => a.last_at - b.last_at },
        ]}
      />
      <ErrorDrawer e={open} onClose={() => setOpen(null)} />
    </div>
  );
}
