#!/usr/bin/env python3
import subprocess

usage = (
    subprocess.run(["./ortfodb", "--help"], capture_output=True)
    .stdout.decode("utf-8")
    .strip()
)

with open("tools/_README.md", mode="r", encoding="utf-8") as file:
    readme = file.read()
    readme = readme.replace("<<<<USAGE>>>>", usage)


with open("README.md", mode="w", encoding="utf-8") as file:
    file.write(readme)
