require "json"

module Ortfodb

class Exporter
  include JSON::Serializable

  # Commands to run after the build finishes. Go text template that receives .Data and
  # .Database, the built database.
  property after : Array(ExporterCommand)?

  # Commands to run before the build starts. Go text template that receives .Data
  property before : Array(ExporterCommand)?

  # Initial data
  property data : Hash(String, JSON::Any?)?

  # Some documentation about the exporter
  property description : String

  # The name of the exporter
  property name : String

  # List of programs that are required to be available in the PATH for the exporter to run.
  property requires : Array(String)?

  # If true, will show every command that is run
  property verbose : Bool?

  # Commands to run during the build, for each work. Go text template that receives .Data and
  # .Work, the current work.
  property work : Array(ExporterCommand)?
end

class ExporterCommand
  include JSON::Serializable

  # Log a message. The first argument is the verb, the second is the color, the third is the
  # message.
  property log : Array(String)?

  # Run a command in a shell
  property run : String?
end
end
