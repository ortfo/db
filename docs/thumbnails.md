# Thumbnail generation

ortfo/db's compiler allows you to automatically generate thumbnails for your media files.

## Configuration

Configuration of thumbnails generation is done via the `make thumbnails` section of the configuration:

```yaml{4-12}
  enabled: false
  file name template: ""

make thumbnails:
  input file: ""
  enabled: true
  sizes:
	- 100
	- 400
	- 600
	- 1200
  file name template: <work id>/<block id>@<size>.webp

media:
```

### `input file`

TODO

### `enabled`

Controls whether thumbnails are generated or not

### `sizes`

Array of sizes to generate thumbnails for. The sizes are in pixels. The generated thumbnail will have its largest side equal to the size specified, while preserving its aspect ratio.

### `file name template`

The template for the file name of the generated thumbnails. The following placeholders are available:

- `<work id>`: The work's identifier
- `<block id>`: The [block](/db/your-first-description-file.md#blocks)'s identifier
- `<size>`: The size of the thumbnail



## Usage

Generated thumbnails will be created in the media directory, using the [file name template](#file-name-template) to determine the full filepath.

The output database.json file will include paths to these files alongside the media blocks:

```json{12-17}
{
	"ideaseed": {
		"content": {
			"en": {
				"blocks": [
					{
						"id": "GBpC-nYDgw",
						"type": "media",
						"anchor": "ideaseed-logo-black-transparent",
						"index": 0,
						...
						"thumbnails": {
							"100": "ideaseed/GBpC-nYDgw@100.webp",
							"400": "ideaseed/GBpC-nYDgw@400.webp",
							"600": "ideaseed/GBpC-nYDgw@600.webp",
							"1200": "ideaseed/GBpC-nYDgw@1200.webp"
						},
						"thumbnailsBuiltAt": "2024-04-14 21:06:29.866092383 +0200 CEST m=+3.902704829",
						...
					},
```

## Image formats

The extension of the file name determines what format the thumbnail will be saved in. As thumbnail generation is handled by [ImageMagick](https://imagemagick.org/index.php), the extension must correspond to one of the [formats supported by ImageMagick](https://imagemagick.org/script/formats.php), [which is _a lot of formats_](./image-formats.md#available-formats).
