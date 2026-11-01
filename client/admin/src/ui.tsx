// Mảnh giao diện dùng chung: thời gian tương đối, cấu hình bảng, hộp xác nhận có lý do.
//
// Chữ do người chơi nhập (tên, góp ý) luôn đi qua JSX dạng TEXT — React tự thoát ký tự.
// Không có `dangerouslySetInnerHTML` ở bất kỳ đâu trong trang này (ADMIN_SPEC §1 "XSS").
import { App, Input, Tooltip, Typography } from 'antd';
import type { TableProps } from 'antd';
import { useState, type ReactNode } from 'react';

import { fullTime, relTime } from './format';

export function RelTime({ at }: { at: number }): ReactNode {
  if (!at) {
    return '—';
  }
  return (
    <Tooltip title={fullTime(at)}>
      <span>{relTime(at)}</span>
    </Tooltip>
  );
}

/** tableBase: quy ước §2.1 — header cố định, phân trang 20/50/100, cuộn ngang trên điện thoại. */
export function tableBase<T>(): Partial<TableProps<T>> {
  return {
    size: 'small',
    sticky: true,
    scroll: { x: 'max-content' },
    pagination: { defaultPageSize: 20, pageSizeOptions: [20, 50, 100], showSizeChanger: true },
  };
}

/** clickRow: cả hàng bấm được để mở Drawer chi tiết. */
export function clickRow<T>(open: (r: T) => void): TableProps<T>['onRow'] {
  return (r) => ({ onClick: () => open(r), style: { cursor: 'pointer' } });
}

/** Pre: khối chữ giữ nguyên xuống dòng (vết gọi hàm, JSON). */
export function Pre({ children }: { children: string }): ReactNode {
  return (
    <Typography.Paragraph style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontFamily: 'ui-monospace, monospace', fontSize: 12 }}>
      {children}
    </Typography.Paragraph>
  );
}

export interface AskOptions {
  title: string;
  content?: ReactNode;
  okText: string;
  danger?: boolean;
  /** required: bắt buộc nêu lý do (kick, cấm, ẩn bảng xếp hạng). */
  required?: boolean;
  initial?: string;
  hint?: string;
}

function ReasonField(props: { initial: string; hint: string; onChange: (v: string) => void }): ReactNode {
  const [v, setV] = useState(props.initial);
  return (
    <div style={{ marginTop: 12 }}>
      <Typography.Text type="secondary">{props.hint}</Typography.Text>
      <Input.TextArea
        autoFocus
        rows={3}
        maxLength={500}
        showCount
        value={v}
        placeholder="Lý do"
        onChange={(e) => {
          setV(e.target.value);
          props.onChange(e.target.value);
        }}
      />
    </div>
  );
}

/**
 * useAsk: `Modal.confirm` + ô "Lý do" (ADMIN_SPEC §2.1). Trả về lý do đã nhập, hoặc `null`
 * nếu bấm Huỷ. Lý do được gửi lên server trong trường `reason` và ghi vào `admin_audit`.
 */
export function useAsk(): (o: AskOptions) => Promise<string | null> {
  const { modal } = App.useApp();
  return (o) =>
    new Promise((resolve) => {
      let reason = o.initial ?? '';
      const okProps = (): { danger?: boolean; disabled: boolean } => ({ danger: o.danger, disabled: !!o.required && !reason.trim() });
      const hint = o.hint ?? (o.required ? 'Lý do (bắt buộc, ghi vào nhật ký quản trị):' : 'Lý do (ghi vào nhật ký quản trị, có thể để trống):');
      const inst = modal.confirm({
        title: o.title,
        width: 520,
        content: (
          <div>
            {o.content}
            <ReasonField
              initial={reason}
              hint={hint}
              onChange={(v) => {
                reason = v;
                inst.update({ okButtonProps: okProps() });
              }}
            />
          </div>
        ),
        okText: o.okText,
        cancelText: 'Huỷ',
        okButtonProps: okProps(),
        onOk: () => resolve(reason.trim()),
        onCancel: () => resolve(null),
      });
    });
}

/** useRun: chạy một thao tác, báo kết quả bằng `message` (§2.1). */
export function useRun(): (work: () => Promise<unknown>, okText: string) => Promise<boolean> {
  const { message } = App.useApp();
  return async (work, okText) => {
    try {
      await work();
      message.success(okText);
      return true;
    } catch (e) {
      message.error(`Không xong: ${e instanceof Error ? e.message : String(e)}`);
      return false;
    }
  };
}
