#!/usr/bin/env python
# Assuming ortfodb is available on PATH
import json
from pathlib import Path
from subprocess import run as _run

root = Path(__file__).parent.parent


def run(cmd) -> str:
    out = _run(cmd.split(" "), capture_output=True)
    return out.stdout.decode("utf-8")


for thing in run("./ortfodb schemas").splitlines():
    schema = run(f"./ortfodb schemas {thing.strip()}")
    formatted_schema = json.dumps(json.loads(schema), indent=2)
    (root / "schemas" / f"{thing}.schema.json").write_text(formatted_schema + "\n")
