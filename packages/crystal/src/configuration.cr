require "json"

module Ortfodb

# Configuration represents what the ortfodb.yaml configuration file describes.
class Configuration
  include JSON::Serializable

  # Exporter-specific configuration. Maps exporter names to their configuration.
  property exporters : Hash(String, Hash(String, JSON::Any?))?

  @[JSON::Field(key: "extract colors")]
  property extract_colors : ExtractColorsConfiguration?

  @[JSON::Field(key: "make gifs")]
  property make_gifs : MakeGiFsConfiguration?

  @[JSON::Field(key: "make thumbnails")]
  property make_thumbnails : MakeThumbnailsConfiguration?

  property media : MediaConfiguration?

  # Path to the directory containing all projects. Must be absolute.
  @[JSON::Field(key: "projects at")]
  property projects_at : String

  @[JSON::Field(key: "scattered mode folder")]
  property scattered_mode_folder : String

  property tags : TagsConfiguration?

  property technologies : TechnologiesConfiguration?
end

class ExtractColorsConfiguration
  include JSON::Serializable

  @[JSON::Field(key: "default files")]
  property default_files : Array(String)

  property enabled : Bool

  property extract : Array(String)
end

class MakeGiFsConfiguration
  include JSON::Serializable

  property enabled : Bool

  @[JSON::Field(key: "file name template")]
  property file_name_template : String
end

class MakeThumbnailsConfiguration
  include JSON::Serializable

  property enabled : Bool

  @[JSON::Field(key: "file name template")]
  property file_name_template : String

  @[JSON::Field(key: "input file")]
  property input_file : String

  property sizes : Array(Int32)
end

class MediaConfiguration
  include JSON::Serializable

  # Path to the media directory.
  property at : String
end

class TagsConfiguration
  include JSON::Serializable

  # Path to file describing all tags.
  property repository : String
end

class TechnologiesConfiguration
  include JSON::Serializable

  # Path to file describing all technologies.
  property repository : String
end
end
