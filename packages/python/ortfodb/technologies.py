from typing import List, Optional, Any, TypeVar, Callable, Type, cast


T = TypeVar("T")


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
    assert isinstance(x, list)
    return [f(y) for y in x]


def from_str(x: Any) -> str:
    assert isinstance(x, str)
    return x


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


class Technology:
    """Technology represents a "technology" (in the very broad sense) that was used to create a
    work.
    """
    aliases: Optional[List[str]]
    """Other technology slugs that refer to this technology. The slugs mentionned here should
    not be used in the definition of other technologies.
    """
    autodetect: Optional[List[str]]
    """Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    free-form unquoted string and PATH is a filepath relative to the work folder.
    If CONTENT is found in PATH, we consider that technology to be used in the work.
    """
    by: Optional[str]
    """Name of the person or organization that created this technology."""

    description: Optional[str]
    files: Optional[List[str]]
    """Files contains a list of gitignore-style patterns. If the work contains any of the
    patterns specified, we consider that technology to be used in the work.
    """
    learn_more_at: Optional[str]
    """URL to a website where more information can be found about this technology."""

    name: str
    slug: str
    """The slug is a unique identifier for this technology, that's suitable for use in a
    website's URL.
    For example, the page that shows all works using a technology with slug "a" could be at
    https://example.org/technologies/a.
    """

    def __init__(self, aliases: Optional[List[str]], autodetect: Optional[List[str]], by: Optional[str], description: Optional[str], files: Optional[List[str]], learn_more_at: Optional[str], name: str, slug: str) -> None:
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
        aliases = from_union([lambda x: from_list(from_str, x), from_none], obj.get("aliases"))
        autodetect = from_union([lambda x: from_list(from_str, x), from_none], obj.get("autodetect"))
        by = from_union([from_str, from_none], obj.get("by"))
        description = from_union([from_str, from_none], obj.get("description"))
        files = from_union([lambda x: from_list(from_str, x), from_none], obj.get("files"))
        learn_more_at = from_union([from_str, from_none], obj.get("learn more at"))
        name = from_str(obj.get("name"))
        slug = from_str(obj.get("slug"))
        return Technology(aliases, autodetect, by, description, files, learn_more_at, name, slug)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.aliases is not None:
            result["aliases"] = from_union([lambda x: from_list(from_str, x), from_none], self.aliases)
        if self.autodetect is not None:
            result["autodetect"] = from_union([lambda x: from_list(from_str, x), from_none], self.autodetect)
        if self.by is not None:
            result["by"] = from_union([from_str, from_none], self.by)
        if self.description is not None:
            result["description"] = from_union([from_str, from_none], self.description)
        if self.files is not None:
            result["files"] = from_union([lambda x: from_list(from_str, x), from_none], self.files)
        if self.learn_more_at is not None:
            result["learn more at"] = from_union([from_str, from_none], self.learn_more_at)
        result["name"] = from_str(self.name)
        result["slug"] = from_str(self.slug)
        return result


def technologies_from_dict(s: Any) -> List[Technology]:
    return from_list(Technology.from_dict, s)


def technologies_to_dict(x: List[Technology]) -> Any:
    return from_list(lambda x: to_class(Technology, x), x)
