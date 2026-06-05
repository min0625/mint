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

MAIN_PKG_NAME = "mint-ai"

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


def build_main_package(main_src: Path, dist_dir: Path, out_dir: Path, version: str) -> list[str]:
    pkg_dir = out_dir / "main"
    if pkg_dir.exists():
        shutil.rmtree(pkg_dir)
    shutil.copytree(main_src, pkg_dir)

    bundled = []
    for goos, goarch, _npm_os, _npm_cpu, suffix in PLATFORM_MAP:
        bin_name = "mint.exe" if goos == "windows" else "mint"
        bin_path = pkg_dir / "bin" / suffix / bin_name
        bin_path.parent.mkdir(parents=True, exist_ok=True)
        if not extract_binary(dist_dir, goos, goarch, bin_path):
            print(f"  SKIP {suffix}: archive not found in {dist_dir}")
            continue
        if goos != "windows":
            bin_path.chmod(bin_path.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
        bundled.append(suffix)

    if not bundled:
        raise RuntimeError(f"no release archives found in {dist_dir}")

    # Inject version and remove platform package dependencies
    pkg_json_path = pkg_dir / "package.json"
    pkg_json = json.loads(pkg_json_path.read_text())
    pkg_json["version"] = version
    pkg_json.pop("optionalDependencies", None)
    pkg_json_path.write_text(json.dumps(pkg_json, indent=2))

    # Copy README into the package so it appears on the npm registry page
    repo_root = Path(__file__).parent.parent
    readme_src = repo_root / "README.md"
    if readme_src.exists():
        shutil.copy2(readme_src, pkg_dir / "README.md")

    # Ensure shim is executable
    shim = pkg_dir / "bin" / "mint"
    if shim.exists():
        shim.chmod(shim.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)

    print(f"  Built {MAIN_PKG_NAME} (main)")
    return bundled


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

    print("Building npm package...")
    bundled = build_main_package(main_src, dist_dir, out_dir, version)

    print(f"\nDone. Output: {out_dir}/")
    print("\nBundled platform binaries:")
    for suffix in bundled:
        print(f"  {suffix}")
    print("\nPublish command:")
    print(f"  npm publish {out_dir}/main --access public")


if __name__ == "__main__":
    main()
