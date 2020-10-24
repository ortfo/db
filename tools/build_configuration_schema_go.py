#!/usr/bin/env python3

with open(".portfoliodb.yml.schema.json", encoding="utf-8") as file:
    schema = file.read()


with open("configuration_schema.go", mode="w", encoding="utf-8") as file:
    file.write(f"""package main

// ConfigurationJSONSchema is the entire json string from configuration.json.schema
const ConfigurationJSONSchema = `{schema}`""")
