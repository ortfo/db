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


class Detect:
    """Various ways to automatically detect that a work is tagged with this tag."""

    files: Optional[List[str]]
    made_with: Optional[List[str]]
    search: Optional[List[str]]

    def __init__(self, files: Optional[List[str]], made_with: Optional[List[str]], search: Optional[List[str]]) -> None:
        self.files = files
        self.made_with = made_with
        self.search = search

    @staticmethod
    def from_dict(obj: Any) -> 'Detect':
        assert isinstance(obj, dict)
        files = from_union([lambda x: from_list(from_str, x), from_none], obj.get("files"))
        made_with = from_union([lambda x: from_list(from_str, x), from_none], obj.get("made with"))
        search = from_union([lambda x: from_list(from_str, x), from_none], obj.get("search"))
        return Detect(files, made_with, search)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.files is not None:
            result["files"] = from_union([lambda x: from_list(from_str, x), from_none], self.files)
        if self.made_with is not None:
            result["made with"] = from_union([lambda x: from_list(from_str, x), from_none], self.made_with)
        if self.search is not None:
            result["search"] = from_union([lambda x: from_list(from_str, x), from_none], self.search)
        return result


class Tag:
    """Tag represents a category that can be assigned to a work."""

    aliases: Optional[List[str]]
    """Other singular-form names of tags that refer to this tag. The names mentionned here
    should not be used to define other tags.
    """
    description: Optional[str]
    detect: Optional[Detect]
    """Various ways to automatically detect that a work is tagged with this tag."""

    learn_more_at: Optional[str]
    """URL to a website where more information can be found about this tag."""

    plural: str
    """Plural-form name of the tag. For example, "Books"."""

    singular: str
    """Singular-form name of the tag. For example, "Book"."""

    def __init__(self, aliases: Optional[List[str]], description: Optional[str], detect: Optional[Detect], learn_more_at: Optional[str], plural: str, singular: str) -> None:
        self.aliases = aliases
        self.description = description
        self.detect = detect
        self.learn_more_at = learn_more_at
        self.plural = plural
        self.singular = singular

    @staticmethod
    def from_dict(obj: Any) -> 'Tag':
        assert isinstance(obj, dict)
        aliases = from_union([lambda x: from_list(from_str, x), from_none], obj.get("aliases"))
        description = from_union([from_str, from_none], obj.get("description"))
        detect = from_union([Detect.from_dict, from_none], obj.get("detect"))
        learn_more_at = from_union([from_str, from_none], obj.get("learn more at"))
        plural = from_str(obj.get("plural"))
        singular = from_str(obj.get("singular"))
        return Tag(aliases, description, detect, learn_more_at, plural, singular)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.aliases is not None:
            result["aliases"] = from_union([lambda x: from_list(from_str, x), from_none], self.aliases)
        if self.description is not None:
            result["description"] = from_union([from_str, from_none], self.description)
        if self.detect is not None:
            result["detect"] = from_union([lambda x: to_class(Detect, x), from_none], self.detect)
        if self.learn_more_at is not None:
            result["learn more at"] = from_union([from_str, from_none], self.learn_more_at)
        result["plural"] = from_str(self.plural)
        result["singular"] = from_str(self.singular)
        return result


def tags_from_dict(s: Any) -> List[Tag]:
    return from_list(Tag.from_dict, s)


def tags_to_dict(x: List[Tag]) -> Any:
    return from_list(lambda x: to_class(Tag, x), x)
