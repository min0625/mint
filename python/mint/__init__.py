import os
import stat
import subprocess
import sys
from pathlib import Path


def get_binary_path() -> str:
    """Return the path to the bundled mint binary."""
    here = Path(__file__).parent
    binary_name = "mint.exe" if sys.platform == "win32" else "mint"
    binary = here / "bin" / binary_name

    if not binary.exists():
        raise FileNotFoundError(f"mint binary not found at {binary}")

    # Ensure executable bit is set (wheels may strip it)
    if sys.platform != "win32":
        mode = binary.stat().st_mode
        if not (mode & stat.S_IXUSR):
            binary.chmod(mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)

    return str(binary)


def main() -> None:
    """Execute the bundled mint binary, forwarding all arguments."""
    binary = get_binary_path()
    if sys.platform == "win32":
        sys.exit(subprocess.call([binary] + sys.argv[1:]))
    else:
        # exec replaces the current process — zero overhead
        os.execvp(binary, [binary] + sys.argv[1:])
