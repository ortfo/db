export async function getRustVersion(localFilename = "") {
  try {
    const version = await fetch(`https://crates.io/api/v1/crates/ortfodb`)
      .then((res) => res.json())
      .then((data) => data.crate.max_version)
    console.info(`Got Rust version from crates.io API: ${version}`)
    return version
  } catch (error) {
    if (localFilename) {
      const fs = await import("node:fs")
      console.error(`Failed to get Rust version from crates.io API: ${error}`)
      console.info(`Falling back to reading from ortfodb Cargo.toml file`)
      const contents = fs.readFileSync(localFilename, "utf-8")
      return contents.match(/version = "(.+)"/)[1]
    } else {
      throw error
    }
  }
}

export async function getGemVersion(localFilename = "") {
  try {
    const version = await fetch(
      `https://rubygems.org/api/v1/versions/ortfodb/latest.json`
    )
      .then((res) => res.json())
      .then((data) => data.version)

    console.info(`Got RubyGem version from API: ${version}`)
    return version
  } catch (error) {
    if (localFilename) {
      const fs = await import("node:fs")
      console.error(`Failed to get RubyGem version from API: ${error}`)
      console.info(
        `Falling back to reading from ortfodb submodule version.rb file`
      )
      const contents = fs.readFileSync(localFilename, "utf-8")
      return contents.match(/VERSION = "(.+)"/)[1]
    } else {
      throw error
    }
  }
}

export async function getGoVersion(localFilename) {
  const packageName = "github.com/ortfo/db"

  try {
    const response = await fetch(
      `https://proxy.golang.org/${packageName}/@latest`
    )
    const data = await response.json()
    if (data.Version) {
      console.info(`Got Go version from proxy.golang.org: ${data.Version}`)
      return data.Version
    } else {
      throw new Error(`Version not found for package ${packageName}`)
    }
  } catch (error) {
    console.info(`Falling back to reading from local file ${localFilename}`)
    const contents = fs.readFileSync(localFilename, "utf-8")
    const match = contents.match(/const Version = "(.+)"/)
    if (match) {
      return match[1]
    } else {
      throw new Error(`Version not found in ${localFilename}`)
    }
  }
}
