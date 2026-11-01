// Hòm thư góp ý — ADMIN_SPEC §7.2. Nội dung và tên là thứ NGƯỜI CHƠI tự nhập: chỉ hiển thị
// dạng text (JSX), không bao giờ dựng HTML từ nó (§1 "XSS", §11).
import { Button, Descriptions, Drawer, Empty, Image, Input, Select, Space, Table, Tag, Typography } from 'antd';
import { useEffect, useMemo, useState, type ReactNode } from 'react';

import { api } from '../api';
import { FB_KINDS, FB_STATUS, FB_STATUS_COLOR, fullTime } from '../format';
import { useFetch } from '../hooks';
import type { Feedback } from '../types';
import { clickRow, Pre, RelTime, tableBase, useAsk, useRun } from '../ui';

const options = (m: Record<string, string>): { value: string; label: string }[] => Object.entries(m).map(([value, label]) => ({ value, label }));

/**
 * imageCount — số ảnh của một lá thư. Server mới gửi `images` (0–3); bản cũ chỉ có
 * `has_image` (một ảnh).
 */
export function imageCount(f: Feedback): number {
  if (typeof f.images === 'number') {
    return Math.max(0, Math.min(3, f.images));
  }
  return f.has_image ? 1 : 0;
}

/** imageUrl — ảnh 0 giữ đúng đường cũ (không `?i=`) để chạy được với server cũ. */
export function imageUrl(id: number, i: number): string {
  return i === 0 ? `/api/admin/feedback/${id}/image` : `/api/admin/feedback/${id}/image?i=${i}`;
}

interface Ctx {
  pairs: [string, string][];
  console: string[];
}

function parseContext(raw: string): Ctx {
  if (!raw || raw === '{}') {
    return { pairs: [], console: [] };
  }
  try {
    const o = JSON.parse(raw) as Record<string, unknown>;
    const lines = Array.isArray(o.console) ? o.console.map(String) : [];
    delete o.console;
    return { pairs: Object.entries(o).map(([k, v]) => [k, typeof v === 'object' ? JSON.stringify(v) : String(v)]), console: lines };
  } catch {
    return { pairs: [['ngữ cảnh', raw]], console: [] };
  }
}

function Images({ f }: { f: Feedback }): ReactNode {
  const n = imageCount(f);
  if (!n) {
    return <Typography.Text type="secondary">Không có ảnh.</Typography.Text>;
  }
  return (
    <Image.PreviewGroup>
      <Space wrap>
        {Array.from({ length: n }, (_, i) => (
          <Image key={i} width={120} height={90} style={{ objectFit: 'cover' }} src={imageUrl(f.id, i)} alt={`ảnh ${i + 1}`} />
        ))}
      </Space>
    </Image.PreviewGroup>
  );
}

function StatusEditor({ f, onSaved }: { f: Feedback; onSaved: () => void }): ReactNode {
  const [status, setStatus] = useState(f.status);
  const [note, setNote] = useState(f.note);
  const ask = useAsk();
  const run = useRun();
  useEffect(() => {
    setStatus(f.status);
    setNote(f.note);
  }, [f]);
  const save = async (): Promise<void> => {
    const reason = await ask({ title: `Lưu góp ý #${f.id}: "${FB_STATUS[status] ?? status}"?`, okText: 'Lưu' });
    if (reason !== null && (await run(() => api(`/api/admin/feedback/${f.id}`, { method: 'PATCH', body: { status, note, reason } }), 'Đã lưu.'))) {
      onSaved();
    }
  };
  return (
    <Space direction="vertical" style={{ width: '100%' }}>
      <Select value={status} onChange={setStatus} options={options(FB_STATUS)} style={{ width: 200 }} />
      <Input.TextArea rows={3} value={note} onChange={(e) => setNote(e.target.value)} placeholder="Ghi chú riêng của admin..." />
      <Button type="primary" onClick={() => void save()}>Lưu</Button>
    </Space>
  );
}

