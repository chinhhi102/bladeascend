// Trang "Dọn dữ liệu" — ADMIN_SPEC §13 (v1.6).
//
// Ba câu hỏi, đúng theo thứ tự người trực hỏi:
//   1. Hiện đang nặng ở đâu? → bảng dung lượng theo bảng dữ liệu và theo tệp.
//   2. Lượt dọn kế tiếp SẼ xoá gì? → bản xem trước (dry-run) chạy sẵn, không phải bấm nút.
//   3. Lần chạy gần nhất xoá bao nhiêu? → đọc từ nhật ký quản trị.
//
// Nút "Dọn ngay" có hộp xác nhận ghi rõ TỔNG SỐ DÒNG sẽ xoá: xoá là việc không lùi được,
// nên con số đó phải nằm ngay trên nút, không phải ở một bảng người ta vừa cuộn qua.
import { Alert, App, Button, Card, Descriptions, Popconfirm, Space, Statistic, Table, Tag, Typography } from 'antd';
import { useState, type ReactNode } from 'react';

import { errMsg, post } from '../api';
import { fullTime } from '../format';
import { useFetch } from '../hooks';
import { RelTime, tableBase } from '../ui';

interface JanitorJob {
  key: string;
  what: string;
  keep: string;
  rows: number;
  err?: string;
}

interface JanitorResult {
  at: number;
  dry_run?: boolean;
  jobs: JanitorJob[] | null;
  rows: number;
  vacuumed?: boolean;
  freed_bytes?: number;
  vacuum_note?: string;
}

interface TableSize {
  name: string;
  rows: number;
  bytes: number;
}

interface FileSize {
  name: string;
  bytes: number;
  warn: boolean;
}

interface AuditEntryLite {
  at: number;
  reason: string;
  details: string;
}

interface JanitorView {
  off?: boolean;
  note?: string;
  tables: TableSize[] | null;
  files: FileSize[] | null;
  plan: JanitorResult;
  last: AuditEntryLite | null;
  warning: string;
  next_at: number;
  armed: boolean;
}

/** mb hiện dung lượng ở đơn vị đọc được bằng mắt, không phải byte thô. */
function mb(bytes: number): string {
  if (bytes >= 1 << 30) {
    return `${(bytes / (1 << 30)).toFixed(2)} GB`;
  }
  if (bytes >= 1 << 20) {
    return `${(bytes / (1 << 20)).toFixed(1)} MB`;
  }
  return `${(bytes / 1024).toFixed(0)} KB`;
}

