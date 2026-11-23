/**
 * destroyDeep — huỷ cả một nhánh cây cảnh, kể cả `GraphicsContext` của riêng nó.
 *
 * ── Vì sao không dùng `destroy({ children: true })` ─────────────────────────────────────
 *
 * Vì nó KHÔNG huỷ `GraphicsContext`, và chuyện đó im lặng tuyệt đối. `Container.destroy`
 * truyền nguyên options xuống từng con, còn `Graphics.destroy(options)` chỉ huỷ context của
 * chính nó khi:
 *
 *     if (this._ownedContext && !options)                      → huỷ
 *     else if (options === true || options?.context === true)   → huỷ
 *
 * `{ children: true }` rơi vào đúng nhánh KHÔNG huỷ: options tồn tại, `options.context`
 * không phải `true`. Context sống sót thì nó vẫn nằm trong sổ `_managedContexts` của
 * renderer — và sổ đó chỉ nhả khi context **bị huỷ** (`unload()` phát sự kiện `unload`). Mỗi
 * context còn sống giữ `GraphicsContextRenderData` → `Batcher` → geometry → bộ đệm đỉnh
 * **~143 KB**.
 *
 * Đo thật 23/11 (ảnh heap, kịch bản cố định §13.9, 20 phút): **717 context mồ côi** — không
 * còn gì trong cảnh trỏ tới chúng, chỉ cái sổ của renderer giữ. Không một chỉ số nào trong
 * F3 nhúc nhích: context không nằm trong cây cảnh, không phải texture.
 *
 * ── Vì sao không chỉ thêm `context: true` vào mọi chỗ ───────────────────────────────────
 *
 * Đã thử, và nó làm vỡ hình ngay 15 giây đầu: `entities.ts` cố ý dùng **context DÙNG CHUNG**
 * (`new Graphics(shared('orb:…'))`) để hàng trăm viên Linh khí cùng một hình chỉ tốn một
 * context. `context: true` huỷ luôn cái context đang được người khác dùng → texture về 0,
 * FPS 16. Context dùng chung được cache theo khoá nên nó KHÔNG rò rỉ; thứ rò rỉ là context
 * của riêng từng Graphics.
 *
 * Nên chỗ này phân biệt đúng hai loại: `_ownedContext` có giá trị thì context ấy là của
 * riêng nó, huỷ được; không thì để nguyên.
 */

/** Nút cây cảnh ở mức tối thiểu hàm này cần. Không ràng vào kiểu của Pixi để test chạy trong Node. */
interface DeepNode {
  children?: DeepNode[];
  destroy?: (options: { children: boolean; context?: boolean }) => void;
  /** Pixi đặt cờ này trên `Graphics` tự tạo context (không truyền context vào constructor). */
  _ownedContext?: unknown;
}

export function destroyDeep(node: DeepNode | null | undefined): void {
  if (!node) {
    return;
  }
  // Huỷ con TRƯỚC: `destroy` của Pixi gỡ nút khỏi cha nó, nên phải chụp lại danh sách con
  // rồi mới đi, và đi từ dưới lên để không ai còn trỏ vào thứ đã huỷ.
  for (const child of [...(node.children ?? [])]) {
    destroyDeep(child);
  }
  node.destroy?.(node._ownedContext ? { children: false, context: true } : { children: false });
}
