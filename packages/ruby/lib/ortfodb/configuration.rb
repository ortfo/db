# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   configuration = Configuration.from_json! "{…}"
#   puts configuration.technologies&.repository
#
# If from_json! succeeds, the value returned matches the schema.

require 'json'
require 'dry-types'
require 'dry-struct'

module Ortfodb
  module Types
    include Dry.Types(default: :nominal)

    Integer = Strict::Integer
    Bool    = Strict::Bool
    Hash    = Strict::Hash
    String  = Strict::String
  end

  class ExtractColorsConfiguration < Dry::Struct
    attribute :default_files, Types.Array(Types::String)
    attribute :enabled,       Types::Bool
    attribute :extract,       Types.Array(Types::String)

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        default_files: d.fetch("default files"),
        enabled:       d.fetch("enabled"),
        extract:       d.fetch("extract"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "default files" => default_files,
        "enabled"       => enabled,
        "extract"       => extract,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class MakeGIFSConfiguration < Dry::Struct
    attribute :enabled,            Types::Bool
    attribute :file_name_template, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        enabled:            d.fetch("enabled"),
        file_name_template: d.fetch("file name template"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "enabled"            => enabled,
        "file name template" => file_name_template,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class MakeThumbnailsConfiguration < Dry::Struct
    attribute :enabled,            Types::Bool
    attribute :file_name_template, Types::String
    attribute :input_file,         Types::String
    attribute :sizes,              Types.Array(Types::Integer)

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        enabled:            d.fetch("enabled"),
        file_name_template: d.fetch("file name template"),
        input_file:         d.fetch("input file"),
        sizes:              d.fetch("sizes"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "enabled"            => enabled,
        "file name template" => file_name_template,
        "input file"         => input_file,
        "sizes"              => sizes,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class MediaConfiguration < Dry::Struct

    # Path to the media directory.
    attribute :at, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        at: d.fetch("at"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "at" => at,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class TagsConfiguration < Dry::Struct

    # Path to file describing all tags.
    attribute :repository, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        repository: d.fetch("repository"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "repository" => repository,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class TechnologiesConfiguration < Dry::Struct

    # Path to file describing all technologies.
    attribute :repository, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        repository: d.fetch("repository"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "repository" => repository,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  # Configuration represents what the ortfodb.yaml configuration file describes.
  class Configuration < Dry::Struct

    # Exporter-specific configuration. Maps exporter names to their configuration.
    attribute :exporters, Types::Hash.meta(of: Types::Hash.meta(of: Types::Any)).optional

    attribute :extract_colors,  ExtractColorsConfiguration.optional
    attribute :make_gifs,       MakeGIFSConfiguration.optional
    attribute :make_thumbnails, MakeThumbnailsConfiguration.optional
    attribute :media,           MediaConfiguration.optional

    # Path to the directory containing all projects. Must be absolute.
    attribute :projects_at, Types::String

    attribute :scattered_mode_folder, Types::String
    attribute :tags,                  TagsConfiguration.optional
    attribute :technologies,          TechnologiesConfiguration.optional

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        exporters:             Types::Hash.optional[d["exporters"]]&.map { |k, v| [k, Types::Hash[v].map { |k, v| [k, Types::Any[v]] }.to_h] }&.to_h,
        extract_colors:        d["extract colors"] ? ExtractColorsConfiguration.from_dynamic!(d["extract colors"]) : nil,
        make_gifs:             d["make gifs"] ? MakeGIFSConfiguration.from_dynamic!(d["make gifs"]) : nil,
        make_thumbnails:       d["make thumbnails"] ? MakeThumbnailsConfiguration.from_dynamic!(d["make thumbnails"]) : nil,
        media:                 d["media"] ? MediaConfiguration.from_dynamic!(d["media"]) : nil,
        projects_at:           d.fetch("projects at"),
        scattered_mode_folder: d.fetch("scattered mode folder"),
        tags:                  d["tags"] ? TagsConfiguration.from_dynamic!(d["tags"]) : nil,
        technologies:          d["technologies"] ? TechnologiesConfiguration.from_dynamic!(d["technologies"]) : nil,
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "exporters"             => exporters,
        "extract colors"        => extract_colors&.to_dynamic,
        "make gifs"             => make_gifs&.to_dynamic,
        "make thumbnails"       => make_thumbnails&.to_dynamic,
        "media"                 => media&.to_dynamic,
        "projects at"           => projects_at,
        "scattered mode folder" => scattered_mode_folder,
        "tags"                  => tags&.to_dynamic,
        "technologies"          => technologies&.to_dynamic,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end
end
