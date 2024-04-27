# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   database = Database.from_json! "{…}"
#   puts database["…"].metadata.tags.first
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
    Double  = Strict::Float | Strict::Integer
  end

  # MediaAttributes stores which HTML attributes should be added to the media.
  class MediaAttributes < Dry::Struct

    # Controlled with attribute character > (adds)
    attribute :autoplay, Types::Bool

    # Controlled with attribute character = (removes)
    attribute :controls, Types::Bool

    # Controlled with attribute character ~ (adds)
    attribute :media_attributes_loop, Types::Bool

    # Controlled with attribute character > (adds)
    attribute :muted, Types::Bool

    # Controlled with attribute character = (adds)
    attribute :playsinline, Types::Bool

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        autoplay:              d.fetch("autoplay"),
        controls:              d.fetch("controls"),
        media_attributes_loop: d.fetch("loop"),
        muted:                 d.fetch("muted"),
        playsinline:           d.fetch("playsinline"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "autoplay"    => autoplay,
        "controls"    => controls,
        "loop"        => media_attributes_loop,
        "muted"       => muted,
        "playsinline" => playsinline,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  # ColorPalette reprensents the object in a Work's metadata.colors.
  class ColorPalette < Dry::Struct
    attribute :primary,   Types::String
    attribute :secondary, Types::String
    attribute :tertiary,  Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        primary:   d.fetch("primary"),
        secondary: d.fetch("secondary"),
        tertiary:  d.fetch("tertiary"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "primary"   => primary,
        "secondary" => secondary,
        "tertiary"  => tertiary,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  # ImageDimensions represents metadata about a media as it's extracted from its file.
  class ImageDimensions < Dry::Struct

    # width / height
    attribute :aspect_ratio, Types::Double

    # Height in pixels
    attribute :height, Types::Integer

    # Width in pixels
    attribute :width, Types::Integer

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        aspect_ratio: d.fetch("aspectRatio"),
        height:       d.fetch("height"),
        width:        d.fetch("width"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "aspectRatio" => aspect_ratio,
        "height"      => height,
        "width"       => width,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class ThumbnailsMap < Dry::Struct

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class ContentBlock < Dry::Struct
    attribute :alt, Types::String

    # whether the media has been analyzed
    attribute :analyzed, Types::Bool

    attribute :anchor,     Types::String
    attribute :attributes, MediaAttributes
    attribute :caption,    Types::String
    attribute :colors,     ColorPalette

    # html
    attribute :content, Types::String

    attribute :content_type, Types::String
    attribute :dimensions,   ImageDimensions
    attribute :dist_source,  Types::String

    # in seconds
    attribute :duration, Types::Double

    # Hash of the media file, used for caching purposes. Could also serve as an integrity
    # check.
    # The value is the MD5 hash, base64-encoded.
    attribute :content_block_hash, Types::String

    attribute :has_sound,       Types::Bool
    attribute :id,              Types::String
    attribute :index,           Types::Integer
    attribute :online,          Types::Bool
    attribute :relative_source, Types::String

    # in bytes
    attribute :size, Types::Integer

    attribute :text,                Types::String
    attribute :thumbnails,          ThumbnailsMap
    attribute :thumbnails_built_at, Types::String
    attribute :title,               Types::String
    attribute :content_block_type,  Types::String
    attribute :url,                 Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        alt:                 d.fetch("alt"),
        analyzed:            d.fetch("analyzed"),
        anchor:              d.fetch("anchor"),
        attributes:          MediaAttributes.from_dynamic!(d.fetch("attributes")),
        caption:             d.fetch("caption"),
        colors:              ColorPalette.from_dynamic!(d.fetch("colors")),
        content:             d.fetch("content"),
        content_type:        d.fetch("contentType"),
        dimensions:          ImageDimensions.from_dynamic!(d.fetch("dimensions")),
        dist_source:         d.fetch("distSource"),
        duration:            d.fetch("duration"),
        content_block_hash:  d.fetch("hash"),
        has_sound:           d.fetch("hasSound"),
        id:                  d.fetch("id"),
        index:               d.fetch("index"),
        online:              d.fetch("online"),
        relative_source:     d.fetch("relativeSource"),
        size:                d.fetch("size"),
        text:                d.fetch("text"),
        thumbnails:          ThumbnailsMap.from_dynamic!(d.fetch("thumbnails")),
        thumbnails_built_at: d.fetch("thumbnailsBuiltAt"),
        title:               d.fetch("title"),
        content_block_type:  d.fetch("type"),
        url:                 d.fetch("url"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "alt"               => alt,
        "analyzed"          => analyzed,
        "anchor"            => anchor,
        "attributes"        => attributes.to_dynamic,
        "caption"           => caption,
        "colors"            => colors.to_dynamic,
        "content"           => content,
        "contentType"       => content_type,
        "dimensions"        => dimensions.to_dynamic,
        "distSource"        => dist_source,
        "duration"          => duration,
        "hash"              => content_block_hash,
        "hasSound"          => has_sound,
        "id"                => id,
        "index"             => index,
        "online"            => online,
        "relativeSource"    => relative_source,
        "size"              => size,
        "text"              => text,
        "thumbnails"        => thumbnails.to_dynamic,
        "thumbnailsBuiltAt" => thumbnails_built_at,
        "title"             => title,
        "type"              => content_block_type,
        "url"               => url,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class LocalizedContent < Dry::Struct
    attribute :abbreviations, Types::Hash.meta(of: Types::String)
    attribute :blocks,        Types.Array(ContentBlock)
    attribute :footnotes,     Types::Hash.meta(of: Types::String)
    attribute :layout,        Types.Array(Types.Array(Types::String))
    attribute :title,         Types::String

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        abbreviations: Types::Hash[d.fetch("abbreviations")].map { |k, v| [k, Types::String[v]] }.to_h,
        blocks:        d.fetch("blocks").map { |x| ContentBlock.from_dynamic!(x) },
        footnotes:     Types::Hash[d.fetch("footnotes")].map { |k, v| [k, Types::String[v]] }.to_h,
        layout:        d.fetch("layout"),
        title:         d.fetch("title"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "abbreviations" => abbreviations,
        "blocks"        => blocks.map { |x| x.to_dynamic },
        "footnotes"     => footnotes,
        "layout"        => layout,
        "title"         => title,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class DatabaseMeta < Dry::Struct

    # Partial is true if the database was not fully built.
    attribute :partial, Types::Bool

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        partial: d.fetch("Partial"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "Partial" => partial,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class WorkMetadata < Dry::Struct
    attribute :additional_metadata, Types::Hash.meta(of: Types::Any)
    attribute :aliases,             Types.Array(Types::String)
    attribute :colors,              ColorPalette
    attribute :database_metadata,   DatabaseMeta
    attribute :finished,            Types::String
    attribute :made_with,           Types.Array(Types::String)
    attribute :page_background,     Types::String
    attribute :private,             Types::Bool
    attribute :started,             Types::String
    attribute :tags,                Types.Array(Types::String)
    attribute :thumbnail,           Types::String
    attribute :title_style,         Types::String
    attribute :wip,                 Types::Bool

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        additional_metadata: Types::Hash[d.fetch("additionalMetadata")].map { |k, v| [k, Types::Any[v]] }.to_h,
        aliases:             d.fetch("aliases"),
        colors:              ColorPalette.from_dynamic!(d.fetch("colors")),
        database_metadata:   DatabaseMeta.from_dynamic!(d.fetch("databaseMetadata")),
        finished:            d.fetch("finished"),
        made_with:           d.fetch("madeWith"),
        page_background:     d.fetch("pageBackground"),
        private:             d.fetch("private"),
        started:             d.fetch("started"),
        tags:                d.fetch("tags"),
        thumbnail:           d.fetch("thumbnail"),
        title_style:         d.fetch("titleStyle"),
        wip:                 d.fetch("wip"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "additionalMetadata" => additional_metadata,
        "aliases"            => aliases,
        "colors"             => colors.to_dynamic,
        "databaseMetadata"   => database_metadata.to_dynamic,
        "finished"           => finished,
        "madeWith"           => made_with,
        "pageBackground"     => page_background,
        "private"            => private,
        "started"            => started,
        "tags"               => tags,
        "thumbnail"          => thumbnail,
        "titleStyle"         => title_style,
        "wip"                => wip,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  # Work represents a given work in the database.
  class Work < Dry::Struct
    attribute :built_at,         Types::String
    attribute :content,          Types::Hash.meta(of: LocalizedContent)
    attribute :description_hash, Types::String
    attribute :id,               Types::String
    attribute :metadata,         WorkMetadata
    attribute :partial,          Types::Bool

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        built_at:         d.fetch("builtAt"),
        content:          Types::Hash[d.fetch("content")].map { |k, v| [k, LocalizedContent.from_dynamic!(v)] }.to_h,
        description_hash: d.fetch("descriptionHash"),
        id:               d.fetch("id"),
        metadata:         WorkMetadata.from_dynamic!(d.fetch("metadata")),
        partial:          d.fetch("Partial"),
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "builtAt"         => built_at,
        "content"         => content.map { |k, v| [k, v.to_dynamic] }.to_h,
        "descriptionHash" => description_hash,
        "id"              => id,
        "metadata"        => metadata.to_dynamic,
        "Partial"         => partial,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  module Ortfodb
    class Database
      def self.from_json!(json)
        Types::Hash[JSON.parse(json, quirks_mode: true)].map { |k, v| [k, Work.from_dynamic!(v)] }.to_h
      end
    end
  end
end
