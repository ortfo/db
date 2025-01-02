import * as z from "zod";


export const ExporterCommandSchema = z.object({
    "log": z.array(z.string()).optional(),
    "run": z.string().optional(),
});
export type ExporterCommand = z.infer<typeof ExporterCommandSchema>;

export const ExporterSchema = z.object({
    "after": z.array(ExporterCommandSchema).optional(),
    "before": z.array(ExporterCommandSchema).optional(),
    "data": z.record(z.string(), z.any()).optional(),
    "description": z.string(),
    "name": z.string(),
    "requires": z.array(z.string()).optional(),
    "verbose": z.boolean().optional(),
    "work": z.array(ExporterCommandSchema).optional(),
});
export type Exporter = z.infer<typeof ExporterSchema>;
