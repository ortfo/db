# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   tags = Tags.from_json! "[…]"
#   puts tags.first.detect&.search&.first
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

  # Various ways to automatically detect that a work is tagged with this tag.
  class Detect < Dry::Struct
    attribute :files,     Types.Array(Types::String).optional
    attribute :made_with, Types.Array(Types::String).optional
    attribute :search,    Types.Array(Types::String).optional

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        files:     d["files"],
        made_with: d["made with"],
        search:    d["search"],
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "files"     => files,
        "made with" => made_with,
        "search"    => search,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  # Tag represents a category that can be assigned to a work.
  class Tag < Dry::Struct

    # Other singular-form names of tags that refer to this tag. The names mentionned here
    # should not be used to define other tags.
    attribute :aliases, Types.Array(Types::String).optional

    attribute :description, Types::String.optional

    # Various ways to automatically detect that a work is tagged with this tag.
    attribute :detect, Detect.optional

    # URL to a website where more information can be found about this tag.
    attribute :learn_more_at, Types::String.optional

    # Plural-form name of the tag. For example, "Books".
    attribute :plural, Types::String

    # Singular-form name of the tag. For example, "Book".
    attribute :singular, Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        aliases:       d["aliases"],
        description:   d["description"],
        detect:        d["detect"] ? Detect.from_dynamic!(d["detect"]) : nil,
        learn_more_at: d["learn more at"],
        plural:        d.fetch("plural"),
        singular:      d.fetch("singular"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "aliases"       => aliases,
        "description"   => description,
        "detect"        => detect&.to_dynamic,
        "learn more at" => learn_more_at,
        "plural"        => plural,
        "singular"      => singular,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  module Ortfodb
    class Tags
      def self.from_json!(json)
        tags = JSON.parse(json, quirks_mode: true).map { |x| Tag.from_dynamic!(x) }
        tags.define_singleton_method(:to_json) do
          JSON.generate(self.map { |x| x.to_dynamic })
        end
        tags
      end
    end
  end
end