function FeedbackDrawer({ f, onClose, onSaved }: { f: Feedback | null; onClose: () => void; onSaved: () => void }): ReactNode {
  const ctx = useMemo(() => parseContext(f?.context ?? ''), [f]);
  return (
    <Drawer open={!!f} onClose={onClose} width={Math.min(640, window.innerWidth)} title={f ? `Góp ý #${f.id}` : ''}>
      {f ? (
        <div className="fb">
          <Space wrap style={{ marginBottom: 12 }}>
            <Tag>{FB_KINDS[f.kind] ?? f.kind}</Tag>
            <Tag color={FB_STATUS_COLOR[f.status]}>{FB_STATUS[f.status] ?? f.status}</Tag>
            <Typography.Text>{fullTime(f.at)}</Typography.Text>
          </Space>
          <Typography.Paragraph className="body" style={{ whiteSpace: 'pre-wrap', fontSize: 15 }}>{f.body}</Typography.Paragraph>
          <Typography.Title level={5}>Ảnh</Typography.Title>
          <Images f={f} />
          <Typography.Title level={5} style={{ marginTop: 16 }}>Người gửi & ngữ cảnh</Typography.Title>
          <Descriptions size="small" column={1} bordered items={[
            { key: 'name', label: 'Tên', children: f.name || '(không tên)' },
            { key: 'gid', label: 'guest_id', children: f.guest_id },
            ...ctx.pairs.map(([k, v]) => ({ key: k, label: k, children: v })),
          ]} />
          {ctx.console.length ? <><Typography.Title level={5} style={{ marginTop: 16 }}>{`${ctx.console.length} dòng lỗi console`}</Typography.Title><Pre>{ctx.console.join('\n')}</Pre></> : null}
          <Typography.Title level={5} style={{ marginTop: 16 }}>Trạng thái</Typography.Title>
          <StatusEditor f={f} onSaved={onSaved} />
        </div>
      ) : null}
    </Drawer>
  );
}

function useQuery(kind: string, status: string, q: string): string {
  const [debounced, setDebounced] = useState(q);
  useEffect(() => {
    const t = window.setTimeout(() => setDebounced(q.trim()), 300);
    return () => window.clearTimeout(t);
  }, [q]);
  const p = new URLSearchParams();
  if (kind) {
    p.set('kind', kind);
  }
  if (status) {
    p.set('status', status);
  }
  if (debounced) {
    p.set('q', debounced);
  }
  return `/api/admin/feedback?${p}`;
}

export function FeedbackPage(): ReactNode {
  const [kind, setKind] = useState('');
  const [status, setStatus] = useState('');
  const [q, setQ] = useState('');
  const { data, loading, reload } = useFetch<Feedback[]>(useQuery(kind, status, q));
  const [openId, setOpenId] = useState<number | null>(null);
  const open = (data ?? []).find((f) => f.id === openId) ?? null;
  return (
    <div className="fb">
      <Space wrap style={{ marginBottom: 12 }}>
        <Select value={kind} onChange={setKind} style={{ width: 150 }} options={[{ value: '', label: 'Mọi loại' }, ...options(FB_KINDS)]} />
        <Select value={status} onChange={setStatus} style={{ width: 160 }} options={[{ value: '', label: 'Mọi trạng thái' }, ...options(FB_STATUS)]} />
        <Input.Search allowClear placeholder="Tìm trong nội dung..." value={q} onChange={(e) => setQ(e.target.value)} style={{ width: 240 }} />
        <Button href="/api/admin/feedback.csv" download>Xuất CSV</Button>
      </Space>
      <Table<Feedback>
        {...tableBase<Feedback>()}
        rowKey="id"
        loading={loading && !data}
        dataSource={data ?? []}
        onRow={clickRow((f) => setOpenId(f.id))}
        locale={{ emptyText: <Empty description="Chưa có góp ý nào." /> }}
        columns={[
          { title: 'Lúc', dataIndex: 'at', render: (v: number) => <RelTime at={v} />, sorter: (a, b) => a.at - b.at, defaultSortOrder: 'descend' },
          { title: 'Loại', dataIndex: 'kind', render: (v: string) => <Tag>{FB_KINDS[v] ?? v}</Tag> },
          { title: 'Trạng thái', dataIndex: 'status', render: (v: string) => <Tag color={FB_STATUS_COLOR[v]}>{FB_STATUS[v] ?? v}</Tag> },
          { title: 'Tên', dataIndex: 'name', render: (v: string) => v || '(không tên)' },
          { title: 'Nội dung', dataIndex: 'body', render: (v: string) => <Typography.Text className="body" ellipsis style={{ maxWidth: 420 }}>{v}</Typography.Text> },
          { title: 'Ảnh', key: 'img', align: 'right', render: (_, f) => imageCount(f) || '' },
          { title: 'Ghi chú', dataIndex: 'note', render: (v: string) => <Typography.Text type="secondary" ellipsis style={{ maxWidth: 200 }}>{v}</Typography.Text> },
        ]}
      />
      <FeedbackDrawer f={open} onClose={() => setOpenId(null)} onSaved={reload} />
    </div>
  );
}
