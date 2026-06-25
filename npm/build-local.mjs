#!/usr/bin/env node

import { mkdirSync, chmodSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { getPlatform } = require("./platform.js");

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const platform = getPlatform();
const outDir = join(root, "npm", "bin", platform.key);
const outPath = join(outDir, platform.binary);

mkdirSync(outDir, { recursive: true });

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
  console.error(`go build failed: ${result.error.message}`);
  process.exit(1);
}

if ((result.status ?? 1) !== 0) {
  process.exit(result.status ?? 1);
}

if (platform.goos !== "windows") {
  chmodSync(outPath, 0o755);
}

console.log(`Built ${platform.key} binary at ${outPath}`);
