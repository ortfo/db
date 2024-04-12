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


class Detect:
    files: List[str]
    made_with: List[str]
    search: List[str]

    def __init__(self, files: List[str], made_with: List[str], search: List[str]) -> None:
        self.files = files
        self.made_with = made_with
        self.search = search

    @staticmethod
    def from_dict(obj: Any) -> 'Detect':
        assert isinstance(obj, dict)
        files = from_list(from_str, obj.get("files"))
        made_with = from_list(from_str, obj.get("made with"))
        search = from_list(from_str, obj.get("search"))
        return Detect(files, made_with, search)

    def to_dict(self) -> dict:
        result: dict = {}
        result["files"] = from_list(from_str, self.files)
        result["made with"] = from_list(from_str, self.made_with)
        result["search"] = from_list(from_str, self.search)
        return result


class Tag:
    aliases: List[str]
    description: str
    detect: Detect
    learn_more_at: str
    plural: str
    singular: str

    def __init__(self, aliases: List[str], description: str, detect: Detect, learn_more_at: str, plural: str, singular: str) -> None:
        self.aliases = aliases
        self.description = description
        self.detect = detect
        self.learn_more_at = learn_more_at
        self.plural = plural
        self.singular = singular

    @staticmethod
    def from_dict(obj: Any) -> 'Tag':
        assert isinstance(obj, dict)
        aliases = from_list(from_str, obj.get("aliases"))
        description = from_str(obj.get("description"))
        detect = Detect.from_dict(obj.get("detect"))
        learn_more_at = from_str(obj.get("learn more at"))
        plural = from_str(obj.get("plural"))
        singular = from_str(obj.get("singular"))
        return Tag(aliases, description, detect, learn_more_at, plural, singular)

    def to_dict(self) -> dict:
        result: dict = {}
        result["aliases"] = from_list(from_str, self.aliases)
        result["description"] = from_str(self.description)
        result["detect"] = to_class(Detect, self.detect)
        result["learn more at"] = from_str(self.learn_more_at)
        result["plural"] = from_str(self.plural)
        result["singular"] = from_str(self.singular)
        return result


def tags_from_dict(s: Any) -> List[Tag]:
    return from_list(Tag.from_dict, s)


def tags_to_dict(x: List[Tag]) -> Any:
    return from_list(lambda x: to_class(Tag, x), x)
