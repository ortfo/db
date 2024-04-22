import { getGemVersion, getGoVersion, getRustVersion } from "./client-libraries-versions"

export default {
  watch: [
    "ortfodb/packages/ruby/lib/ortfodb/version.rb",
    "ortfodb/packages/rust/Cargo.toml",
	"ortfodb/meta.go"
  ],
  async load(watchedFiles) {
    return {
      versions: {
        gem: await getGemVersion(watchedFiles[0]),
        rust: await getRustVersion(watchedFiles[1]),
		go: await getGoVersion(watchedFiles[2]),
      },
    }
  },
}
