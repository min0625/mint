#!/usr/bin/env python3
"""
Build npm packages from goreleaser dist/ binaries.

Usage:
    python npm/build_npm_packages.py --version 1.2.3 --dist-dir dist --out-dir dist/npm
"""

import argparse
import json
import shutil
import stat
import tarfile
import zipfile
from pathlib import Path

SCOPE = "@min0625"
BASE_NAME = "mint-ai"
MAIN_PKG_NAME = "mint-ai"  # main package has no scope

PLATFORM_MAP = [
    # (goreleaser_os, goreleaser_arch, npm_os,   npm_cpu,  suffix)
    ("linux",   "amd64", "linux",  "x64",   "linux-x64"),
    ("linux",   "arm64", "linux",  "arm64", "linux-arm64"),
    ("darwin",  "amd64", "darwin", "x64",   "darwin-x64"),
    ("darwin",  "arm64", "darwin", "arm64", "darwin-arm64"),
    ("windows", "amd64", "win32",  "x64",   "windows-x64"),
    ("windows", "arm64", "win32",  "arm64", "windows-arm64"),
]


def extract_binary(dist_dir: Path, goos: str, goarch: str, out_path: Path) -> bool:
    """Extract binary from goreleaser archive and write to out_path."""
    bin_name = "mint.exe" if goos == "windows" else "mint"
    archive_stem = f"mint_{goos}_{goarch}"

    for suffix in (".tar.gz", ".zip"):
        archive = dist_dir / f"{archive_stem}{suffix}"
        if not archive.exists():
            continue
        if suffix == ".zip":
            with zipfile.ZipFile(archive) as zf:
                with zf.open(bin_name) as src, open(out_path, "wb") as dst:
                    dst.write(src.read())
        else:
            with tarfile.open(archive, "r:gz") as tf:
                src = tf.extractfile(tf.getmember(bin_name))
                with open(out_path, "wb") as dst:
                    dst.write(src.read())
        return True
    return False


def build_platform_package(
    dist_dir: Path, out_dir: Path, version: str,
    goos: str, goarch: str, npm_os: str, npm_cpu: str, suffix: str,
) -> bool:
    pkg_name = f"{SCOPE}/{BASE_NAME}-{suffix}"
    pkg_dir = out_dir / suffix
    bin_dir = pkg_dir / "bin"
    bin_dir.mkdir(parents=True, exist_ok=True)

    bin_name = "mint.exe" if goos == "windows" else "mint"
    bin_path = bin_dir / bin_name

    if not extract_binary(dist_dir, goos, goarch, bin_path):
        print(f"  SKIP {pkg_name}: archive not found in {dist_dir}")
        return False

    if goos != "windows":
        bin_path.chmod(bin_path.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)

    (pkg_dir / "package.json").write_text(json.dumps({
        "name": pkg_name,
        "version": version,
        "description": f"mint binary for {npm_os} {npm_cpu}",
        "os": [npm_os],
        "cpu": [npm_cpu],
        "bin": {"mint": f"bin/{bin_name}"},
        "license": "Apache-2.0",
        "repository": {"type": "git", "url": "https://github.com/min0625/mint.git"},
    }, indent=2))

    print(f"  Built {pkg_name}")
    return True


def build_main_package(main_src: Path, out_dir: Path, version: str) -> None:
    pkg_dir = out_dir / "main"
    if pkg_dir.exists():
        shutil.rmtree(pkg_dir)
    shutil.copytree(main_src, pkg_dir)

    # Inject version and exact optionalDependencies versions
    pkg_json_path = pkg_dir / "package.json"
    pkg_json = json.loads(pkg_json_path.read_text())
    pkg_json["version"] = version
    pkg_json["optionalDependencies"] = {
        f"{SCOPE}/{BASE_NAME}-{suffix}": version
        for *_, suffix in PLATFORM_MAP
    }
    pkg_json_path.write_text(json.dumps(pkg_json, indent=2))

    # Ensure shim is executable
    shim = pkg_dir / "bin" / "mint"
    if shim.exists():
        shim.chmod(shim.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)

    print(f"  Built {MAIN_PKG_NAME} (main)")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--version", required=True, help="e.g. 1.2.3 or v1.2.3")
    parser.add_argument("--dist-dir", default="dist")
    parser.add_argument("--out-dir", default="dist/npm")
    args = parser.parse_args()

    version = args.version.lstrip("v")
    dist_dir = Path(args.dist_dir).resolve()
    out_dir = Path(args.out_dir).resolve()
    out_dir.mkdir(parents=True, exist_ok=True)
    main_src = Path(__file__).parent / "main"

    print("Building platform packages...")
    built = []
    for goos, goarch, npm_os, npm_cpu, suffix in PLATFORM_MAP:
        ok = build_platform_package(dist_dir, out_dir, version, goos, goarch, npm_os, npm_cpu, suffix)
        if ok:
            built.append(suffix)

    print("\nBuilding main package...")
    build_main_package(main_src, out_dir, version)

    print(f"\nDone. Output: {out_dir}/")
    print("\nPublish order (platform packages must be published BEFORE main):")
    for suffix in built:
        print(f"  npm publish {out_dir}/{suffix} --access public")
    print(f"  npm publish {out_dir}/main --access public")


if __name__ == "__main__":
    main()
