/// <reference types="vitest/config" />
import { defineConfig } from 'vite'

// A module of the fleet's gateway webapp, served on the gateway's origin at
// /-/apps/tenantecho/: the fleet builder passes that path as TACPLUS_WEB_BASE
// (tacplus docs/USER.md, "Contributing to the gateway").
export default defineConfig({
  base: process.env.TACPLUS_WEB_BASE,
  define: { global: 'globalThis' },
  build: { outDir: 'dist' },
  test: { environment: 'happy-dom' },
})
