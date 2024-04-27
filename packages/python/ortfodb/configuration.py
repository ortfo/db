from typing import List, Any, Dict, Optional, TypeVar, Callable, Type, cast


T = TypeVar("T")


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


def from_bool(x: Any) -> bool:
    assert isinstance(x, bool)
    return x


def from_int(x: Any) -> int:
    assert isinstance(x, int) and not isinstance(x, bool)
    return x


def from_dict(f: Callable[[Any], T], x: Any) -> Dict[str, T]:
    assert isinstance(x, dict)
    return { k: f(v) for (k, v) in x.items() }


def from_none(x: Any) -> Any:
    assert x is None
    return x


def from_union(fs, x):
    for f in fs:
        try:
            return f(x)
        except:
            pass
    assert False


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


class ExtractColorsConfiguration:
    default_files: List[str]
    enabled: bool
    extract: List[str]

    def __init__(self, default_files: List[str], enabled: bool, extract: List[str]) -> None:
        self.default_files = default_files
        self.enabled = enabled
        self.extract = extract

    @staticmethod
    def from_dict(obj: Any) -> 'ExtractColorsConfiguration':
        assert isinstance(obj, dict)
        default_files = from_list(from_str, obj.get("default files"))
        enabled = from_bool(obj.get("enabled"))
        extract = from_list(from_str, obj.get("extract"))
        return ExtractColorsConfiguration(default_files, enabled, extract)

    def to_dict(self) -> dict:
        result: dict = {}
        result["default files"] = from_list(from_str, self.default_files)
        result["enabled"] = from_bool(self.enabled)
        result["extract"] = from_list(from_str, self.extract)
        return result


class MakeGIFSConfiguration:
    enabled: bool
    file_name_template: str

    def __init__(self, enabled: bool, file_name_template: str) -> None:
        self.enabled = enabled
        self.file_name_template = file_name_template

    @staticmethod
    def from_dict(obj: Any) -> 'MakeGIFSConfiguration':
        assert isinstance(obj, dict)
        enabled = from_bool(obj.get("enabled"))
        file_name_template = from_str(obj.get("file name template"))
        return MakeGIFSConfiguration(enabled, file_name_template)

    def to_dict(self) -> dict:
        result: dict = {}
        result["enabled"] = from_bool(self.enabled)
        result["file name template"] = from_str(self.file_name_template)
        return result


class MakeThumbnailsConfiguration:
    enabled: bool
    file_name_template: str
    input_file: str
    sizes: List[int]

    def __init__(self, enabled: bool, file_name_template: str, input_file: str, sizes: List[int]) -> None:
        self.enabled = enabled
        self.file_name_template = file_name_template
        self.input_file = input_file
        self.sizes = sizes

    @staticmethod
    def from_dict(obj: Any) -> 'MakeThumbnailsConfiguration':
        assert isinstance(obj, dict)
        enabled = from_bool(obj.get("enabled"))
        file_name_template = from_str(obj.get("file name template"))
        input_file = from_str(obj.get("input file"))
        sizes = from_list(from_int, obj.get("sizes"))
        return MakeThumbnailsConfiguration(enabled, file_name_template, input_file, sizes)

    def to_dict(self) -> dict:
        result: dict = {}
        result["enabled"] = from_bool(self.enabled)
        result["file name template"] = from_str(self.file_name_template)
        result["input file"] = from_str(self.input_file)
        result["sizes"] = from_list(from_int, self.sizes)
        return result


class MediaConfiguration:
    at: str
    """Path to the media directory."""

    def __init__(self, at: str) -> None:
        self.at = at

    @staticmethod
    def from_dict(obj: Any) -> 'MediaConfiguration':
        assert isinstance(obj, dict)
        at = from_str(obj.get("at"))
        return MediaConfiguration(at)

    def to_dict(self) -> dict:
        result: dict = {}
        result["at"] = from_str(self.at)
        return result


class TagsConfiguration:
    repository: str
    """Path to file describing all tags."""

    def __init__(self, repository: str) -> None:
        self.repository = repository

    @staticmethod
    def from_dict(obj: Any) -> 'TagsConfiguration':
        assert isinstance(obj, dict)
        repository = from_str(obj.get("repository"))
        return TagsConfiguration(repository)

    def to_dict(self) -> dict:
        result: dict = {}
        result["repository"] = from_str(self.repository)
        return result


