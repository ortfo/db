import { getGoVersion } from "./client-libraries-versions"

export default {
  watch: ["ortfodb/meta.go"],
  async load(watchedFiles) {
    return {
      version: await getGoVersion(watchedFiles[0]),
    }
  },
}
