from typing import Any, Dict, List, TypeVar, Type, cast, Callable
from datetime import datetime
import dateutil.parser


T = TypeVar("T")


def from_bool(x: Any) -> bool:
    assert isinstance(x, bool)
    return x


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


def from_float(x: Any) -> float:
    assert isinstance(x, (float, int)) and not isinstance(x, bool)
    return float(x)


def from_int(x: Any) -> int:
    assert isinstance(x, int) and not isinstance(x, bool)
    return x


def to_float(x: Any) -> float:
    assert isinstance(x, (int, float))
    return x


def from_datetime(x: Any) -> datetime:
    return dateutil.parser.parse(x)


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


def from_dict(f: Callable[[Any], T], x: Any) -> Dict[str, T]:
    assert isinstance(x, dict)
    return { k: f(v) for (k, v) in x.items() }


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


class MediaAttributes:
    """MediaAttributes stores which HTML attributes should be added to the media."""

    autoplay: bool
    """Controlled with attribute character > (adds)"""

    controls: bool
    """Controlled with attribute character = (removes)"""

    loop: bool
    """Controlled with attribute character ~ (adds)"""

    muted: bool
    """Controlled with attribute character > (adds)"""

    playsinline: bool
    """Controlled with attribute character = (adds)"""

    def __init__(self, autoplay: bool, controls: bool, loop: bool, muted: bool, playsinline: bool) -> None:
        self.autoplay = autoplay
        self.controls = controls
        self.loop = loop
        self.muted = muted
        self.playsinline = playsinline

    @staticmethod
    def from_dict(obj: Any) -> 'MediaAttributes':
        assert isinstance(obj, dict)
        autoplay = from_bool(obj.get("autoplay"))
        controls = from_bool(obj.get("controls"))
        loop = from_bool(obj.get("loop"))
        muted = from_bool(obj.get("muted"))
        playsinline = from_bool(obj.get("playsinline"))
        return MediaAttributes(autoplay, controls, loop, muted, playsinline)

    def to_dict(self) -> dict:
        result: dict = {}
        result["autoplay"] = from_bool(self.autoplay)
        result["controls"] = from_bool(self.controls)
        result["loop"] = from_bool(self.loop)
        result["muted"] = from_bool(self.muted)
        result["playsinline"] = from_bool(self.playsinline)
        return result


class ColorPalette:
    """ColorPalette reprensents the object in a Work's metadata.colors."""

    primary: str
    secondary: str
    tertiary: str

    def __init__(self, primary: str, secondary: str, tertiary: str) -> None:
        self.primary = primary
        self.secondary = secondary
        self.tertiary = tertiary

    @staticmethod
    def from_dict(obj: Any) -> 'ColorPalette':
        assert isinstance(obj, dict)
        primary = from_str(obj.get("primary"))
        secondary = from_str(obj.get("secondary"))
        tertiary = from_str(obj.get("tertiary"))
        return ColorPalette(primary, secondary, tertiary)

    def to_dict(self) -> dict:
        result: dict = {}
        result["primary"] = from_str(self.primary)
        result["secondary"] = from_str(self.secondary)
        result["tertiary"] = from_str(self.tertiary)
        return result


class ImageDimensions:
    """ImageDimensions represents metadata about a media as it's extracted from its file."""

    aspect_ratio: float
    """width / height"""

    height: int
    """Height in pixels"""

    width: int
    """Width in pixels"""

    def __init__(self, aspect_ratio: float, height: int, width: int) -> None:
        self.aspect_ratio = aspect_ratio
        self.height = height
        self.width = width

    @staticmethod
    def from_dict(obj: Any) -> 'ImageDimensions':
        assert isinstance(obj, dict)
        aspect_ratio = from_float(obj.get("aspectRatio"))
        height = from_int(obj.get("height"))
        width = from_int(obj.get("width"))
        return ImageDimensions(aspect_ratio, height, width)

    def to_dict(self) -> dict:
        result: dict = {}
        result["aspectRatio"] = to_float(self.aspect_ratio)
        result["height"] = from_int(self.height)
        result["width"] = from_int(self.width)
        return result


