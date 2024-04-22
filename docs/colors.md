# Primary colors extraction

ortfo/db supports extracting the most proeminent colors from the works' images. This can be useful to give a 🎀 colorful touch 🎀 to your website

## Usage example

<figure>
  <video src="/examples/color-extraction.mp4" autoplay muted loop></video>
  <figcaption>On <a href="https://ewen.works">ewen.works</a></figcaption>
</figure>

## Configuration

Enable it in [`ortfodb.yaml`](/db/configuration.md):

```yaml
extract colors:
  enabled: true
  extract: []
  default files: []
```

### `enabled`

Controls whether ortfo/db will try to extract colors or not

### `extract`

Array of

### `default files`

TODO

## In `database.json`

Each media [content block](/db/your-first-description-file.md#blocks) will have a `colors` object that contains the proeminent colors of the media:

```json
{
  "ideaseed": {
    "content": {
      "en": {
        "blocks": [
          {
            "id": "Sw0WJU8osY",
            "type": "media",
            ...
            "hasSound": false,
            "colors": {// [!code focus]
                "primary": "#FEF3ED",// [!code focus]
                "secondary": "#858585",// [!code focus]
                "tertiary": "#FCC696" // [!code focus]
            }, // [!code focus]
            "thumbnails": {
                "100": "ideaseed/Sw0WJU8osY@100.webp",
            ...
```
