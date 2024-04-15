# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   tags = Tags.from_json! "[…]"
#   puts tags.first.detect.search.first
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

  class Detect < Dry::Struct
    attribute :files,     Types.Array(Types::String)
    attribute :made_with, Types.Array(Types::String)
    attribute :search,    Types.Array(Types::String)

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        files:     d.fetch("files"),
        made_with: d.fetch("made with"),
        search:    d.fetch("search"),
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

  class Tag < Dry::Struct
    attribute :aliases,       Types.Array(Types::String)
    attribute :description,   Types::String
    attribute :detect,        Detect
    attribute :learn_more_at, Types::String
    attribute :plural,        Types::String
    attribute :singular,      Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        aliases:       d.fetch("aliases"),
        description:   d.fetch("description"),
        detect:        Detect.from_dynamic!(d.fetch("detect")),
        learn_more_at: d.fetch("learn more at"),
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
        "detect"        => detect.to_dynamic,
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
