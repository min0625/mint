"use strict";

const { existsSync } = require("fs");
const { join } = require("path");

const PLATFORM_MAP = {
  "linux-x64":    "mint-ai-linux-x64",
  "linux-arm64":  "mint-ai-linux-arm64",
  "darwin-x64":   "mint-ai-darwin-x64",
  "darwin-arm64": "mint-ai-darwin-arm64",
  "win32-x64":    "mint-ai-windows-x64",
  "win32-arm64":  "mint-ai-windows-arm64",
};

function getBinaryPath() {
  const key = `${process.platform}-${process.arch}`;
  const pkgName = PLATFORM_MAP[key];

  if (!pkgName) {
    throw new Error(
      `mint-ai: unsupported platform ${key}\n` +
      `Please open an issue at https://github.com/min0625/mint`
    );
  }

  const binName = process.platform === "win32" ? "mint.exe" : "mint";

  let pkgDir;
  try {
    pkgDir = require.resolve(`${pkgName}/package.json`);
    pkgDir = pkgDir.replace(/[\\/]package\.json$/, "");
  } catch {
    throw new Error(
      `mint-ai: platform package ${pkgName} not found.\n` +
      `Try reinstalling: npm install -g mint-ai`
    );
  }

  const bin = join(pkgDir, "bin", binName);
  if (!existsSync(bin)) {
    throw new Error(`mint-ai: binary not found at ${bin}`);
  }
  return bin;
}

module.exports = { getBinaryPath };
