// To parse this data:
//
//   import { Convert, Configuration } from "./file";
//
//   const configuration = Convert.toConfiguration(json);
//
// These functions will throw an error if the JSON doesn't
// match the expected interface, even if the JSON is valid.

/**
 * Configuration represents what the ortfodb.yaml configuration file describes.
 */
export interface Configuration {
    /**
     * Exporter-specific configuration. Maps exporter names to their configuration.
     */
    exporters?:         { [key: string]: { [key: string]: any } };
    "extract colors"?:  ExtractColorsConfiguration;
    "make gifs"?:       MakeGIFSConfiguration;
    "make thumbnails"?: MakeThumbnailsConfiguration;
    media?:             MediaConfiguration;
    /**
     * Path to the directory containing all projects. Must be absolute.
     */
    "projects at":           string;
    "scattered mode folder": string;
    tags?:                   TagsConfiguration;
    technologies?:           TechnologiesConfiguration;
}

export interface ExtractColorsConfiguration {
    "default files": string[];
    enabled:         boolean;
    extract:         string[];
}

export interface MakeGIFSConfiguration {
    enabled:              boolean;
    "file name template": string;
}

export interface MakeThumbnailsConfiguration {
    enabled:              boolean;
    "file name template": string;
    "input file":         string;
    sizes:                number[];
}

export interface MediaConfiguration {
    /**
     * Path to the media directory.
     */
    at: string;
}

export interface TagsConfiguration {
    /**
     * Path to file describing all tags.
     */
    repository: string;
}

export interface TechnologiesConfiguration {
    /**
     * Path to file describing all technologies.
     */
    repository: string;
}

// Converts JSON strings to/from your types
// and asserts the results of JSON.parse at runtime
export class Convert {
    public static toConfiguration(json: string): Configuration {
        return cast(JSON.parse(json), r("Configuration"));
    }

    public static configurationToJson(value: Configuration): string {
        return JSON.stringify(uncast(value, r("Configuration")), null, 2);
    }
}

function invalidValue(typ: any, val: any, key: any, parent: any = ''): never {
    const prettyTyp = prettyTypeName(typ);
    const parentText = parent ? ` on ${parent}` : '';
    const keyText = key ? ` for key "${key}"` : '';
    throw Error(`Invalid value${keyText}${parentText}. Expected ${prettyTyp} but got ${JSON.stringify(val)}`);
}

function prettyTypeName(typ: any): string {
    if (Array.isArray(typ)) {
        if (typ.length === 2 && typ[0] === undefined) {
            return `an optional ${prettyTypeName(typ[1])}`;
        } else {
            return `one of [${typ.map(a => { return prettyTypeName(a); }).join(", ")}]`;
        }
    } else if (typeof typ === "object" && typ.literal !== undefined) {
        return typ.literal;
    } else {
        return typeof typ;
    }
}

function jsonToJSProps(typ: any): any {
    if (typ.jsonToJS === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.json] = { key: p.js, typ: p.typ });
        typ.jsonToJS = map;
    }
    return typ.jsonToJS;
}

function jsToJSONProps(typ: any): any {
    if (typ.jsToJSON === undefined) {
        const map: any = {};
        typ.props.forEach((p: any) => map[p.js] = { key: p.json, typ: p.typ });
        typ.jsToJSON = map;
    }
    return typ.jsToJSON;
}

