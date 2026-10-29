/**
 * Đọc `config/game.json` thẳng từ đĩa cho các script chạy bằng Node.
 *
 * Trong game, client lấy config qua `GET /api/config`; script đối chiếu va chạm không có
 * server nên đọc chính file nguồn — vẫn đúng một bộ số với server, đó mới là điều quan
 * trọng.
 */
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import type { GameConfig } from '../src/config.ts';

export function loadConfigFromJSON(): GameConfig {
  const here = dirname(fileURLToPath(import.meta.url));
  return JSON.parse(readFileSync(join(here, '../../config/game.json'), 'utf8')) as GameConfig;
}
