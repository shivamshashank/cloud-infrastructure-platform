"""Reject trailing spaces and tabs in text files passed by pre-commit."""

import sys
from pathlib import Path


def main() -> int:
    failures = 0
    for name in sys.argv[1:]:
        for number, line in enumerate(Path(name).read_bytes().splitlines(), start=1):
            if line.endswith((b" ", b"\t")):
                print(f"{name}:{number}: trailing whitespace")
                failures += 1
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main())