function transform(val: any, typ: any, getProps: any, key: any = '', parent: any = ''): any {
    function transformPrimitive(typ: string, val: any): any {
        if (typeof typ === typeof val) return val;
        return invalidValue(typ, val, key, parent);
    }

    function transformUnion(typs: any[], val: any): any {
        // val must validate against one typ in typs
        const l = typs.length;
        for (let i = 0; i < l; i++) {
            const typ = typs[i];
            try {
                return transform(val, typ, getProps);
            } catch (_) {}
        }
        return invalidValue(typs, val, key, parent);
    }

    function transformEnum(cases: string[], val: any): any {
        if (cases.indexOf(val) !== -1) return val;
        return invalidValue(cases.map(a => { return l(a); }), val, key, parent);
    }

    function transformArray(typ: any, val: any): any {
        // val must be an array with no invalid elements
        if (!Array.isArray(val)) return invalidValue(l("array"), val, key, parent);
        return val.map(el => transform(el, typ, getProps));
    }

    function transformDate(val: any): any {
        if (val === null) {
            return null;
        }
        const d = new Date(val);
        if (isNaN(d.valueOf())) {
            return invalidValue(l("Date"), val, key, parent);
        }
        return d;
    }

    function transformObject(props: { [k: string]: any }, additional: any, val: any): any {
        if (val === null || typeof val !== "object" || Array.isArray(val)) {
            return invalidValue(l(ref || "object"), val, key, parent);
        }
        const result: any = {};
        Object.getOwnPropertyNames(props).forEach(key => {
            const prop = props[key];
            const v = Object.prototype.hasOwnProperty.call(val, key) ? val[key] : undefined;
            result[prop.key] = transform(v, prop.typ, getProps, key, ref);
        });
        Object.getOwnPropertyNames(val).forEach(key => {
            if (!Object.prototype.hasOwnProperty.call(props, key)) {
                result[key] = transform(val[key], additional, getProps, key, ref);
            }
        });
        return result;
    }

    if (typ === "any") return val;
    if (typ === null) {
        if (val === null) return val;
        return invalidValue(typ, val, key, parent);
    }
    if (typ === false) return invalidValue(typ, val, key, parent);
    let ref: any = undefined;
    while (typeof typ === "object" && typ.ref !== undefined) {
        ref = typ.ref;
        typ = typeMap[typ.ref];
    }
    if (Array.isArray(typ)) return transformEnum(typ, val);
    if (typeof typ === "object") {
        return typ.hasOwnProperty("unionMembers") ? transformUnion(typ.unionMembers, val)
            : typ.hasOwnProperty("arrayItems")    ? transformArray(typ.arrayItems, val)
            : typ.hasOwnProperty("props")         ? transformObject(getProps(typ), typ.additional, val)
            : invalidValue(typ, val, key, parent);
    }
    // Numbers can be parsed by Date but shouldn't be.
    if (typ === Date && typeof val !== "number") return transformDate(val);
    return transformPrimitive(typ, val);
}

function cast<T>(val: any, typ: any): T {
    return transform(val, typ, jsonToJSProps);
}

function uncast<T>(val: T, typ: any): any {
    return transform(val, typ, jsToJSONProps);
}

function l(typ: any) {
    return { literal: typ };
}

function a(typ: any) {
    return { arrayItems: typ };
}

function u(...typs: any[]) {
    return { unionMembers: typs };
}

function o(props: any[], additional: any) {
    return { props, additional };
}

function m(additional: any) {
    return { props: [], additional };
}

function r(name: string) {
    return { ref: name };
}

const typeMap: any = {
    "Configuration": o([
        { json: "exporters", js: "exporters", typ: u(undefined, m(m("any"))) },
        { json: "extract colors", js: "extract colors", typ: u(undefined, r("ExtractColorsConfiguration")) },
        { json: "make gifs", js: "make gifs", typ: u(undefined, r("MakeGIFSConfiguration")) },
        { json: "make thumbnails", js: "make thumbnails", typ: u(undefined, r("MakeThumbnailsConfiguration")) },
        { json: "media", js: "media", typ: u(undefined, r("MediaConfiguration")) },
        { json: "projects at", js: "projects at", typ: "" },
        { json: "scattered mode folder", js: "scattered mode folder", typ: "" },
        { json: "tags", js: "tags", typ: u(undefined, r("TagsConfiguration")) },
        { json: "technologies", js: "technologies", typ: u(undefined, r("TechnologiesConfiguration")) },
    ], false),
    "ExtractColorsConfiguration": o([
        { json: "default files", js: "default files", typ: a("") },
        { json: "enabled", js: "enabled", typ: true },
        { json: "extract", js: "extract", typ: a("") },
    ], false),
    "MakeGIFSConfiguration": o([
        { json: "enabled", js: "enabled", typ: true },
        { json: "file name template", js: "file name template", typ: "" },
    ], false),
    "MakeThumbnailsConfiguration": o([
        { json: "enabled", js: "enabled", typ: true },
        { json: "file name template", js: "file name template", typ: "" },
        { json: "input file", js: "input file", typ: "" },
        { json: "sizes", js: "sizes", typ: a(0) },
    ], false),
    "MediaConfiguration": o([
        { json: "at", js: "at", typ: "" },
    ], false),
    "TagsConfiguration": o([
        { json: "repository", js: "repository", typ: "" },
    ], false),
    "TechnologiesConfiguration": o([
        { json: "repository", js: "repository", typ: "" },
    ], false),
};
