#!/usr/bin/env python
# Assuming ortfodb is available on PATH
from pathlib import Path
from subprocess import run

root = Path(__file__).parent.parent

for thing in ["tags", "technologies", "configuration", "database"]:
	out = run(["./ortfodb", "schemas", thing], capture_output=True)
	schema = out.stdout.decode("utf-8")
	(root / "schemas" / f"{thing}.schema.json").write_text(schema)
