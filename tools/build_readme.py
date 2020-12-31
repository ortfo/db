#!/usr/bin/env python3

with open("usage.go", encoding="utf-8") as file:
    usage = "\n".join(file.read().split("\n")[4:-1])


with open("tools/_README.md", mode="r", encoding="utf-8") as file:
    readme = file.read()
    readme = readme.replace("<<<<USAGE>>>>", usage)


with open("README.md", mode="w", encoding="utf-8") as file:
    file.write(readme)
