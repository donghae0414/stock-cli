#!/usr/bin/env node

import { chmodSync, mkdirSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { PLATFORM_MAP } = require("./platform.js");

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

for (const [key, platform] of Object.entries(PLATFORM_MAP)) {
  const outDir = join(root, "npm", "bin", key);
  const outPath = join(outDir, platform.binary);
  mkdirSync(outDir, { recursive: true });

  console.log(`Building ${key}...`);
  const result = spawnSync("go", ["build", "-o", outPath, "./cmd/stock"], {
    cwd: root,
    stdio: "inherit",
    env: {
      ...process.env,
      CGO_ENABLED: process.env.CGO_ENABLED ?? "0",
      GOOS: platform.goos,
      GOARCH: platform.goarch
    }
  });

  if (result.error) {
    console.error(`go build failed for ${key}: ${result.error.message}`);
    process.exit(1);
  }

  if ((result.status ?? 1) !== 0) {
    process.exit(result.status ?? 1);
  }

  if (platform.goos !== "windows") {
    chmodSync(outPath, 0o755);
  }
}

console.log("Built all npm platform binaries.");