export function JanitorPage(): ReactNode {
  const { data, error, loading, reload } = useFetch<JanitorView>('/api/admin/janitor');
  const [busy, setBusy] = useState(false);
  const { message } = App.useApp();

  if (data?.off) {
    return <Alert type="info" showIcon message={data.note ?? 'Bộ dọn chưa bật.'} />;
  }
  const plan = data?.plan;
  const jobs = (plan?.jobs ?? []).filter((j) => j.rows > 0 || j.err);
  const willDelete = plan?.rows ?? 0;

  const runNow = async (): Promise<void> => {
    setBusy(true);
    try {
      const res = await post<JanitorResult>('/api/admin/janitor/run', {});
      message.success(`Đã xoá ${res.rows} dòng${res.vacuumed ? ` · VACUUM giải phóng ${mb(res.freed_bytes ?? 0)}` : ''}`);
      reload();
    } catch (e) {
      message.error(errMsg(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      {data?.warning ? <Alert type="warning" showIcon message={data.warning} style={{ marginBottom: 12 }} /> : null}

      <Space wrap align="start" style={{ marginBottom: 16 }}>
        <Card size="small">
          <Statistic title="Lượt dọn kế tiếp sẽ xoá" value={willDelete} suffix="dòng" loading={loading && !data} />
        </Card>
        <Card size="small">
          <Statistic title="Tổng dung lượng dữ liệu" value={mb((data?.files ?? []).reduce((a, f) => a + f.bytes, 0))} loading={loading && !data} />
        </Card>
        <Card size="small" title="Lịch">
          <div>{data?.armed ? 'Tự chạy 04:30 mỗi ngày' : 'Tự chạy 04:30 — hiện CHỈ CHẠY THỬ'}</div>
          <Typography.Text type="secondary">
            kế tiếp: {data?.next_at ? fullTime(data.next_at) : '—'}
          </Typography.Text>
          <div style={{ marginTop: 8 }}>
            <Popconfirm
              title="Dọn dữ liệu ngay?"
              description={`Sẽ xoá ${willDelete} dòng theo bảng giữ dữ liệu §13. Không lùi lại được.`}
              okText="Dọn ngay"
              cancelText="Thôi"
              onConfirm={() => void runNow()}
            >
              <Button type="primary" danger loading={busy} disabled={!data}>Dọn ngay</Button>
            </Popconfirm>
          </div>
        </Card>
        {data?.last ? (
          <Card size="small" title="Lần chạy gần nhất">
            <div><RelTime at={data.last.at} /></div>
            <Typography.Text type="secondary">{data.last.reason}</Typography.Text>
            <Pre details={data.last.details} />
          </Card>
        ) : null}
      </Space>

      {data && !data.armed ? (
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
          message="Chưa bật xoá thật — mỗi đêm chỉ chạy thử và ghi danh sách vào Nhật ký."
          description="Đọc bảng “Sẽ xoá gì” bên dưới. Bấm “Dọn ngay” một lần là đồng ý, và từ đêm sau job 04:30 xoá thật."
        />
      ) : null}
      {plan?.vacuum_note ? <Alert type="info" showIcon message={plan.vacuum_note} style={{ marginBottom: 12 }} /> : null}

      <Typography.Title level={5}>Sẽ xoá gì (chạy thử)</Typography.Title>
      <Table<JanitorJob>
        {...tableBase<JanitorJob>()}
        rowKey="key"
        loading={loading && !data}
        dataSource={jobs}
        pagination={false}
        locale={{ emptyText: 'Không có gì quá hạn. Tốt.' }}
        columns={[
          { title: 'Dữ liệu', dataIndex: 'what' },
          { title: 'Giữ', dataIndex: 'keep', render: (v: string) => <Tag>{v}</Tag> },
          { title: 'Sẽ xoá', dataIndex: 'rows', align: 'right', sorter: (a, b) => a.rows - b.rows, defaultSortOrder: 'descend' },
          { title: 'Lỗi', dataIndex: 'err', render: (v?: string) => (v ? <Typography.Text type="danger">{v}</Typography.Text> : '') },
        ]}
      />

      <Typography.Title level={5} style={{ marginTop: 20 }}>Dung lượng theo tệp</Typography.Title>
      <Table<FileSize>
        {...tableBase<FileSize>()}
        rowKey="name"
        dataSource={data?.files ?? []}
        pagination={false}
        columns={[
          { title: 'Tệp / thư mục', dataIndex: 'name' },
          { title: 'Dung lượng', dataIndex: 'bytes', align: 'right', render: (v: number, r) => (r.warn ? <Typography.Text type="danger">{mb(v)}</Typography.Text> : mb(v)) },
        ]}
      />

      <Typography.Title level={5} style={{ marginTop: 20 }}>Dung lượng theo bảng (ước lượng)</Typography.Title>
      <Table<TableSize>
        {...tableBase<TableSize>()}
        rowKey="name"
        dataSource={data?.tables ?? []}
        pagination={false}
        columns={[
          { title: 'Bảng', dataIndex: 'name' },
          { title: 'Số dòng', dataIndex: 'rows', align: 'right', sorter: (a, b) => a.rows - b.rows },
          { title: 'Ước lượng', dataIndex: 'bytes', align: 'right', render: (v: number) => mb(v), sorter: (a, b) => a.bytes - b.bytes, defaultSortOrder: 'descend' },
        ]}
      />
    </div>
  );
}

/** Pre hiện chi tiết JSON của lượt chạy gần nhất, gọn trong thẻ. */
function Pre({ details }: { details: string }): ReactNode {
  if (!details) {
    return null;
  }
  let rows = 0;
  let freed = 0;
  try {
    const d = JSON.parse(details) as { rows?: number; freed_bytes?: number };
    rows = d.rows ?? 0;
    freed = d.freed_bytes ?? 0;
  } catch {
    return null;
  }
  return (
    <Descriptions size="small" column={1} style={{ marginTop: 8 }} items={[
      { key: 'r', label: 'Đã xoá', children: `${rows} dòng` },
      { key: 'f', label: 'Giải phóng', children: mb(freed) },
    ]} />
  );
}
