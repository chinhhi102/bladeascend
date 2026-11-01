// Khung trang quản trị — ADMIN_SPEC §2, §2.1: Layout + Sider thu gọn được + Header (build,
// đăng xuất) + Content; sáng/tối theo máy (đổi tay được).
import { App as AntApp, Badge, Button, ConfigProvider, Layout, Menu, Result, Segmented, Spin, Typography, theme } from 'antd';
import viVN from 'antd/locale/vi_VN';
import { lazy, Suspense, useEffect, useState, type ReactNode } from 'react';

import { api, post, setCsrf } from './api';
import { go, StatsContext, useHashRoute, useStats, useStatsPoll } from './hooks';
import { AuditPage } from './pages/Audit';
import { BansPage } from './pages/Bans';
import { BoardPage } from './pages/Board';
import { CharsPage } from './pages/Chars';
import { DashboardPage } from './pages/Dashboard';
import { ErrorsPage } from './pages/Errors';
import { FeedbackPage } from './pages/Feedback';
import { PlayersPage } from './pages/Players';
import { RoomsPage } from './pages/Rooms';
import type { Me } from './types';

// Trang Bản đồ kéo theo PixiJS + bộ vẽ của game: tải riêng, chỉ khi mở trang đó.
const MapPage = lazy(() => import('./pages/MapPage').then((m) => ({ default: m.MapPage })));

declare const __BRAND__: { name: string; name_vi: string };

type ThemeMode = 'auto' | 'light' | 'dark';

const THEME_KEY = 'admin_theme';

function readTheme(): ThemeMode {
  try {
    const v = localStorage.getItem(THEME_KEY);
    return v === 'light' || v === 'dark' ? v : 'auto';
  } catch {
    return 'auto';
  }
}

function useSystemDark(): boolean {
  const q = '(prefers-color-scheme: dark)';
  const [dark, setDark] = useState(() => window.matchMedia(q).matches);
  useEffect(() => {
    const m = window.matchMedia(q);
    const on = (): void => setDark(m.matches);
    m.addEventListener('change', on);
    return () => m.removeEventListener('change', on);
  }, []);
  return dark;
}

interface PageDef {
  key: string;
  label: string;
  render: (args: string[]) => ReactNode;
}

const PAGES: PageDef[] = [
  { key: 'dashboard', label: 'Dashboard', render: () => <DashboardPage /> },
  { key: 'rooms', label: 'Phòng & Khu', render: () => <RoomsPage /> },
  { key: 'feedback', label: 'Góp ý', render: () => <FeedbackPage /> },
  { key: 'map', label: 'Bản đồ', render: (a) => <MapPage args={a} /> },
  { key: 'players', label: 'Người chơi', render: () => <PlayersPage /> },
  { key: 'errors', label: 'Lỗi người chơi', render: () => <ErrorsPage /> },
  { key: 'bans', label: 'Danh sách cấm', render: () => <BansPage /> },
  { key: 'chars', label: 'Nhân vật', render: () => <CharsPage /> },
  { key: 'board', label: 'Bảng xếp hạng', render: () => <BoardPage /> },
  { key: 'audit', label: 'Nhật ký', render: () => <AuditPage /> },
];

function MenuLabel({ text, count, color }: { text: string; count?: number; color?: string }): ReactNode {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
      {text}
      {count ? <Badge count={count} color={color} size="small" /> : null}
    </span>
  );
}

function SideMenu({ page, onPick }: { page: string; onPick: () => void }): ReactNode {
  const { now } = useStats();
  const counts: Record<string, [number | undefined, string]> = {
    feedback: [now?.feedback_new, '#1677ff'],
    players: [now?.flagged, '#faad14'],
    errors: [now?.errors_today, '#ff4d4f'],
  };
  const items = PAGES.map((p) => ({
    key: p.key,
    label: <MenuLabel text={p.label} count={counts[p.key]?.[0]} color={counts[p.key]?.[1]} />,
  }));
  return (
    <Menu
      theme="dark"
      mode="inline"
      selectedKeys={[page]}
      items={items}
      onClick={(e) => {
        go(e.key);
        onPick();
      }}
    />
  );
}

