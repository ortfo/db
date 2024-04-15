# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   technologies = Technologies.from_json! "[…]"
#   puts technologies.first.files.first
#
# If from_json! succeeds, the value returned matches the schema.

require 'json'
require 'dry-types'
require 'dry-struct'

module Ortfodb
  module Types
    include Dry.Types(default: :nominal)

    Hash   = Strict::Hash
    String = Strict::String
  end

  class Technology < Dry::Struct
    attribute :aliases, Types.Array(Types::String)

    # Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    # free-form unquoted string and PATH is a filepath relative to the work folder.
    # If CONTENT is found in PATH, we consider that technology to be used in the work.
    attribute :autodetect, Types.Array(Types::String)

    attribute :by,          Types::String
    attribute :description, Types::String

    # Files contains a list of gitignore-style patterns. If the work contains any of the
    # patterns specified, we consider that technology to be used in the work.
    attribute :files, Types.Array(Types::String)

    attribute :learn_more_at,   Types::String
    attribute :technology_name, Types::String
    attribute :slug,            Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        aliases:         d.fetch("aliases"),
        autodetect:      d.fetch("autodetect"),
        by:              d.fetch("by"),
        description:     d.fetch("description"),
        files:           d.fetch("files"),
        learn_more_at:   d.fetch("learn more at"),
        technology_name: d.fetch("name"),
        slug:            d.fetch("slug"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "aliases"       => aliases,
        "autodetect"    => autodetect,
        "by"            => by,
        "description"   => description,
        "files"         => files,
        "learn more at" => learn_more_at,
        "name"          => technology_name,
        "slug"          => slug,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  module Ortfodb
    class Technologies
      def self.from_json!(json)
        technologies = JSON.parse(json, quirks_mode: true).map { |x| Technology.from_dynamic!(x) }
        technologies.define_singleton_method(:to_json) do
          JSON.generate(self.map { |x| x.to_dynamic })
        end
        technologies
      end
    end
  end
end
