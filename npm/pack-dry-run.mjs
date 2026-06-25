#!/usr/bin/env node

import { mkdtempSync, rmSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const npmCache = mkdtempSync(join(tmpdir(), "stock-cli-npm-cache-"));
let exitCode = 1;

try {
  const result = spawnSync("npm", ["pack", "--dry-run", "--json"], {
    cwd: root,
    stdio: "inherit",
    env: {
      ...process.env,
      npm_config_cache: npmCache
    }
  });

  if (result.error) {
    throw new Error(`npm pack failed: ${result.error.message}`);
  }

  exitCode = result.status ?? 1;
} finally {
  rmSync(npmCache, { recursive: true, force: true });
}

process.exitCode = exitCode;
