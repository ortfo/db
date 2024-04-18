from os import PathLike
from ortfodb.configuration import Configuration, configuration_from_dict
from ortfodb.database import AnalyzedWork, database_from_dict
from ortfodb.tags import Tag, tags_from_dict
from ortfodb.technologies import Technology, technologies_from_dict
import json

def load_configuration(path: PathLike) -> Configuration:
    with open(path, 'r') as f:
        data = json.load(f)
    return configuration_from_dict(data)

def load_database(path: PathLike) -> dict[str, AnalyzedWork]:
    with open(path, 'r') as f:
        data = json.load(f)
    return database_from_dict(data)

def load_tags_repository(path: PathLike) -> list[Tag]:
    with open(path, 'r') as f:
        data = json.load(f)
    return tags_from_dict(data)

def load_technologies_repository(path: PathLike) -> list[Technology]:
    with open(path, 'r') as f:
        data = json.load(f)
    return technologies_from_dict(data)
