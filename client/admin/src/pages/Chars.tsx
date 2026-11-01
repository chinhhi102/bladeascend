// Bảng theo nhân vật — GAME_SPEC v6.30 §6.3 (Q36): cân bằng theo số liệu người chơi thật.
import { Alert, Descriptions, Drawer, Empty, Select, Space, Table, Typography } from 'antd';
import { useState, type ReactNode } from 'react';

import { f1 } from '../format';
import { useFetch } from '../hooks';
import type { CharRow } from '../types';
import { clickRow, tableBase } from '../ui';

interface CharsResp {
  days: number;
  rows: CharRow[] | null;
}

export function CharsPage(): ReactNode {
  const [days, setDays] = useState(2);
  // Số liệu theo ngày: làm mới mỗi phút là đủ.
  const { data, error, loading } = useFetch<CharsResp>(`/api/admin/characters?days=${days}`, 60000);
  const [open, setOpen] = useState<CharRow | null>(null);
  return (
    <div>
      <Typography.Paragraph type="secondary">Cân bằng theo số liệu người chơi thật ≥ 2 ngày (GAME_SPEC v6.30 §6.3). Bỏ mạng &lt; 30 s và mạng đã ẩn.</Typography.Paragraph>
      <Space style={{ marginBottom: 12 }}>
        <span>Số ngày</span>
        <Select value={days} onChange={setDays} style={{ width: 100 }} options={[1, 2, 7, 30].map((d) => ({ value: d, label: String(d) }))} />
      </Space>
      {error ? <Alert type="error" message={error} style={{ marginBottom: 12 }} /> : null}
      <Table<CharRow>
        {...tableBase<CharRow>()}
        pagination={false}
        rowKey="character"
        loading={loading && !data}
        dataSource={data?.rows ?? []}
        onRow={clickRow(setOpen)}
        locale={{ emptyText: <Empty description="Chưa có mạng nào đủ 30 giây trong khoảng này." /> }}
        columns={[
          { title: 'Nhân vật', dataIndex: 'character', sorter: (a, b) => a.character.localeCompare(b.character) },
          { title: 'Số mạng', dataIndex: 'lives', align: 'right', sorter: (a, b) => a.lives - b.lives, defaultSortOrder: 'descend' },
          { title: 'Tỉ lệ chọn', dataIndex: 'pick_share', align: 'right', render: (v: number) => `${f1(v)}%`, sorter: (a, b) => a.pick_share - b.pick_share },
          { title: 'Uy Danh/phút', dataIndex: 'score_per_min', align: 'right', render: f1, sorter: (a, b) => a.score_per_min - b.score_per_min },
          { title: 'Sống TB', dataIndex: 'avg_survive', align: 'right', render: (v: number) => `${Math.round(v)} s`, sorter: (a, b) => a.avg_survive - b.avg_survive },
          { title: 'Cảnh giới cao nhất TB', dataIndex: 'avg_peak_realm', align: 'right', render: f1, sorter: (a, b) => a.avg_peak_realm - b.avg_peak_realm },
          { title: 'Dùng kỹ năng/phút', dataIndex: 'casts_per_min', align: 'right', render: f1, sorter: (a, b) => a.casts_per_min - b.casts_per_min },
        ]}
      />
      <Drawer open={!!open} onClose={() => setOpen(null)} title={open?.character} width={Math.min(480, window.innerWidth)}>
        {open ? (
          <Descriptions size="small" column={1} bordered items={[
            { key: 'l', label: 'Số mạng', children: open.lives },
            { key: 'p', label: 'Tỉ lệ chọn', children: `${f1(open.pick_share)}%` },
            { key: 's', label: 'Uy Danh / phút sống', children: f1(open.score_per_min) },
            { key: 'a', label: 'Sống TB', children: `${Math.round(open.avg_survive)} s` },
            { key: 'r', label: 'Cảnh giới cao nhất TB', children: f1(open.avg_peak_realm) },
            { key: 'c', label: 'Dùng kỹ năng / phút', children: f1(open.casts_per_min) },
            { key: 't', label: 'Tổng lần dùng kỹ năng', children: open.casts },
          ]} />
        ) : null}
      </Drawer>
    </div>
  );
}
