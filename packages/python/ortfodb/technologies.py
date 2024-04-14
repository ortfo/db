from typing import List, Any, TypeVar, Callable, Type, cast


T = TypeVar("T")


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


class Technology:
    aliases: List[str]
    autodetect: List[str]
    """Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    free-form unquoted string and PATH is a filepath relative to the work folder.
    If CONTENT is found in PATH, we consider that technology to be used in the work.
    """
    by: str
    description: str
    files: List[str]
    """Files contains a list of gitignore-style patterns. If the work contains any of the
    patterns specified, we consider that technology to be used in the work.
    """
    learn_more_at: str
    name: str
    slug: str

    def __init__(self, aliases: List[str], autodetect: List[str], by: str, description: str, files: List[str], learn_more_at: str, name: str, slug: str) -> None:
        self.aliases = aliases
        self.autodetect = autodetect
        self.by = by
        self.description = description
        self.files = files
        self.learn_more_at = learn_more_at
        self.name = name
        self.slug = slug

    @staticmethod
    def from_dict(obj: Any) -> 'Technology':
        assert isinstance(obj, dict)
        aliases = from_list(from_str, obj.get("aliases"))
        autodetect = from_list(from_str, obj.get("autodetect"))
        by = from_str(obj.get("by"))
        description = from_str(obj.get("description"))
        files = from_list(from_str, obj.get("files"))
        learn_more_at = from_str(obj.get("learn more at"))
        name = from_str(obj.get("name"))
        slug = from_str(obj.get("slug"))
        return Technology(aliases, autodetect, by, description, files, learn_more_at, name, slug)

    def to_dict(self) -> dict:
        result: dict = {}
        result["aliases"] = from_list(from_str, self.aliases)
        result["autodetect"] = from_list(from_str, self.autodetect)
        result["by"] = from_str(self.by)
        result["description"] = from_str(self.description)
        result["files"] = from_list(from_str, self.files)
        result["learn more at"] = from_str(self.learn_more_at)
        result["name"] = from_str(self.name)
        result["slug"] = from_str(self.slug)
        return result


def technologies_from_dict(s: Any) -> List[Technology]:
    return from_list(Technology.from_dict, s)


def technologies_to_dict(x: List[Technology]) -> Any:
    return from_list(lambda x: to_class(Technology, x), x)
