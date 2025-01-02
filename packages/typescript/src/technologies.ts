import * as z from "zod";


export const TechnologySchema = z.object({
    "aliases": z.array(z.string()).optional(),
    "autodetect": z.array(z.string()).optional(),
    "by": z.string().optional(),
    "description": z.string().optional(),
    "files": z.array(z.string()).optional(),
    "learn more at": z.string().optional(),
    "name": z.string(),
    "slug": z.string(),
});
export type Technology = z.infer<typeof TechnologySchema>;
