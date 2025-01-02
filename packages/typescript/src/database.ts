import * as z from "zod";


export const MediaAttributesSchema = z.object({
    "autoplay": z.boolean(),
    "controls": z.boolean(),
    "loop": z.boolean(),
    "muted": z.boolean(),
    "playsinline": z.boolean(),
});
export type MediaAttributes = z.infer<typeof MediaAttributesSchema>;

export const ColorPaletteSchema = z.object({
    "primary": z.string(),
    "secondary": z.string(),
    "tertiary": z.string(),
});
export type ColorPalette = z.infer<typeof ColorPaletteSchema>;

export const ImageDimensionsSchema = z.object({
    "aspectRatio": z.number(),
    "height": z.number(),
    "width": z.number(),
});
export type ImageDimensions = z.infer<typeof ImageDimensionsSchema>;

export const ThumbnailsMapSchema = z.object({
});
export type ThumbnailsMap = z.infer<typeof ThumbnailsMapSchema>;

export const DatabaseMetaSchema = z.object({
    "Partial": z.boolean(),
});
export type DatabaseMeta = z.infer<typeof DatabaseMetaSchema>;

export const ContentBlockSchema = z.object({
    "alt": z.string(),
    "analyzed": z.boolean(),
    "anchor": z.string(),
    "attributes": MediaAttributesSchema,
    "caption": z.string(),
    "colors": ColorPaletteSchema,
    "content": z.string(),
    "contentType": z.string(),
    "dimensions": ImageDimensionsSchema,
    "distSource": z.string(),
    "duration": z.number(),
    "hash": z.string(),
    "hasSound": z.boolean(),
    "id": z.string(),
    "index": z.number(),
    "online": z.boolean(),
    "relativeSource": z.string(),
    "size": z.number(),
    "text": z.string(),
    "thumbnails": ThumbnailsMapSchema,
    "thumbnailsBuiltAt": z.coerce.date(),
    "title": z.string(),
    "type": z.string(),
    "url": z.string(),
});
export type ContentBlock = z.infer<typeof ContentBlockSchema>;

export const WorkMetadataSchema = z.object({
    "additionalMetadata": z.record(z.string(), z.any()),
    "aliases": z.array(z.string()),
    "colors": ColorPaletteSchema,
    "databaseMetadata": DatabaseMetaSchema,
    "finished": z.string(),
    "madeWith": z.array(z.string()),
    "pageBackground": z.string(),
    "private": z.boolean(),
    "started": z.string(),
    "tags": z.array(z.string()),
    "thumbnail": z.string(),
    "titleStyle": z.string(),
    "wip": z.boolean(),
});
export type WorkMetadata = z.infer<typeof WorkMetadataSchema>;

export const LocalizedContentSchema = z.object({
    "abbreviations": z.record(z.string(), z.string()),
    "blocks": z.array(ContentBlockSchema),
    "footnotes": z.record(z.string(), z.string()),
    "layout": z.array(z.array(z.string())),
    "title": z.string(),
});
export type LocalizedContent = z.infer<typeof LocalizedContentSchema>;

export const WorkSchema = z.object({
    "builtAt": z.coerce.date(),
    "content": z.record(z.string(), LocalizedContentSchema),
    "descriptionHash": z.string(),
    "id": z.string(),
    "metadata": WorkMetadataSchema,
    "Partial": z.boolean(),
});
export type Work = z.infer<typeof WorkSchema>;
