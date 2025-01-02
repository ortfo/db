import * as z from "zod";


export const DetectSchema = z.object({
    "files": z.array(z.string()).optional(),
    "made with": z.array(z.string()).optional(),
    "search": z.array(z.string()).optional(),
});
export type Detect = z.infer<typeof DetectSchema>;

export const TagSchema = z.object({
    "aliases": z.array(z.string()).optional(),
    "description": z.string().optional(),
    "detect": DetectSchema.optional(),
    "learn more at": z.string().optional(),
    "plural": z.string(),
    "singular": z.string(),
});
export type Tag = z.infer<typeof TagSchema>;
