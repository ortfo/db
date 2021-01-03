#!/usr/bin/env python3

with open(".portfoliodb.yml.schema.json", encoding="utf-8") as file:
    config_schema = file.read()
with open("database.schema.json", encoding="utf-8") as file:
    database_schema = file.read()


with open("json_schemas.go", mode="w", encoding="utf-8") as file:
    file.write(f"""package main

// ConfigurationJSONSchema is the entire json string from .portfoliodb.yml.json.schema
const ConfigurationJSONSchema = `{config_schema}`

// DatabaseJSONSchema is the entire json string from database.json.schema
const DatabaseJSONSchema = `{database_schema}`
""")
