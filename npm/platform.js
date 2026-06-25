#!/usr/bin/env node

"use strict";

const PLATFORM_MAP = {
  "darwin-arm64": { goos: "darwin", goarch: "arm64", binary: "stock" },
  "darwin-x64": { goos: "darwin", goarch: "amd64", binary: "stock" },
  "linux-arm64": { goos: "linux", goarch: "arm64", binary: "stock" },
  "linux-x64": { goos: "linux", goarch: "amd64", binary: "stock" },
  "win32-x64": { goos: "windows", goarch: "amd64", binary: "stock.exe" }
};

function getPlatformKey() {
  return `${process.platform}-${process.arch}`;
}

function getPlatform(key = getPlatformKey()) {
  const entry = PLATFORM_MAP[key];
  if (!entry) {
    throw new Error(
      `Unsupported platform: ${process.platform} ${process.arch}\n` +
      `Supported: ${Object.keys(PLATFORM_MAP).join(", ")}`
    );
  }
  return { key, ...entry };
}

module.exports = { PLATFORM_MAP, getPlatform, getPlatformKey };
