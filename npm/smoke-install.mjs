#!/usr/bin/env node

import { existsSync, mkdtempSync, rmSync, unlinkSync } from "node:fs";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? root,
    stdio: options.stdio ?? "inherit",
    env: options.env ?? process.env,
    encoding: "utf8"
  });

  if (result.error) {
    throw new Error(`${command} failed: ${result.error.message}`);
  }

  if ((result.status ?? 1) !== 0) {
    throw new Error(`${command} ${args.join(" ")} exited ${result.status ?? 1}`);
  }

  return result;
}

const npmCache = mkdtempSync(join(tmpdir(), "stock-cli-npm-cache-"));
const npmEnv = {
  ...process.env,
  npm_config_cache: npmCache
};

let tarball;
let tmp;

try {
  const pack = run("npm", ["pack", "--json", "--dry-run=false"], { stdio: "pipe", env: npmEnv });
  const jsonStart = pack.stdout.lastIndexOf("[\n  {");
  if (jsonStart < 0) {
    throw new Error("npm pack did not return a JSON array");
  }

  const packed = JSON.parse(pack.stdout.slice(jsonStart))[0];
  if (!packed?.filename) {
    throw new Error("npm pack did not return a tarball filename");
  }

  tarball = join(root, packed.filename);
  tmp = mkdtempSync(join(tmpdir(), "stock-cli-install-"));

  run("npm", ["install", tarball, "--no-audit", "--ignore-scripts", "--dry-run=false"], { cwd: tmp, env: npmEnv });

  const stockBin = process.platform === "win32"
    ? join(tmp, "node_modules", ".bin", "stock.cmd")
    : join(tmp, "node_modules", ".bin", "stock");

  if (!existsSync(stockBin)) {
    throw new Error(`Installed stock shim not found at ${stockBin}`);
  }

  run(stockBin, ["--help"], {
    cwd: tmp,
    env: {
      ...process.env,
      STOCK_CLI_DISABLE_REPO_FALLBACK: "1"
    }
  });
} finally {
  if (tmp) rmSync(tmp, { recursive: true, force: true });
  if (tarball && existsSync(tarball)) unlinkSync(tarball);
  rmSync(npmCache, { recursive: true, force: true });
}

console.log("Installed package smoke test passed.");
