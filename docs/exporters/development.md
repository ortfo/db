<script setup>
	import schema from '/schemas/latest/exporter.schema.json';
</script>

# Exporters development guide

## Two types of exporters

ortfo/db has two types of exporters, with different levels of complexity and expressive power:

YAML exporters
: Most of the exporters can be expressed that way. This does not require any development environment setup, but only allows running shell commands.

Go exporters
: For more complex exporters, you can write a Go program that implements the `Exporter` interface. However, for now, Go exporters can only be made available to ortfo by [contributing to the project](https://github.com/ortfo/db).

## YAML exporters

### Bootstrapping

You can quickly initialize an example exporter by running

```shellsession
ortfodb exporters init my-exporter
```

<video src="/db/demo-exporters-init.mp4" muted autoplay controls />

### Configuration

<JSONSchema :schema :headings="4" />

#### Exporter command

<JSONSchema :schema type="ExporterCommand" />

### Example

As an example, this is the manifest for the built-in [SSH exporter](./uploading.md#ssh)

<<< @/ortfodb/exporters/ssh.yaml

## Go exporters

See the Go package documentation for the [`Exporter` interface](https://pkg.go.dev/github.com/ortfo/db/#Exporter).

### Examples

Some examples can be found in ortfo/db's source code:

- the [SQL exporter](https://github.com/ortfo/db/blob/main/exporter_sql.go)
- the [Localize exporter](https://github.com/ortfo/db/blob/main/exporter_localize.go)
