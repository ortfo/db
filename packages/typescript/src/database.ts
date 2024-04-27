// To parse this data:
//
//   import { Convert } from "./file";
//
//   const database = Convert.toDatabase(json);
//
// These functions will throw an error if the JSON doesn't
// match the expected interface, even if the JSON is valid.

/**
 * Work represents a given work in the database.
 */
export interface Database {
    builtAt:         Date;
    content:         { [key: string]: LocalizedContent };
    descriptionHash: string;
    id:              string;
    metadata:        WorkMetadata;
    Partial:         boolean;
}

export interface LocalizedContent {
    abbreviations: { [key: string]: string };
    blocks:        ContentBlock[];
    footnotes:     { [key: string]: string };
    layout:        Array<string[]>;
    title:         string;
}

export interface ContentBlock {
    alt: string;
    /**
     * whether the media has been analyzed
     */
    analyzed:   boolean;
    anchor:     string;
    attributes: MediaAttributes;
    caption:    string;
    colors:     ColorPalette;
    /**
     * html
     */
    content:     string;
    contentType: string;
    dimensions:  ImageDimensions;
    distSource:  string;
    /**
     * in seconds
     */
    duration: number;
    /**
     * Hash of the media file, used for caching purposes. Could also serve as an integrity
     * check.
     * The value is the MD5 hash, base64-encoded.
     */
    hash:           string;
    hasSound:       boolean;
    id:             string;
    index:          number;
    online:         boolean;
    relativeSource: string;
    /**
     * in bytes
     */
    size:              number;
    text:              string;
    thumbnails:        ThumbnailsMap;
    thumbnailsBuiltAt: Date;
    title:             string;
    type:              string;
    url:               string;
}

/**
 * MediaAttributes stores which HTML attributes should be added to the media.
 */
export interface MediaAttributes {
    /**
     * Controlled with attribute character > (adds)
     */
    autoplay: boolean;
    /**
     * Controlled with attribute character = (removes)
     */
    controls: boolean;
    /**
     * Controlled with attribute character ~ (adds)
     */
    loop: boolean;
    /**
     * Controlled with attribute character > (adds)
     */
    muted: boolean;
    /**
     * Controlled with attribute character = (adds)
     */
    playsinline: boolean;
}

/**
 * ColorPalette reprensents the object in a Work's metadata.colors.
 */
export interface ColorPalette {
    primary:   string;
    secondary: string;
    tertiary:  string;
}

/**
 * ImageDimensions represents metadata about a media as it's extracted from its file.
 */
export interface ImageDimensions {
    /**
     * width / height
     */
    aspectRatio: number;
    /**
     * Height in pixels
     */
    height: number;
    /**
     * Width in pixels
     */
    width: number;
}

export interface ThumbnailsMap {
}

export interface WorkMetadata {
    additionalMetadata: { [key: string]: any };
    aliases:            string[];
    colors:             ColorPalette;
    databaseMetadata:   DatabaseMeta;
    finished:           string;
    madeWith:           string[];
    pageBackground:     string;
    private:            boolean;
    started:            string;
    tags:               string[];
    thumbnail:          string;
    titleStyle:         string;
    wip:                boolean;
}

export interface DatabaseMeta {
    /**
     * Partial is true if the database was not fully built.
     */
    Partial: boolean;
}

// Converts JSON strings to/from your types
// and asserts the results of JSON.parse at runtime
export class Convert {
    public static toDatabase(json: string): { [key: string]: Database } {
        return cast(JSON.parse(json), m(r("Database")));
    }

