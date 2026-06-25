#!/usr/bin/env node

import { existsSync, statSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { PLATFORM_MAP } = require("./platform.js");

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const missing = [];

for (const [key, platform] of Object.entries(PLATFORM_MAP)) {
  const binPath = join(root, "npm", "bin", key, platform.binary);

  if (!existsSync(binPath)) {
    missing.push(`${key}: missing ${binPath}`);
    continue;
  }

  const stat = statSync(binPath);
  if (!stat.isFile() || stat.size === 0) {
    missing.push(`${key}: invalid or empty ${binPath}`);
    continue;
  }

  if (platform.goos !== "windows" && (stat.mode & 0o111) === 0) {
    missing.push(`${key}: not executable ${binPath}`);
  }
}

if (missing.length > 0) {
  console.error("npm package binary verification failed:");
  for (const item of missing) {
    console.error(`- ${item}`);
  }
  console.error('Run "npm run build:targets" before packing or publishing.');
  process.exit(1);
}

console.error("npm package binary verification passed.");
