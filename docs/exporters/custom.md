<script setup>
  import schema from '/schemas/latest/exporter.schema.json';
</script>
# The `custom` exporter

A convinient way to run your own shell commands at various stages of the build process, without having to [create your own](./development.md) exporter in a separate YAML file.

## Configuration

The configuration of the custom exporter corresponds to the fields used to [create an exporter](./development.md#yaml-exporters):

<JSONSchema :schema :pick="['after', 'before', 'export']" />

### Exporter command

<JSONSchema :schema type="ExporterCommand" />

### Example

```yaml
exporters:
  custom:
    after:
      - run: echo {{ .Database | len }} works in the resulting database
