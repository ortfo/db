from typing import Optional, Any, TypeVar, Type, cast


T = TypeVar("T")


def from_bool(x: Any) -> bool:
    assert isinstance(x, bool)
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


class Meta:
    partial: Optional[bool]

    def __init__(self, partial: Optional[bool]) -> None:
        self.partial = partial

    @staticmethod
    def from_dict(obj: Any) -> 'Meta':
        assert isinstance(obj, dict)
        partial = from_union([from_bool, from_none], obj.get("Partial"))
        return Meta(partial)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.partial is not None:
            result["Partial"] = from_union([from_bool, from_none], self.partial)
        return result


class Database:
    meta: Optional[Meta]

    def __init__(self, meta: Optional[Meta]) -> None:
        self.meta = meta

    @staticmethod
    def from_dict(obj: Any) -> 'Database':
        assert isinstance(obj, dict)
        meta = from_union([Meta.from_dict, from_none], obj.get("#meta"))
        return Database(meta)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.meta is not None:
            result["#meta"] = from_union([lambda x: to_class(Meta, x), from_none], self.meta)
        return result


def database_from_dict(s: Any) -> Database:
    return Database.from_dict(s)


def database_to_dict(x: Database) -> Any:
    return to_class(Database, x)
