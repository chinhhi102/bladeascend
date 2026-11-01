import { createRoot } from 'react-dom/client';

import { AdminApp } from './App';

const root = document.getElementById('root');
if (root) {
  createRoot(root).render(<AdminApp />);
}
