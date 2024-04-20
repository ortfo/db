# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   technologies = Technologies.from_json! "[…]"
#   puts technologies.first.files&.first
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

  # Technology represents a "technology" (in the very broad sense) that was used to create a
  # work.
  class Technology < Dry::Struct

    # Other technology slugs that refer to this technology. The slugs mentionned here should
    # not be used in the definition of other technologies.
    attribute :aliases, Types.Array(Types::String).optional

    # Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
    # free-form unquoted string and PATH is a filepath relative to the work folder.
    # If CONTENT is found in PATH, we consider that technology to be used in the work.
    attribute :autodetect, Types.Array(Types::String).optional

    # Name of the person or organization that created this technology.
    attribute :by, Types::String.optional

    attribute :description, Types::String.optional

    # Files contains a list of gitignore-style patterns. If the work contains any of the
    # patterns specified, we consider that technology to be used in the work.
    attribute :files, Types.Array(Types::String).optional

    # URL to a website where more information can be found about this technology.
    attribute :learn_more_at, Types::String.optional

    attribute :technology_name, Types::String

    # The slug is a unique identifier for this technology, that's suitable for use in a
    # website's URL.
    # For example, the page that shows all works using a technology with slug "a" could be at
    # https://example.org/technologies/a.
    attribute :slug, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        aliases:         d["aliases"],
        autodetect:      d["autodetect"],
        by:              d["by"],
        description:     d["description"],
        files:           d["files"],
        learn_more_at:   d["learn more at"],
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
