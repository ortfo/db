require "json"

module Ortfodb

alias Technologies = Array(Technology)

# Technology represents a "technology" (in the very broad sense) that was used to create a
# work.
class Technology
  include JSON::Serializable

  # Other technology slugs that refer to this technology. The slugs mentionned here should
  # not be used in the definition of other technologies.
  property aliases : Array(String)?

  # Autodetect contains an expression of the form 'CONTENT in PATH' where CONTENT is a
  # free-form unquoted string and PATH is a filepath relative to the work folder.
  # If CONTENT is found in PATH, we consider that technology to be used in the work.
  property autodetect : Array(String)?

  # Name of the person or organization that created this technology.
  property by : String?

  property description : String?

  # Files contains a list of gitignore-style patterns. If the work contains any of the
  # patterns specified, we consider that technology to be used in the work.
  property files : Array(String)?

  # URL to a website where more information can be found about this technology.
  @[JSON::Field(key: "learn more at")]
  property learn_more_at : String?

  property name : String

  # The slug is a unique identifier for this technology, that's suitable for use in a
  # website's URL.
  # For example, the page that shows all works using a technology with slug "a" could be at
  # https://example.org/technologies/a.
  property slug : String
end
end
