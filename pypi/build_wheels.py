#!/usr/bin/env python3
"""
Build platform-specific Python wheels from goreleaser dist/ binaries.

Usage:
    python build_wheels.py --version 1.2.3 --dist-dir ../dist --out-dir wheels/
"""

import argparse
import platform
import re
import stat
import sys
import tarfile
import textwrap
import zipfile
from pathlib import Path

# Mapping: (goreleaser OS, goreleaser Arch) -> list of Python wheel platform tags
# Linux binaries are CGO_ENABLED=0 (fully static), so the same binary works on
# both glibc and musl environments. We publish both manylinux and musllinux wheels
# from the same binary so Alpine Linux users can also `pipx install mint-ai`.
PLATFORM_MAP: dict[tuple[str, str], list[str]] = {
    ("linux", "amd64"): [
        "manylinux_2_17_x86_64.manylinux2014_x86_64",
        "musllinux_1_2_x86_64",
    ],
    ("linux", "arm64"): [
        "manylinux_2_17_aarch64.manylinux2014_aarch64",
        "musllinux_1_2_aarch64",
    ],
    ("darwin", "amd64"):  ["macosx_10_9_x86_64"],
    ("darwin", "arm64"):  ["macosx_11_0_arm64"],
    ("windows", "amd64"): ["win_amd64"],
    ("windows", "arm64"): ["win_arm64"],
}

# Maps (platform.system(), platform.machine()) → (goos, goarch)
PY_TO_GO_PLATFORM: dict[tuple[str, str], tuple[str, str]] = {
    ("Linux",   "x86_64"):  ("linux",   "amd64"),
    ("Linux",   "aarch64"): ("linux",   "arm64"),
    ("Darwin",  "x86_64"):  ("darwin",  "amd64"),
    ("Darwin",  "arm64"):   ("darwin",  "arm64"),
    ("Windows", "AMD64"):   ("windows", "amd64"),
    ("Windows", "ARM64"):   ("windows", "arm64"),
}


def find_binary(dist_dir: Path, goos: str, goarch: str) -> Path | None:
    """Locate the goreleaser archive for the given platform."""
    archive_stem = f"mint_{goos}_{goarch}"
    for suffix in (".tar.gz", ".zip"):
        archive = dist_dir / f"{archive_stem}{suffix}"
        if archive.exists():
            return archive
    return None


