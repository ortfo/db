# This code may look unusually verbose for Ruby (and it is), but
# it performs some subtle and complex validation of JSON data.
#
# To parse this JSON, add 'dry-struct' and 'dry-types' gems, then do:
#
#   exporter = Exporter.from_json! "{…}"
#   puts exporter.work&.first.log&.first
#
# If from_json! succeeds, the value returned matches the schema.

require 'json'
require 'dry-types'
require 'dry-struct'

module Ortfodb
  module Types
    include Dry.Types(default: :nominal)

    Bool   = Strict::Bool
    Hash   = Strict::Hash
    String = Strict::String
  end

  class ExporterCommand < Dry::Struct

    # Log a message. The first argument is the verb, the second is the color, the third is the
    # message.
    attribute :log, Types.Array(Types::String).optional

    # Run a command in a shell
    attribute :run, Types::String.optional

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        log: d["log"],
        run: d["run"],
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "log" => log,
        "run" => run,
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end

  class Exporter < Dry::Struct

    # Commands to run after the build finishes. Go text template that receives .Data and
    # .Database, the built database.
    attribute :after, Types.Array(ExporterCommand).optional

    # Commands to run before the build starts. Go text template that receives .Data
    attribute :before, Types.Array(ExporterCommand).optional

    # Initial data
    attribute :data, Types::Hash.meta(of: Types::Any).optional

    # Some documentation about the exporter
    attribute :description, Types::String

    # The name of the exporter
    attribute :exporter_name, Types::String

    # List of programs that are required to be available in the PATH for the exporter to run.
    attribute :requires, Types.Array(Types::String).optional

    # If true, will show every command that is run
    attribute :verbose, Types::Bool.optional

    # Commands to run during the build, for each work. Go text template that receives .Data and
    # .Work, the current work.
    attribute :work, Types.Array(ExporterCommand).optional

    def self.from_dynamic!(d)
      d = Types::Hash[d]
      new(
        after:         d["after"]&.map { |x| ExporterCommand.from_dynamic!(x) },
        before:        d["before"]&.map { |x| ExporterCommand.from_dynamic!(x) },
        data:          Types::Hash.optional[d["data"]]&.map { |k, v| [k, Types::Any[v]] }&.to_h,
        description:   d.fetch("description"),
        exporter_name: d.fetch("name"),
        requires:      d["requires"],
        verbose:       d["verbose"],
        work:          d["work"]&.map { |x| ExporterCommand.from_dynamic!(x) },
      )
    end

    def self.from_json!(json)
      from_dynamic!(JSON.parse(json))
    end

    def to_dynamic
      {
        "after"       => after&.map { |x| x.to_dynamic },
        "before"      => before&.map { |x| x.to_dynamic },
        "data"        => data,
        "description" => description,
        "name"        => exporter_name,
        "requires"    => requires,
        "verbose"     => verbose,
        "work"        => work&.map { |x| x.to_dynamic },
      }
    end

    def to_json(options = nil)
      JSON.generate(to_dynamic, options)
    end
  end
end
