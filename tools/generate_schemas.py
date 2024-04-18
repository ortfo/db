#!/usr/bin/env python
import json
from pathlib import Path
from subprocess import run as _run

root = Path(__file__).parent.parent


def run(cmd) -> str:
    out = _run(cmd.split(" "), capture_output=True)
    return out.stdout.decode("utf-8")


def add_titles(schema: dict) -> dict:
    """
    Adds the title property to every $defs'd type, by setting them to the type's name
    Needed for client libraries generation, see https://github.com/glideapps/quicktype/blob/master/FAQ.md#why-do-my-types-have-weird-names
    """

    return {
        **schema,
        "$defs": {
            name: {**typedef, "title": name}
            for name, typedef in schema.get("$defs", {}).items()
        },
    }


for thing in run("./ortfodb schemas").splitlines():
    if not thing:
        continue
    raw_schema = run(f"./ortfodb schemas {thing.strip()}")
    try:
        schema = json.loads(raw_schema)
    except:
        print(f"Failed to parse schema for {thing}")
        print(raw_schema)
        continue

    schema = add_titles(schema)
    formatted_schema = json.dumps(schema, indent=2)
    (root / "schemas" / f"{thing}.schema.json").write_text(formatted_schema + "\n")