def build_wheel(
    binary_path: Path,
    version: str,
    platform_tag: str,
    out_dir: Path,
    is_windows: bool,
) -> Path:
    """Assemble a single platform wheel (.whl is just a zip)."""
    # PyPI package name (used for `pip install`)
    dist_name = "mint_ai"   # wheel filename uses underscores (PEP 427)
    # Python module name (import and directory name)
    pkg_name = "mint"
    wheel_name = f"{dist_name}-{version}-py3-none-{platform_tag}.whl"
    wheel_path = out_dir / wheel_name

    bin_filename = "mint.exe" if is_windows else "mint"
    record_lines: list[str] = []

    with zipfile.ZipFile(wheel_path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        # 1. binary → mint/bin/<binary>
        arc_bin = f"{pkg_name}/bin/{bin_filename}"
        zf.write(binary_path, arc_bin)
        record_lines.append(f"{arc_bin},,")

        # 2. __init__.py
        init_src = Path(__file__).parent / "mint" / "__init__.py"
        zf.write(init_src, f"{pkg_name}/__init__.py")
        record_lines.append(f"{pkg_name}/__init__.py,,")

        # 3. __main__.py
        main_src = Path(__file__).parent / "mint" / "__main__.py"
        zf.write(main_src, f"{pkg_name}/__main__.py")
        record_lines.append(f"{pkg_name}/__main__.py,,")

        # 4. METADATA
        # Name field uses PyPI package name (mint-ai) with hyphens (normalized format)
        readme_path = Path(__file__).parent.parent / "README.md"
        readme_content = readme_path.read_text(encoding="utf-8") if readme_path.exists() else ""
        metadata = textwrap.dedent(f"""\
            Metadata-Version: 2.1
            Name: mint-ai
            Version: {version}
            Summary: Minimalist AI translation CLI powered by LLMs
            Home-page: https://github.com/min0625/mint
            License: Apache-2.0
            Requires-Python: >=3.8
            Classifier: Programming Language :: Python :: 3
            Classifier: License :: OSI Approved :: Apache Software License
            Description-Content-Type: text/markdown

        """) + readme_content
        dist_info = f"{dist_name}-{version}.dist-info"
        zf.writestr(f"{dist_info}/METADATA", metadata)
        record_lines.append(f"{dist_info}/METADATA,,")

        # 5. WHEEL
        wheel_meta = textwrap.dedent(f"""\
            Wheel-Version: 1.0
            Generator: build_wheels.py
            Root-Is-Purelib: false
            Tag: py3-none-{platform_tag}
        """)
        zf.writestr(f"{dist_info}/WHEEL", wheel_meta)
        record_lines.append(f"{dist_info}/WHEEL,,")

        # 6. entry_points.txt — left key `mint` determines the terminal command name
        # right `mint:main` points to the main() function in Python module mint
        ep = "[console_scripts]\nmint = mint:main\n"
        zf.writestr(f"{dist_info}/entry_points.txt", ep)
        record_lines.append(f"{dist_info}/entry_points.txt,,")

        # 7. RECORD (added last)
        record_lines.append(f"{dist_info}/RECORD,,")
        zf.writestr(f"{dist_info}/RECORD", "\n".join(record_lines) + "\n")

    print(f"  Built {wheel_path.name}")
    return wheel_path


def normalize_version(version: str) -> str:
    """Convert a version string to PEP 440 format.

    Strips the leading 'v' and converts semver pre-release suffixes:
      v0.0.0-alpha.4  -> 0.0.0a4
      v0.0.0-beta.2   -> 0.0.0b2
      v0.0.0-rc.1     -> 0.0.0rc1
    """
    v = version.lstrip("v")
    v = re.sub(r"-alpha\.(\d+)$", r"a\1", v)
    v = re.sub(r"-beta\.(\d+)$", r"b\1", v)
    v = re.sub(r"-rc\.(\d+)$", r"rc\1", v)
    return v


def main() -> None:
    parser = argparse.ArgumentParser(description="Build mint Python wheels")
    parser.add_argument("--version", required=True, help="Release version, e.g. 1.2.3")
    parser.add_argument("--dist-dir", default="../dist", help="goreleaser dist/ dir")
    parser.add_argument("--out-dir", default="wheels", help="Output dir for .whl files")
    parser.add_argument(
        "--current-platform-only",
        action="store_true",
        help="Only build wheel for current platform (useful for local testing)",
    )
    args = parser.parse_args()

    dist_dir = Path(args.dist_dir).resolve()
    out_dir = Path(args.out_dir).resolve()
    out_dir.mkdir(parents=True, exist_ok=True)
    version = normalize_version(args.version)

    current_goos = None
    current_goarch = None
    if args.current_platform_only:
        system = platform.system()
        machine = platform.machine()
        go_platform = PY_TO_GO_PLATFORM.get((system, machine))
        if go_platform:
            current_goos, current_goarch = go_platform
            print(f"Building only for current platform: {system} {machine} ({current_goos}/{current_goarch})")
        else:
            print(f"Warning: Could not map current platform ({system} {machine}) to Go platform", file=sys.stderr)

    for (goos, goarch), py_tags in PLATFORM_MAP.items():
        if args.current_platform_only and (goos != current_goos or goarch != current_goarch):
            continue

        archive = find_binary(dist_dir, goos, goarch)
        if archive is None:
            print(f"  SKIP {goos}/{goarch}: archive not found in {dist_dir}")
            continue

        bin_filename = "mint.exe" if goos == "windows" else "mint"
        tmp_bin = out_dir / f"_tmp_{goos}_{goarch}_{bin_filename}"

        try:
            # Extract binary directly from archive (avoids path-traversal risk)
            if archive.suffix == ".zip":
                with zipfile.ZipFile(archive) as zf:
                    tmp_bin.write_bytes(zf.read(bin_filename))
            else:
                with tarfile.open(archive, "r:gz") as tf:
                    member = tf.getmember(bin_filename)
                    with tf.extractfile(member) as f:  # type: ignore[union-attr]
                        tmp_bin.write_bytes(f.read())

            if goos != "windows":
                tmp_bin.chmod(tmp_bin.stat().st_mode | stat.S_IXUSR)

            for py_tag in py_tags:
                build_wheel(
                    binary_path=tmp_bin,
                    version=version,
                    platform_tag=py_tag,
                    out_dir=out_dir,
                    is_windows=(goos == "windows"),
                )
        finally:
            tmp_bin.unlink(missing_ok=True)

    print(f"\nDone. Wheels written to {out_dir}/")


if __name__ == "__main__":
    main()
