import { load } from "cheerio"

const hiddenBackends = ["The local filesystem", "FTP", "HTTP"]

export default {
  async load() {
    // Parse the table on rclone.org/overview to get the list of supported cloud service providers
    const response = await fetch("https://rclone.org/overview/")
    const html = await response.text()
    const $ = load(html)
    const backendsTable = $("#features ~ table")[0]
	// get first td of each tr in tbody
	const backends = $(backendsTable)
	  .find("tbody tr")
	  .map((_, row) => $(row).find("td").first().text())
	  .get()
	  .filter(t => !hiddenBackends.includes(t))
	return { backends }
  },
}