class TechnologiesConfiguration:
    repository: str
    """Path to file describing all technologies."""

    def __init__(self, repository: str) -> None:
        self.repository = repository

    @staticmethod
    def from_dict(obj: Any) -> 'TechnologiesConfiguration':
        assert isinstance(obj, dict)
        repository = from_str(obj.get("repository"))
        return TechnologiesConfiguration(repository)

    def to_dict(self) -> dict:
        result: dict = {}
        result["repository"] = from_str(self.repository)
        return result


class Configuration:
    """Configuration represents what the ortfodb.yaml configuration file describes."""

    exporters: Optional[Dict[str, Dict[str, Any]]]
    """Exporter-specific configuration. Maps exporter names to their configuration."""

    extract_colors: Optional[ExtractColorsConfiguration]
    make_gifs: Optional[MakeGIFSConfiguration]
    make_thumbnails: Optional[MakeThumbnailsConfiguration]
    media: Optional[MediaConfiguration]
    projects_at: str
    """Path to the directory containing all projects. Must be absolute."""

    scattered_mode_folder: str
    tags: Optional[TagsConfiguration]
    technologies: Optional[TechnologiesConfiguration]

    def __init__(self, exporters: Optional[Dict[str, Dict[str, Any]]], extract_colors: Optional[ExtractColorsConfiguration], make_gifs: Optional[MakeGIFSConfiguration], make_thumbnails: Optional[MakeThumbnailsConfiguration], media: Optional[MediaConfiguration], projects_at: str, scattered_mode_folder: str, tags: Optional[TagsConfiguration], technologies: Optional[TechnologiesConfiguration]) -> None:
        self.exporters = exporters
        self.extract_colors = extract_colors
        self.make_gifs = make_gifs
        self.make_thumbnails = make_thumbnails
        self.media = media
        self.projects_at = projects_at
        self.scattered_mode_folder = scattered_mode_folder
        self.tags = tags
        self.technologies = technologies

    @staticmethod
    def from_dict(obj: Any) -> 'Configuration':
        assert isinstance(obj, dict)
        exporters = from_union([lambda x: from_dict(lambda x: from_dict(lambda x: x, x), x), from_none], obj.get("exporters"))
        extract_colors = from_union([ExtractColorsConfiguration.from_dict, from_none], obj.get("extract colors"))
        make_gifs = from_union([MakeGIFSConfiguration.from_dict, from_none], obj.get("make gifs"))
        make_thumbnails = from_union([MakeThumbnailsConfiguration.from_dict, from_none], obj.get("make thumbnails"))
        media = from_union([MediaConfiguration.from_dict, from_none], obj.get("media"))
        projects_at = from_str(obj.get("projects at"))
        scattered_mode_folder = from_str(obj.get("scattered mode folder"))
        tags = from_union([TagsConfiguration.from_dict, from_none], obj.get("tags"))
        technologies = from_union([TechnologiesConfiguration.from_dict, from_none], obj.get("technologies"))
        return Configuration(exporters, extract_colors, make_gifs, make_thumbnails, media, projects_at, scattered_mode_folder, tags, technologies)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.exporters is not None:
            result["exporters"] = from_union([lambda x: from_dict(lambda x: from_dict(lambda x: x, x), x), from_none], self.exporters)
        if self.extract_colors is not None:
            result["extract colors"] = from_union([lambda x: to_class(ExtractColorsConfiguration, x), from_none], self.extract_colors)
        if self.make_gifs is not None:
            result["make gifs"] = from_union([lambda x: to_class(MakeGIFSConfiguration, x), from_none], self.make_gifs)
        if self.make_thumbnails is not None:
            result["make thumbnails"] = from_union([lambda x: to_class(MakeThumbnailsConfiguration, x), from_none], self.make_thumbnails)
        if self.media is not None:
            result["media"] = from_union([lambda x: to_class(MediaConfiguration, x), from_none], self.media)
        result["projects at"] = from_str(self.projects_at)
        result["scattered mode folder"] = from_str(self.scattered_mode_folder)
        if self.tags is not None:
            result["tags"] = from_union([lambda x: to_class(TagsConfiguration, x), from_none], self.tags)
        if self.technologies is not None:
            result["technologies"] = from_union([lambda x: to_class(TechnologiesConfiguration, x), from_none], self.technologies)
        return result


def configuration_from_dict(s: Any) -> Configuration:
    return Configuration.from_dict(s)


def configuration_to_dict(x: Configuration) -> Any:
    return to_class(Configuration, x)
