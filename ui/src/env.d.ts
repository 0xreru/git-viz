/// <reference types="vite/client" />

import type { ReconData } from './types';

declare global {
  interface Window {
    GIT_RECON_DATA?: ReconData;
  }
}
