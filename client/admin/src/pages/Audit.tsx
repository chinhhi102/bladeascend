// Nhật ký thao tác — ADMIN_SPEC §1 "Nhật ký thao tác", §7: 200 dòng gần nhất của `admin_audit`.
import { Alert, Descriptions, Drawer, Empty, Input, Space, Table, Tag } from 'antd';
import { useState, type ReactNode } from 'react';

import { fullTime, includesCI } from '../format';
import { useFetch } from '../hooks';
import type { AuditEntry } from '../types';
import { clickRow, Pre, RelTime, tableBase } from '../ui';

function prettyJSON(raw: string): string {
  if (!raw) {
    return '';
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

const auditKey = (e: AuditEntry, i?: number): string => `${e.at}|${e.action}|${e.target}|${i ?? ''}`;

export function AuditPage(): ReactNode {
  const { data, error, loading } = useFetch<AuditEntry[]>('/api/admin/audit');
  const [q, setQ] = useState('');
  const [open, setOpen] = useState<AuditEntry | null>(null);
  const all = data ?? [];
  const rows = all.filter((e) => !q || includesCI(e.target, q) || includesCI(e.reason, q) || includesCI(e.action, q));
  const actions = [...new Set(all.map((e) => e.action))].sort();
  return (
    <div>
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Space style={{ marginBottom: 12 }}>
        <Input.Search allowClear placeholder="Tìm hành động, đối tượng, lý do..." value={q} onChange={(e) => setQ(e.target.value)} style={{ width: 280 }} />
      </Space>
      <Table<AuditEntry>
        {...tableBase<AuditEntry>()}
        rowKey={(e) => auditKey(e, all.indexOf(e))}
        loading={loading && !data}
        dataSource={rows}
        onRow={clickRow(setOpen)}
        locale={{ emptyText: <Empty description="Chưa có thao tác nào." /> }}
        columns={[
          { title: 'Thời gian', dataIndex: 'at', render: (v: number) => <RelTime at={v} />, sorter: (a, b) => a.at - b.at, defaultSortOrder: 'descend' },
          { title: 'Hành động', dataIndex: 'action', filters: actions.map((a) => ({ text: a, value: a })), onFilter: (v, e) => e.action === v, render: (v: string) => <Tag>{v}</Tag> },
          { title: 'Đối tượng', dataIndex: 'target' },
          { title: 'Lý do', dataIndex: 'reason', render: (v: string) => v || '—' },
          { title: 'IP', dataIndex: 'ip' },
        ]}
      />
      <Drawer open={!!open} onClose={() => setOpen(null)} title={open?.action} width={Math.min(560, window.innerWidth)}>
        {open ? (
          <>
            <Descriptions size="small" column={1} bordered items={[
              { key: 't', label: 'Thời gian', children: fullTime(open.at) },
              { key: 'a', label: 'Hành động', children: open.action },
              { key: 'o', label: 'Đối tượng', children: open.target },
              { key: 'r', label: 'Lý do', children: open.reason || '—' },
              { key: 'i', label: 'IP admin', children: open.ip },
            ]} />
            {open.details ? <Pre>{prettyJSON(open.details)}</Pre> : null}
          </>
        ) : null}
      </Drawer>
    </div>
  );
}
