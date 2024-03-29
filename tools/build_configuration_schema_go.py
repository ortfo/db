#!/usr/bin/env python3

with open("ortfodb.yaml.schema.json", encoding="utf-8") as file:
    config_schema = file.read()
with open("database.schema.json", encoding="utf-8") as file:
    database_schema = file.read()


with open("json_schemas.go", mode="w", encoding="utf-8") as file:
    file.write(f"""package ortfodb

// configurationJSONSchema is the entire json string from ortfodb.yaml.json.schema
const configurationJSONSchema = `{config_schema}`

// databaseJSONSchema is the entire json string from database.json.schema
const databaseJSONSchema = `{database_schema}`
""")
