from typing import List, Optional, Any, Dict, TypeVar, Callable, Type, cast


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


def from_dict(f: Callable[[Any], T], x: Any) -> Dict[str, T]:
    assert isinstance(x, dict)
    return { k: f(v) for (k, v) in x.items() }


def from_bool(x: Any) -> bool:
    assert isinstance(x, bool)
    return x


def to_class(c: Type[T], x: Any) -> dict:
    assert isinstance(x, c)
    return cast(Any, x).to_dict()


class ExporterCommand:
    log: Optional[List[str]]
    """Log a message. The first argument is the verb, the second is the color, the third is the
    message.
    """
    run: Optional[str]
    """Run a command in a shell"""

    def __init__(self, log: Optional[List[str]], run: Optional[str]) -> None:
        self.log = log
        self.run = run

    @staticmethod
    def from_dict(obj: Any) -> 'ExporterCommand':
        assert isinstance(obj, dict)
        log = from_union([lambda x: from_list(from_str, x), from_none], obj.get("log"))
        run = from_union([from_str, from_none], obj.get("run"))
        return ExporterCommand(log, run)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.log is not None:
            result["log"] = from_union([lambda x: from_list(from_str, x), from_none], self.log)
        if self.run is not None:
            result["run"] = from_union([from_str, from_none], self.run)
        return result


class Exporter:
    after: Optional[List[ExporterCommand]]
    """Commands to run after the build finishes. Go text template that receives .Data and
    .Database, the built database.
    """
    before: Optional[List[ExporterCommand]]
    """Commands to run before the build starts. Go text template that receives .Data"""

    data: Optional[Dict[str, Any]]
    """Initial data"""

    description: str
    """Some documentation about the exporter"""

    name: str
    """The name of the exporter"""

    requires: Optional[List[str]]
    """List of programs that are required to be available in the PATH for the exporter to run."""

    verbose: Optional[bool]
    """If true, will show every command that is run"""

    work: Optional[List[ExporterCommand]]
    """Commands to run during the build, for each work. Go text template that receives .Data and
    .Work, the current work.
    """

    def __init__(self, after: Optional[List[ExporterCommand]], before: Optional[List[ExporterCommand]], data: Optional[Dict[str, Any]], description: str, name: str, requires: Optional[List[str]], verbose: Optional[bool], work: Optional[List[ExporterCommand]]) -> None:
        self.after = after
        self.before = before
        self.data = data
        self.description = description
        self.name = name
        self.requires = requires
        self.verbose = verbose
        self.work = work

    @staticmethod
    def from_dict(obj: Any) -> 'Exporter':
        assert isinstance(obj, dict)
        after = from_union([lambda x: from_list(ExporterCommand.from_dict, x), from_none], obj.get("after"))
        before = from_union([lambda x: from_list(ExporterCommand.from_dict, x), from_none], obj.get("before"))
        data = from_union([lambda x: from_dict(lambda x: x, x), from_none], obj.get("data"))
        description = from_str(obj.get("description"))
        name = from_str(obj.get("name"))
        requires = from_union([lambda x: from_list(from_str, x), from_none], obj.get("requires"))
        verbose = from_union([from_bool, from_none], obj.get("verbose"))
        work = from_union([lambda x: from_list(ExporterCommand.from_dict, x), from_none], obj.get("work"))
        return Exporter(after, before, data, description, name, requires, verbose, work)

    def to_dict(self) -> dict:
        result: dict = {}
        if self.after is not None:
            result["after"] = from_union([lambda x: from_list(lambda x: to_class(ExporterCommand, x), x), from_none], self.after)
        if self.before is not None:
            result["before"] = from_union([lambda x: from_list(lambda x: to_class(ExporterCommand, x), x), from_none], self.before)
        if self.data is not None:
            result["data"] = from_union([lambda x: from_dict(lambda x: x, x), from_none], self.data)
        result["description"] = from_str(self.description)
        result["name"] = from_str(self.name)
        if self.requires is not None:
            result["requires"] = from_union([lambda x: from_list(from_str, x), from_none], self.requires)
        if self.verbose is not None:
            result["verbose"] = from_union([from_bool, from_none], self.verbose)
        if self.work is not None:
            result["work"] = from_union([lambda x: from_list(lambda x: to_class(ExporterCommand, x), x), from_none], self.work)
        return result


def exporter_from_dict(s: Any) -> Exporter:
    return Exporter.from_dict(s)


def exporter_to_dict(x: Exporter) -> Any:
    return to_class(Exporter, x)
