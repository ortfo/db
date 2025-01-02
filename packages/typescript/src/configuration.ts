import * as z from "zod";


export const ExtractColorsConfigurationSchema = z.object({
    "default files": z.array(z.string()),
    "enabled": z.boolean(),
    "extract": z.array(z.string()),
});
export type ExtractColorsConfiguration = z.infer<typeof ExtractColorsConfigurationSchema>;

export const MakeGiFsConfigurationSchema = z.object({
    "enabled": z.boolean(),
    "file name template": z.string(),
});
export type MakeGiFsConfiguration = z.infer<typeof MakeGiFsConfigurationSchema>;

export const MakeThumbnailsConfigurationSchema = z.object({
    "enabled": z.boolean(),
    "file name template": z.string(),
    "input file": z.string(),
    "sizes": z.array(z.number()),
});
export type MakeThumbnailsConfiguration = z.infer<typeof MakeThumbnailsConfigurationSchema>;

export const MediaConfigurationSchema = z.object({
    "at": z.string(),
});
export type MediaConfiguration = z.infer<typeof MediaConfigurationSchema>;

export const TagsConfigurationSchema = z.object({
    "repository": z.string(),
});
export type TagsConfiguration = z.infer<typeof TagsConfigurationSchema>;

export const TechnologiesConfigurationSchema = z.object({
    "repository": z.string(),
});
export type TechnologiesConfiguration = z.infer<typeof TechnologiesConfigurationSchema>;

export const ConfigurationSchema = z.object({
    "exporters": z.record(z.string(), z.record(z.string(), z.any())).optional(),
    "extract colors": ExtractColorsConfigurationSchema.optional(),
    "make gifs": MakeGiFsConfigurationSchema.optional(),
    "make thumbnails": MakeThumbnailsConfigurationSchema.optional(),
    "media": MediaConfigurationSchema.optional(),
    "projects at": z.string(),
    "scattered mode folder": z.string(),
    "tags": TagsConfigurationSchema.optional(),
    "technologies": TechnologiesConfigurationSchema.optional(),
});
export type Configuration = z.infer<typeof ConfigurationSchema>;