    public static databaseToJson(value: { [key: string]: Database }): string {
        return JSON.stringify(uncast(value, m(r("Database"))), null, 2);
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
    "Database": o([
        { json: "builtAt", js: "builtAt", typ: Date },
        { json: "content", js: "content", typ: m(r("LocalizedContent")) },
        { json: "descriptionHash", js: "descriptionHash", typ: "" },
        { json: "id", js: "id", typ: "" },
        { json: "metadata", js: "metadata", typ: r("WorkMetadata") },
        { json: "Partial", js: "Partial", typ: true },
    ], false),
    "LocalizedContent": o([
        { json: "abbreviations", js: "abbreviations", typ: m("") },
        { json: "blocks", js: "blocks", typ: a(r("ContentBlock")) },
        { json: "footnotes", js: "footnotes", typ: m("") },
        { json: "layout", js: "layout", typ: a(a("")) },
        { json: "title", js: "title", typ: "" },
    ], false),
    "ContentBlock": o([
        { json: "alt", js: "alt", typ: "" },
        { json: "analyzed", js: "analyzed", typ: true },
        { json: "anchor", js: "anchor", typ: "" },
        { json: "attributes", js: "attributes", typ: r("MediaAttributes") },
        { json: "caption", js: "caption", typ: "" },
        { json: "colors", js: "colors", typ: r("ColorPalette") },
        { json: "content", js: "content", typ: "" },
        { json: "contentType", js: "contentType", typ: "" },
        { json: "dimensions", js: "dimensions", typ: r("ImageDimensions") },
        { json: "distSource", js: "distSource", typ: "" },
        { json: "duration", js: "duration", typ: 3.14 },
        { json: "hash", js: "hash", typ: "" },
        { json: "hasSound", js: "hasSound", typ: true },
        { json: "id", js: "id", typ: "" },
        { json: "index", js: "index", typ: 0 },
        { json: "online", js: "online", typ: true },
        { json: "relativeSource", js: "relativeSource", typ: "" },
        { json: "size", js: "size", typ: 0 },
        { json: "text", js: "text", typ: "" },
        { json: "thumbnails", js: "thumbnails", typ: r("ThumbnailsMap") },
        { json: "thumbnailsBuiltAt", js: "thumbnailsBuiltAt", typ: Date },
        { json: "title", js: "title", typ: "" },
        { json: "type", js: "type", typ: "" },
        { json: "url", js: "url", typ: "" },
    ], false),
    "MediaAttributes": o([
        { json: "autoplay", js: "autoplay", typ: true },
        { json: "controls", js: "controls", typ: true },
        { json: "loop", js: "loop", typ: true },
        { json: "muted", js: "muted", typ: true },
        { json: "playsinline", js: "playsinline", typ: true },
    ], false),
    "ColorPalette": o([
        { json: "primary", js: "primary", typ: "" },
        { json: "secondary", js: "secondary", typ: "" },
        { json: "tertiary", js: "tertiary", typ: "" },
    ], false),
    "ImageDimensions": o([
        { json: "aspectRatio", js: "aspectRatio", typ: 3.14 },
        { json: "height", js: "height", typ: 0 },
        { json: "width", js: "width", typ: 0 },
    ], false),
    "ThumbnailsMap": o([
    ], false),
    "WorkMetadata": o([
        { json: "additionalMetadata", js: "additionalMetadata", typ: m("any") },
        { json: "aliases", js: "aliases", typ: a("") },
        { json: "colors", js: "colors", typ: r("ColorPalette") },
        { json: "databaseMetadata", js: "databaseMetadata", typ: r("DatabaseMeta") },
        { json: "finished", js: "finished", typ: "" },
        { json: "madeWith", js: "madeWith", typ: a("") },
        { json: "pageBackground", js: "pageBackground", typ: "" },
        { json: "private", js: "private", typ: true },
        { json: "started", js: "started", typ: "" },
        { json: "tags", js: "tags", typ: a("") },
        { json: "thumbnail", js: "thumbnail", typ: "" },
        { json: "titleStyle", js: "titleStyle", typ: "" },
        { json: "wip", js: "wip", typ: true },
    ], false),
    "DatabaseMeta": o([
        { json: "Partial", js: "Partial", typ: true },
    ], false),
};
