from typing import List, Any, TypeVar, Callable, Type, cast


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


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


class ExtractColors:
    default_files: List[str]
    enabled: bool
    extract: List[str]

    def __init__(self, default_files: List[str], enabled: bool, extract: List[str]) -> None:
        self.default_files = default_files
        self.enabled = enabled
        self.extract = extract

    @staticmethod
    def from_dict(obj: Any) -> 'ExtractColors':
        assert isinstance(obj, dict)
        default_files = from_list(from_str, obj.get("default files"))
        enabled = from_bool(obj.get("enabled"))
        extract = from_list(from_str, obj.get("extract"))
        return ExtractColors(default_files, enabled, extract)

    def to_dict(self) -> dict:
        result: dict = {}
        result["default files"] = from_list(from_str, self.default_files)
        result["enabled"] = from_bool(self.enabled)
        result["extract"] = from_list(from_str, self.extract)
        return result


class MakeGifs:
    enabled: bool
    file_name_template: str

    def __init__(self, enabled: bool, file_name_template: str) -> None:
        self.enabled = enabled
        self.file_name_template = file_name_template

    @staticmethod
    def from_dict(obj: Any) -> 'MakeGifs':
        assert isinstance(obj, dict)
        enabled = from_bool(obj.get("enabled"))
        file_name_template = from_str(obj.get("file name template"))
        return MakeGifs(enabled, file_name_template)

    def to_dict(self) -> dict:
        result: dict = {}
        result["enabled"] = from_bool(self.enabled)
        result["file name template"] = from_str(self.file_name_template)
        return result


class MakeThumbnails:
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
    def from_dict(obj: Any) -> 'MakeThumbnails':
        assert isinstance(obj, dict)
        enabled = from_bool(obj.get("enabled"))
        file_name_template = from_str(obj.get("file name template"))
        input_file = from_str(obj.get("input file"))
        sizes = from_list(from_int, obj.get("sizes"))
        return MakeThumbnails(enabled, file_name_template, input_file, sizes)

    def to_dict(self) -> dict:
        result: dict = {}
        result["enabled"] = from_bool(self.enabled)
        result["file name template"] = from_str(self.file_name_template)
        result["input file"] = from_str(self.input_file)
        result["sizes"] = from_list(from_int, self.sizes)
        return result


class Media:
    at: str

    def __init__(self, at: str) -> None:
        self.at = at

    @staticmethod
    def from_dict(obj: Any) -> 'Media':
        assert isinstance(obj, dict)
        at = from_str(obj.get("at"))
        return Media(at)

    def to_dict(self) -> dict:
        result: dict = {}
        result["at"] = from_str(self.at)
        return result


class Tags:
    repository: str

    def __init__(self, repository: str) -> None:
        self.repository = repository

    @staticmethod
    def from_dict(obj: Any) -> 'Tags':
        assert isinstance(obj, dict)
        repository = from_str(obj.get("repository"))
        return Tags(repository)

    def to_dict(self) -> dict:
        result: dict = {}
        result["repository"] = from_str(self.repository)
        return result


class Technologies:
    repository: str

    def __init__(self, repository: str) -> None:
        self.repository = repository

    @staticmethod
    def from_dict(obj: Any) -> 'Technologies':
        assert isinstance(obj, dict)
        repository = from_str(obj.get("repository"))
        return Technologies(repository)

    def to_dict(self) -> dict:
        result: dict = {}
        result["repository"] = from_str(self.repository)
        return result


class Configuration:
    build_metadata_file: str
    extract_colors: ExtractColors
    make_gifs: MakeGifs
    make_thumbnails: MakeThumbnails
    media: Media
    scattered_mode_folder: str
    tags: Tags
    technologies: Technologies

    def __init__(self, build_metadata_file: str, extract_colors: ExtractColors, make_gifs: MakeGifs, make_thumbnails: MakeThumbnails, media: Media, scattered_mode_folder: str, tags: Tags, technologies: Technologies) -> None:
        self.build_metadata_file = build_metadata_file
        self.extract_colors = extract_colors
        self.make_gifs = make_gifs
        self.make_thumbnails = make_thumbnails
        self.media = media
        self.scattered_mode_folder = scattered_mode_folder
        self.tags = tags
        self.technologies = technologies

    @staticmethod
    def from_dict(obj: Any) -> 'Configuration':
        assert isinstance(obj, dict)
        build_metadata_file = from_str(obj.get("build metadata file"))
        extract_colors = ExtractColors.from_dict(obj.get("extract colors"))
        make_gifs = MakeGifs.from_dict(obj.get("make gifs"))
        make_thumbnails = MakeThumbnails.from_dict(obj.get("make thumbnails"))
        media = Media.from_dict(obj.get("media"))
        scattered_mode_folder = from_str(obj.get("scattered mode folder"))
        tags = Tags.from_dict(obj.get("tags"))
        technologies = Technologies.from_dict(obj.get("technologies"))
        return Configuration(build_metadata_file, extract_colors, make_gifs, make_thumbnails, media, scattered_mode_folder, tags, technologies)

    def to_dict(self) -> dict:
        result: dict = {}
        result["build metadata file"] = from_str(self.build_metadata_file)
        result["extract colors"] = to_class(ExtractColors, self.extract_colors)
        result["make gifs"] = to_class(MakeGifs, self.make_gifs)
        result["make thumbnails"] = to_class(MakeThumbnails, self.make_thumbnails)
        result["media"] = to_class(Media, self.media)
        result["scattered mode folder"] = from_str(self.scattered_mode_folder)
        result["tags"] = to_class(Tags, self.tags)
        result["technologies"] = to_class(Technologies, self.technologies)
        return result


def configuration_from_dict(s: Any) -> Configuration:
    return Configuration.from_dict(s)


def configuration_to_dict(x: Configuration) -> Any:
    return to_class(Configuration, x)
