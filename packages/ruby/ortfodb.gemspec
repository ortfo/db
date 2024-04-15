require_relative "lib/ortfodb/version"

Gem::Specification.new do |spec|
  spec.name          = "ortfodb"
  spec.version       = Ortfodb::VERSION
  spec.authors       = ["Ewen Le Bihan"]
  spec.email         = ["ortfo@ewen.works"]
  spec.summary       = "Client library for working with ortfo/db databases"
  spec.description   = "Client library for working with ortfo/db databases. Generated from ortfo/db's JSON schemas (see https://ortfo.org/db/client-libraries)."
  spec.homepage      = "https://ortfo.org/db"
  spec.license       = "MIT"
  if spec.respond_to?(:metadata=)
    spec.metadata = {
      "allowed_push_host" => "https://rubygems.org",
      "bug_tracker_uri"   => "https://github.com/ortfo/db/issues",
      "changelog_uri"     => "https://github.com/ortfo/db/blob/main/CHANGELOG.md",
      "documentation_uri" => "https://ortfo.org/db",
      "homepage_uri"      => spec.homepage,
      "source_code_uri"   => "https://github.com/ortfo/db/tree/main/packages/ruby"
    }
  end
  spec.files         = Dir["lib/**/*"]
  spec.bindir        = "exe"
  spec.require_paths = ["lib"]
  spec.required_ruby_version = Gem::Requirement.new(">= 2.0.0")

  spec.add_dependency 'dry-struct', '~> 1.6'
  spec.add_dependency 'dry-types', '~> 1.7', '>= 1.7.2'
end