class ThumbnailsMap:
    pass

    def __init__(self, ) -> None:
        pass

    @staticmethod
    def from_dict(obj: Any) -> 'ThumbnailsMap':
        assert isinstance(obj, dict)
        return ThumbnailsMap()

    def to_dict(self) -> dict:
        result: dict = {}
        return result


class ContentBlock:
    alt: str
    analyzed: bool
    """whether the media has been analyzed"""

    anchor: str
    attributes: MediaAttributes
    caption: str
    colors: ColorPalette
    content: str
    """html"""

    content_type: str
    dimensions: ImageDimensions
    dist_source: str
    duration: float
    """in seconds"""

    hash: str
    """Hash of the media file, used for caching purposes. Could also serve as an integrity
    check.
    The value is the MD5 hash, base64-encoded.
    """
    has_sound: bool
    id: str
    index: int
    online: bool
    relative_source: str
    size: int
    """in bytes"""

    text: str
    thumbnails: ThumbnailsMap
    thumbnails_built_at: datetime
    title: str
    type: str
    url: str

    def __init__(self, alt: str, analyzed: bool, anchor: str, attributes: MediaAttributes, caption: str, colors: ColorPalette, content: str, content_type: str, dimensions: ImageDimensions, dist_source: str, duration: float, hash: str, has_sound: bool, id: str, index: int, online: bool, relative_source: str, size: int, text: str, thumbnails: ThumbnailsMap, thumbnails_built_at: datetime, title: str, type: str, url: str) -> None:
        self.alt = alt
        self.analyzed = analyzed
        self.anchor = anchor
        self.attributes = attributes
        self.caption = caption
        self.colors = colors
        self.content = content
        self.content_type = content_type
        self.dimensions = dimensions
        self.dist_source = dist_source
        self.duration = duration
        self.hash = hash
        self.has_sound = has_sound
        self.id = id
        self.index = index
        self.online = online
        self.relative_source = relative_source
        self.size = size
        self.text = text
        self.thumbnails = thumbnails
        self.thumbnails_built_at = thumbnails_built_at
        self.title = title
        self.type = type
        self.url = url

    @staticmethod
    def from_dict(obj: Any) -> 'ContentBlock':
        assert isinstance(obj, dict)
        alt = from_str(obj.get("alt"))
        analyzed = from_bool(obj.get("analyzed"))
        anchor = from_str(obj.get("anchor"))
        attributes = MediaAttributes.from_dict(obj.get("attributes"))
        caption = from_str(obj.get("caption"))
        colors = ColorPalette.from_dict(obj.get("colors"))
        content = from_str(obj.get("content"))
        content_type = from_str(obj.get("contentType"))
        dimensions = ImageDimensions.from_dict(obj.get("dimensions"))
        dist_source = from_str(obj.get("distSource"))
        duration = from_float(obj.get("duration"))
        hash = from_str(obj.get("hash"))
        has_sound = from_bool(obj.get("hasSound"))
        id = from_str(obj.get("id"))
        index = from_int(obj.get("index"))
        online = from_bool(obj.get("online"))
        relative_source = from_str(obj.get("relativeSource"))
        size = from_int(obj.get("size"))
        text = from_str(obj.get("text"))
        thumbnails = ThumbnailsMap.from_dict(obj.get("thumbnails"))
        thumbnails_built_at = from_datetime(obj.get("thumbnailsBuiltAt"))
        title = from_str(obj.get("title"))
        type = from_str(obj.get("type"))
        url = from_str(obj.get("url"))
        return ContentBlock(alt, analyzed, anchor, attributes, caption, colors, content, content_type, dimensions, dist_source, duration, hash, has_sound, id, index, online, relative_source, size, text, thumbnails, thumbnails_built_at, title, type, url)

    def to_dict(self) -> dict:
        result: dict = {}
        result["alt"] = from_str(self.alt)
        result["analyzed"] = from_bool(self.analyzed)
        result["anchor"] = from_str(self.anchor)
        result["attributes"] = to_class(MediaAttributes, self.attributes)
        result["caption"] = from_str(self.caption)
        result["colors"] = to_class(ColorPalette, self.colors)
        result["content"] = from_str(self.content)
        result["contentType"] = from_str(self.content_type)
        result["dimensions"] = to_class(ImageDimensions, self.dimensions)
        result["distSource"] = from_str(self.dist_source)
        result["duration"] = to_float(self.duration)
        result["hash"] = from_str(self.hash)
        result["hasSound"] = from_bool(self.has_sound)
        result["id"] = from_str(self.id)
        result["index"] = from_int(self.index)
        result["online"] = from_bool(self.online)
        result["relativeSource"] = from_str(self.relative_source)
        result["size"] = from_int(self.size)
        result["text"] = from_str(self.text)
        result["thumbnails"] = to_class(ThumbnailsMap, self.thumbnails)
        result["thumbnailsBuiltAt"] = self.thumbnails_built_at.isoformat()
        result["title"] = from_str(self.title)
        result["type"] = from_str(self.type)
        result["url"] = from_str(self.url)
        return result


