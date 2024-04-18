require "json"

module Ortfodb

alias Tags = Array(Tag)

# Tag represents a category that can be assigned to a work.
class Tag
  include JSON::Serializable

  # Other singular-form names of tags that refer to this tag. The names mentionned here
  # should not be used to define other tags.
  property aliases : Array(String)?

  property description : String?

  # Various ways to automatically detect that a work is tagged with this tag.
  property detect : Detect?

  # URL to a website where more information can be found about this tag.
  @[JSON::Field(key: "learn more at")]
  property learn_more_at : String?

  # Plural-form name of the tag. For example, "Books".
  property plural : String

  # Singular-form name of the tag. For example, "Book".
  property singular : String
end

# Various ways to automatically detect that a work is tagged with this tag.
class Detect
  include JSON::Serializable

  property files : Array(String)?

  @[JSON::Field(key: "made with")]
  property made_with : Array(String)?

  property search : Array(String)?
end
end
