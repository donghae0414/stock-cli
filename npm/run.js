#!/usr/bin/env node

"use strict";

const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");
const { getPlatform } = require("./platform");

function repoRoot() {
  return path.resolve(__dirname, "..");
}

function resolveBinary() {
  const platform = getPlatform();
  const packaged = path.join(__dirname, "bin", platform.key, platform.binary);
  if (fs.existsSync(packaged)) return packaged;

  const root = repoRoot();
  const repoBinary = path.join(root, "bin", process.platform === "win32" ? "stock.exe" : "stock");
  const allowRepoFallback = !process.env.STOCK_CLI_DISABLE_REPO_FALLBACK;
  if (allowRepoFallback && fs.existsSync(path.join(root, "go.mod")) && fs.existsSync(repoBinary)) {
    return repoBinary;
  }

  throw new Error(
    `stock binary not found for ${platform.key}.\n` +
    `Run "npm run build:local" in a repository checkout, or publish/install a package that includes npm/bin/${platform.key}/${platform.binary}.`
  );
}

let binPath;
try {
  binPath = resolveBinary();
} catch (err) {
  process.stderr.write(`${err.message}\n`);
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), {
  cwd: process.cwd(),
  stdio: "inherit"
});

if (result.error) {
  process.stderr.write(`Error running stock: ${result.error.message}\n`);
  process.exit(1);
}

process.exit(result.status ?? 1);