class LocalizedContent:
    abbreviations: Dict[str, str]
    blocks: List[ContentBlock]
    footnotes: Dict[str, str]
    layout: List[List[str]]
    title: str

    def __init__(self, abbreviations: Dict[str, str], blocks: List[ContentBlock], footnotes: Dict[str, str], layout: List[List[str]], title: str) -> None:
        self.abbreviations = abbreviations
        self.blocks = blocks
        self.footnotes = footnotes
        self.layout = layout
        self.title = title

    @staticmethod
    def from_dict(obj: Any) -> 'LocalizedContent':
        assert isinstance(obj, dict)
        abbreviations = from_dict(from_str, obj.get("abbreviations"))
        blocks = from_list(ContentBlock.from_dict, obj.get("blocks"))
        footnotes = from_dict(from_str, obj.get("footnotes"))
        layout = from_list(lambda x: from_list(from_str, x), obj.get("layout"))
        title = from_str(obj.get("title"))
        return LocalizedContent(abbreviations, blocks, footnotes, layout, title)

    def to_dict(self) -> dict:
        result: dict = {}
        result["abbreviations"] = from_dict(from_str, self.abbreviations)
        result["blocks"] = from_list(lambda x: to_class(ContentBlock, x), self.blocks)
        result["footnotes"] = from_dict(from_str, self.footnotes)
        result["layout"] = from_list(lambda x: from_list(from_str, x), self.layout)
        result["title"] = from_str(self.title)
        return result


class DatabaseMeta:
    partial: bool
    """Partial is true if the database was not fully built."""

    def __init__(self, partial: bool) -> None:
        self.partial = partial

    @staticmethod
    def from_dict(obj: Any) -> 'DatabaseMeta':
        assert isinstance(obj, dict)
        partial = from_bool(obj.get("Partial"))
        return DatabaseMeta(partial)

    def to_dict(self) -> dict:
        result: dict = {}
        result["Partial"] = from_bool(self.partial)
        return result