async function logout(): Promise<void> {
  await post('/api/admin/logout', {}).catch(() => undefined);
  location.href = '/admin/login';
}

function Shell({ build, mode, setMode }: { build: string; mode: ThemeMode; setMode: (m: ThemeMode) => void }): ReactNode {
  const route = useHashRoute();
  const stats = useStatsPoll();
  const [collapsed, setCollapsed] = useState(false);
  const [mobile, setMobile] = useState(false);
  const def = PAGES.find((p) => p.key === route.page) ?? PAGES[0];
  const { token } = theme.useToken();
  return (
    <StatsContext.Provider value={stats}>
      <Layout style={{ minHeight: '100vh' }}>
        <Layout.Sider
          breakpoint="lg"
          collapsedWidth={mobile ? 0 : 64}
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          onBreakpoint={(b) => {
            setMobile(b);
            setCollapsed(b);
          }}
          style={mobile ? { position: 'fixed', zIndex: 20, height: '100vh' } : undefined}
        >
          <div style={{ color: '#fff', padding: '16px 20px', fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden' }}>{collapsed && !mobile ? 'QT' : __BRAND__.name}</div>
          <SideMenu page={def.key} onPick={() => mobile && setCollapsed(true)} />
        </Layout.Sider>
        <Layout>
          <Layout.Header style={{ background: token.colorBgContainer, padding: '0 16px', display: 'flex', alignItems: 'center', gap: 12 }}>
            <Typography.Title level={4} style={{ margin: 0, flex: 1, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{def.label}</Typography.Title>
            {mobile ? null : <Typography.Text type="secondary" style={{ fontSize: 12 }} title="Bản build đang chạy">{stats.now?.build ?? build}</Typography.Text>}
            {mobile ? null : <Segmented size="small" value={mode} onChange={(v) => setMode(v as ThemeMode)} options={[{ value: 'auto', label: 'Tự' }, { value: 'light', label: 'Sáng' }, { value: 'dark', label: 'Tối' }]} />}
            <Button size={mobile ? 'small' : 'middle'} onClick={() => void logout()}>Đăng xuất</Button>
          </Layout.Header>
          <Layout.Content style={{ padding: 16, minWidth: 0 }}>
            <Suspense fallback={<Spin />}>{def.render(route.args)}</Suspense>
          </Layout.Content>
        </Layout>
      </Layout>
    </StatsContext.Provider>
  );
}

function Boot({ mode, setMode }: { mode: ThemeMode; setMode: (m: ThemeMode) => void }): ReactNode {
  const [me, setMe] = useState<Me | null>(null);
  const [err, setErr] = useState('');
  useEffect(() => {
    api<Me>('/api/admin/me')
      .then((m) => {
        setCsrf(m.csrf);
        setMe(m);
      })
      .catch((e: unknown) => setErr(e instanceof Error ? e.message : String(e)));
  }, []);
  if (err) {
    return <Result status="warning" title="Không mở được trang quản trị" subTitle={err} extra={<Button href="/admin/login">Đăng nhập lại</Button>} />;
  }
  if (!me) {
    return <Spin fullscreen />;
  }
  return <Shell build={me.build} mode={mode} setMode={setMode} />;
}

export function AdminApp(): ReactNode {
  const [mode, setModeState] = useState<ThemeMode>(readTheme);
  const sysDark = useSystemDark();
  const dark = mode === 'dark' || (mode === 'auto' && sysDark);
  const setMode = (m: ThemeMode): void => {
    setModeState(m);
    try {
      localStorage.setItem(THEME_KEY, m);
    } catch {
      // Không lưu được (chế độ riêng tư) thì chỉ mất lựa chọn ở lần mở sau.
    }
  };
  useEffect(() => {
    document.body.style.background = dark ? '#000' : '#f5f5f5';
  }, [dark]);
  return (
    <ConfigProvider locale={viVN} theme={{ algorithm: dark ? theme.darkAlgorithm : theme.defaultAlgorithm }}>
      <AntApp>
        <Boot mode={mode} setMode={setMode} />
      </AntApp>
    </ConfigProvider>
  );
}
