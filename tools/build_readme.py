#!/usr/bin/env python3
import subprocess

usage = (
    subprocess.run(["./ortfodb", "--help"], capture_output=True)
    .stdout.decode("utf-8")
    .strip()
)

def strip_ansi(text):
	import re
	ansi_escape = re.compile(r"\x1B\[[0-?]*[ -/]*[@-~]")
	return ansi_escape.sub("", text)

with open("tools/_README.md", mode="r", encoding="utf-8") as file:
    readme = file.read()
    readme = readme.replace("<<<<USAGE>>>>", strip_ansi(usage))


with open("README.md", mode="w", encoding="utf-8") as file:
    file.write(readme)