class WorkMetadata:
    additional_metadata: Dict[str, Any]
    aliases: List[str]
    colors: ColorPalette
    database_metadata: DatabaseMeta
    finished: str
    made_with: List[str]
    page_background: str
    private: bool
    started: str
    tags: List[str]
    thumbnail: str
    title_style: str
    wip: bool

    def __init__(self, additional_metadata: Dict[str, Any], aliases: List[str], colors: ColorPalette, database_metadata: DatabaseMeta, finished: str, made_with: List[str], page_background: str, private: bool, started: str, tags: List[str], thumbnail: str, title_style: str, wip: bool) -> None:
        self.additional_metadata = additional_metadata
        self.aliases = aliases
        self.colors = colors
        self.database_metadata = database_metadata
        self.finished = finished
        self.made_with = made_with
        self.page_background = page_background
        self.private = private
        self.started = started
        self.tags = tags
        self.thumbnail = thumbnail
        self.title_style = title_style
        self.wip = wip

    @staticmethod
    def from_dict(obj: Any) -> 'WorkMetadata':
        assert isinstance(obj, dict)
        additional_metadata = from_dict(lambda x: x, obj.get("additionalMetadata"))
        aliases = from_list(from_str, obj.get("aliases"))
        colors = ColorPalette.from_dict(obj.get("colors"))
        database_metadata = DatabaseMeta.from_dict(obj.get("databaseMetadata"))
        finished = from_str(obj.get("finished"))
        made_with = from_list(from_str, obj.get("madeWith"))
        page_background = from_str(obj.get("pageBackground"))
        private = from_bool(obj.get("private"))
        started = from_str(obj.get("started"))
        tags = from_list(from_str, obj.get("tags"))
        thumbnail = from_str(obj.get("thumbnail"))
        title_style = from_str(obj.get("titleStyle"))
        wip = from_bool(obj.get("wip"))
        return WorkMetadata(additional_metadata, aliases, colors, database_metadata, finished, made_with, page_background, private, started, tags, thumbnail, title_style, wip)

    def to_dict(self) -> dict:
        result: dict = {}
        result["additionalMetadata"] = from_dict(lambda x: x, self.additional_metadata)
        result["aliases"] = from_list(from_str, self.aliases)
        result["colors"] = to_class(ColorPalette, self.colors)
        result["databaseMetadata"] = to_class(DatabaseMeta, self.database_metadata)
        result["finished"] = from_str(self.finished)
        result["madeWith"] = from_list(from_str, self.made_with)
        result["pageBackground"] = from_str(self.page_background)
        result["private"] = from_bool(self.private)
        result["started"] = from_str(self.started)
        result["tags"] = from_list(from_str, self.tags)
        result["thumbnail"] = from_str(self.thumbnail)
        result["titleStyle"] = from_str(self.title_style)
        result["wip"] = from_bool(self.wip)
        return result


class Work:
    """Work represents a given work in the database."""

    built_at: datetime
    content: Dict[str, LocalizedContent]
    description_hash: str
    id: str
    metadata: WorkMetadata
    partial: bool

    def __init__(self, built_at: datetime, content: Dict[str, LocalizedContent], description_hash: str, id: str, metadata: WorkMetadata, partial: bool) -> None:
        self.built_at = built_at
        self.content = content
        self.description_hash = description_hash
        self.id = id
        self.metadata = metadata
        self.partial = partial

    @staticmethod
    def from_dict(obj: Any) -> 'Work':
        assert isinstance(obj, dict)
        built_at = from_datetime(obj.get("builtAt"))
        content = from_dict(LocalizedContent.from_dict, obj.get("content"))
        description_hash = from_str(obj.get("descriptionHash"))
        id = from_str(obj.get("id"))
        metadata = WorkMetadata.from_dict(obj.get("metadata"))
        partial = from_bool(obj.get("Partial"))
        return Work(built_at, content, description_hash, id, metadata, partial)

    def to_dict(self) -> dict:
        result: dict = {}
        result["builtAt"] = self.built_at.isoformat()
        result["content"] = from_dict(lambda x: to_class(LocalizedContent, x), self.content)
        result["descriptionHash"] = from_str(self.description_hash)
        result["id"] = from_str(self.id)
        result["metadata"] = to_class(WorkMetadata, self.metadata)
        result["Partial"] = from_bool(self.partial)
        return result


def database_from_dict(s: Any) -> Dict[str, Work]:
    return from_dict(Work.from_dict, s)


def database_to_dict(x: Dict[str, Work]) -> Any:
    return from_dict(lambda x: to_class(Work, x), x)
