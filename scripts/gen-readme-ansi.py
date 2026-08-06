#!/usr/bin/env python3
"""Regenerate README.ansi from README.md.

README.md is the single source of truth. This script re-renders its body as a
colourised terminal document written with real ESC (0x1B) bytes, so that
`less -R README.ansi` and `make readme | less -R` display colour instead of
literal "\\x1b[36m" text. It also refreshes the ```ansi preview block at the
top of README.md so the two never drift apart.

Usage:
    python3 scripts/gen-readme-ansi.py [--check]

--check exits non-zero if either file is out of date instead of writing.
"""

from __future__ import annotations

import argparse
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
README_MD = ROOT / "README.md"
README_ANSI = ROOT / "README.ansi"

ESC = "\x1b"
CYAN = f"{ESC}[36m"
GREEN = f"{ESC}[32m"
GREY = f"{ESC}[90m"
RESET = f"{ESC}[0m"

RULE = "-" * 60
BANNER = [
    "███████╗██████╗ ███████╗██╗  ██╗███╗   ███╗ ██████╗ ███╗   ██╗",
    "██╔════╝██╔══██╗██╔════╝╚██╗██╔╝████╗ ████║██╔═══██╗████╗  ██║",
    "███████╗██║  ██║█████╗   ╚███╔╝ ██╔████╔██║██║   ██║██╔██╗ ██║",
    "╚════██║██║  ██║██╔══╝   ██╔██╗ ██║╚██╔╝██║██║   ██║██║╚██╗██║",
    "███████║██████╔╝███████╗██╔╝ ██╗██║ ╚═╝ ██║╚██████╔╝██║ ╚████║",
    "╚══════╝╚═════╝ ╚══════╝╚═╝  ╚═╝╚═╝         ╚═════╝ ╚═╝  ╚═══╝",
]

# "[ 3 ] QUICKSTART" and friends. Only counts as a heading when the next line
# underlines it, which is what separates real headings from the table of
# contents near the top of README.md.
HEADING_RE = re.compile(r"^\[ (\d+) \] (.+)$")
UNDERLINE_RE = re.compile(r"^-{3,}$")
RULE_RE = re.compile(r"^[-=]{10,}$")


def colour_banner() -> list[str]:
    lines = list(BANNER)
    lines[0] = CYAN + lines[0]
    lines[-1] = lines[-1] + RESET
    return lines


def parse_readme(text: str) -> tuple[str, list[str], list[tuple[str, list[str]]], str]:
    """Split README.md into (subtitle, intro, sections, tagline)."""
    lines = text.split("\n")

    try:
        start = lines.index("# SDEXMON")
    except ValueError as exc:  # pragma: no cover - guards against a gutted README
        raise SystemExit("README.md: could not find the '# SDEXMON' title") from exc

    subtitle = lines[start + 1]
    body = lines[start + 2 :]

    # Drop the '======' underline that follows the subtitle.
    while body and (RULE_RE.match(body[0]) or body[0] == ""):
        body.pop(0)

    tagline = ""
    for line in reversed(body):
        if line.strip():
            if line.startswith("SDEXMON"):
                tagline = line
            break

    intro: list[str] = []
    sections: list[tuple[str, list[str]]] = []
    current: list[str] | None = None

    index = 0
    while index < len(body):
        line = body[index]
        nxt = body[index + 1] if index + 1 < len(body) else ""

        if HEADING_RE.match(line) and UNDERLINE_RE.match(nxt):
            sections.append((line, []))
            current = sections[-1][1]
            index += 2
            continue

        # Table-of-contents entries and horizontal rules are re-created by the
        # renderer, so they are never carried over verbatim.
        if HEADING_RE.match(line) and current is None:
            index += 1
            continue
        if RULE_RE.match(line):
            index += 1
            continue
        if line == tagline and tagline:
            index += 1
            continue

        (intro if current is None else current).append(line)
        index += 1

    return subtitle, trim(intro), [(h, trim(b)) for h, b in sections], tagline


def trim(lines: list[str]) -> list[str]:
    while lines and not lines[0].strip():
        lines.pop(0)
    while lines and not lines[-1].strip():
        lines.pop()
    return lines


