require "json"

module Ortfodb

alias Database = Hash(String, AnalyzedWork)

# AnalyzedWork represents a complete work, with analyzed mediae.
class AnalyzedWork
  include JSON::Serializable

  @[JSON::Field(key: "builtAt")]
  property built_at : String

  property content : Hash(String, LocalizedContent)

  @[JSON::Field(key: "descriptionHash")]
  property description_hash : String

  property id : String

  property metadata : WorkMetadata

  @[JSON::Field(key: "Partial")]
  property partial : Bool
end

class LocalizedContent
  include JSON::Serializable

  property blocks : Array(ContentBlock)

  property footnotes : Hash(String, String)

  property layout : Array(Array(String))

  property title : String
end

class ContentBlock
  include JSON::Serializable

  property alt : String

  # whether the media has been analyzed
  property analyzed : Bool

  property anchor : String

  property attributes : MediaAttributes

  property caption : String

  property colors : ColorPalette

  # html
  property content : String

  @[JSON::Field(key: "contentType")]
  property content_type : String

  property dimensions : ImageDimensions

  @[JSON::Field(key: "distSource")]
  property dist_source : String

  # in seconds
  property duration : Float64

  @[JSON::Field(key: "hasSound")]
  property has_sound : Bool

  property id : String

  property index : Int32

  property online : Bool

  @[JSON::Field(key: "relativeSource")]
  property relative_source : String

  # in bytes
  property size : Int32

  property text : String

  property thumbnails : ThumbnailsMap

  @[JSON::Field(key: "thumbnailsBuiltAt")]
  property thumbnails_built_at : String

  property title : String

  @[JSON::Field(key: "type")]
  property content_block_type : String

  property url : String
end

# MediaAttributes stores which HTML attributes should be added to the media.
class MediaAttributes
  include JSON::Serializable

  # Controlled with attribute character > (adds)
  property autoplay : Bool

  # Controlled with attribute character = (removes)
  property controls : Bool

  # Controlled with attribute character ~ (adds)
  property loop : Bool

  # Controlled with attribute character > (adds)
  property muted : Bool

  # Controlled with attribute character = (adds)
  property playsinline : Bool
end

# ColorPalette reprensents the object in a Work's metadata.colors.
class ColorPalette
  include JSON::Serializable

  property primary : String

  property secondary : String

  property tertiary : String
end

# ImageDimensions represents metadata about a media as it's extracted from its file.
class ImageDimensions
  include JSON::Serializable

  # width / height
  @[JSON::Field(key: "aspectRatio")]
  property aspect_ratio : Float64

  # Height in pixels
  property height : Int32

  # Width in pixels
  property width : Int32
end

class ThumbnailsMap
  include JSON::Serializable

end

class WorkMetadata
  include JSON::Serializable

  @[JSON::Field(key: "additionalMetadata")]
  property additional_metadata : Hash(String, JSON::Any?)

  property aliases : Array(String)

  property colors : ColorPalette

  @[JSON::Field(key: "databaseMetadata")]
  property database_metadata : DatabaseMeta

  property finished : String

  @[JSON::Field(key: "madeWith")]
  property made_with : Array(String)

  @[JSON::Field(key: "pageBackground")]
  property page_background : String

  @[JSON::Field(key: "private")]
  property work_metadata_private : Bool

  property started : String

  property tags : Array(String)

  property thumbnail : String

  @[JSON::Field(key: "titleStyle")]
  property title_style : String

  property wip : Bool
end

class DatabaseMeta
  include JSON::Serializable

  # Partial is true if the database was not fully built.
  @[JSON::Field(key: "Partial")]
  property partial : Bool
end
end
