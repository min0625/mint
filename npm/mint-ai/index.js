"use strict";

const { existsSync } = require("fs");
const { join } = require("path");

const PLATFORM_MAP = {
  "linux-x64":    "linux-x64",
  "linux-arm64":  "linux-arm64",
  "darwin-x64":   "darwin-x64",
  "darwin-arm64": "darwin-arm64",
  "win32-x64":    "windows-x64",
  "win32-arm64":  "windows-arm64",
};

function getBinaryPath() {
  const key = `${process.platform}-${process.arch}`;
  const platformDir = PLATFORM_MAP[key];

  if (!platformDir) {
    throw new Error(
      `mint-ai: unsupported platform ${key}\n` +
      `Please open an issue at https://github.com/min0625/mint`
    );
  }

  const binName = process.platform === "win32" ? "mint.exe" : "mint";
  const bin = join(__dirname, "scripts", platformDir, binName);
  if (!existsSync(bin)) {
    throw new Error(
      `mint-ai: bundled binary for ${key} not found at ${bin}.\n` +
      `Try reinstalling: npm install -g mint-ai`
    );
  }
  return bin;
}

module.exports = { getBinaryPath };