def render_heading(heading: str) -> list[str]:
    match = HEADING_RE.match(heading)
    assert match is not None
    number, title = match.group(1), match.group(2)
    return [
        f"{GREEN}[ {number} ]{RESET} {CYAN}{title}{RESET}",
        f"{GREY}{'-' * len(heading)}{RESET}",
    ]


def render_body(lines: list[str]) -> list[str]:
    out = []
    for index, line in enumerate(lines):
        # Highlight sub-headings: a plain line that introduces a bullet list
        # without ending in a colon, e.g. the per-screen key groups in USAGE.
        following = next((l for l in lines[index + 1 :] if l.strip()), "")
        is_subheading = (
            line.strip()
            and not line.startswith((" ", "-", "#"))
            and not line.rstrip().endswith(":")
            and following.startswith("- ")
        )
        out.append(f"{CYAN}{line}{RESET}" if is_subheading else line)
    return out


def render_ansi(subtitle: str, intro, sections, tagline: str) -> str:
    out = colour_banner()
    out.append(f"{CYAN}{subtitle}{RESET}")
    out.append(f"{GREY}{'=' * 60}{RESET}")
    out.append("")
    out.extend(intro)

    for heading, body in sections:
        out.extend(["", f"{GREY}{RULE}{RESET}", ""])
        out.extend(render_heading(heading))
        out.extend(render_body(body))

    out.extend(["", f"{GREY}{RULE}{RESET}"])
    if tagline:
        out.append(f"{CYAN}{tagline}{RESET}")
    return "\n".join(out) + "\n"


COLUMNS = 3


def render_preview(subtitle: str, sections) -> str:
    """The short ```ansi block shown at the top of README.md."""
    out = colour_banner()
    out.append("")
    out.append(f"{CYAN}{subtitle}{RESET}")
    out.append(f"{GREY}{RULE}{RESET}")

    titles = []
    for heading, _ in sections:
        match = HEADING_RE.match(heading)
        assert match is not None
        titles.append((match.group(1), match.group(2)))

    # Pad on visible length so the columns line up despite the escape codes.
    widths = [
        max((len(t) for i, (_, t) in enumerate(titles) if i % COLUMNS == column), default=0)
        for column in range(COLUMNS)
    ]
    entries = [
        f"{GREEN}[ {number} ]{RESET} {CYAN}{title}{RESET}"
        + " " * (widths[index % COLUMNS] - len(title))
        for index, (number, title) in enumerate(titles)
    ]
    for row in range(0, len(entries), COLUMNS):
        out.append("   ".join(entries[row : row + COLUMNS]).rstrip())

    out.append(f"{GREY}{RULE}{RESET}")
    out.append(
        f"{CYAN}Terminal view:{RESET} {GREEN}less -R README.ansi{RESET}"
        f"  {GREY}|{RESET}  {GREEN}make readme | less -R{RESET}"
    )
    return "\n".join(out)


def replace_preview(text: str, preview: str) -> str:
    pattern = re.compile(r"(```ansi\n).*?(\n```)", re.DOTALL)
    if not pattern.search(text):
        raise SystemExit("README.md: could not find the ```ansi preview block")
    return pattern.sub(lambda _: f"```ansi\n{preview}\n```", text, count=1)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="verify without writing")
    args = parser.parse_args()

    source = README_MD.read_text(encoding="utf-8")
    subtitle, intro, sections, tagline = parse_readme(source)

    ansi = render_ansi(subtitle, intro, sections, tagline)
    updated_md = replace_preview(source, render_preview(subtitle, sections))

    if args.check:
        stale = []
        if not README_ANSI.exists() or README_ANSI.read_text(encoding="utf-8") != ansi:
            stale.append("README.ansi")
        if updated_md != source:
            stale.append("README.md preview block")
        if stale:
            print(
                "out of date: " + ", ".join(stale) + "\nrun: make readme-gen",
                file=sys.stderr,
            )
            return 1
        print("README.ansi is up to date")
        return 0

    README_ANSI.write_text(ansi, encoding="utf-8")
    if updated_md != source:
        README_MD.write_text(updated_md, encoding="utf-8")
    print(f"wrote {README_ANSI.relative_to(ROOT)} ({len(sections)} sections)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
