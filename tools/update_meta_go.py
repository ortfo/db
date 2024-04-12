#!/usr/bin/env python
from pathlib import Path
import re
import sys

version = sys.argv[1]
if not version:
    print("Version not provided")
    sys.exit(1)

root = Path(__file__).parent.parent
metafile = root / "meta.go"

version_declaration = re.compile(r"const Version = \"\d+\.\d+\.\d+\"")
new_declaration = f'const Version = "{version}"'
print(f"Updating {metafile}: {new_declaration}")
metafile.write_text(version_declaration.sub(new_declaration, metafile.read_text()))

