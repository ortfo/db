# Complex layouts

ortfo/db allows you to declare that some [content blocks](/db/your-first-description-file.md#blocks) are arranged in a more interesting way than just one after the other.

## Usage example

<figure>
  <img src="/examples/complex-layouts.png" autoplay muted loop></img>
  <figcaption><a href="https://ewen.works/subfeed-for-spotify">On ewen.works</a></figcaption>
</figure>


## Declaring layouts

Layouts are declared in the metadata frontmatter of your description file:

```md{5-7}
---
started: 2023-04-12
tags: [web, design, ux, ui]
made with: [figma, react, go]
layouts:
	- [p1, m1]
	- [l1, l2, l3]
---

# My awesome project

This is a paragraph of text. It can contain **bold**, *italic*, and [links](https://example.com).

![](./demo.mp4 "Some caption")

[Link to the source code](https://github.com/ortfo/db)

[Documentation](https://ortfo.org/db)

[Website using ortfodb](https://net7.dev/realisation.html)
```

You declare layouts as a grid. To refer to a content block, you put a letter that specifies the type of block (`p` for paragraphs, `m` for media and `l` for standalone links), and a number that refers to the position of that content block in the file.

For example, to refer to the second link, you would write `l2`.

The example above declares the following layout:

<figure style="display:flex;justify-content:center">
	<table>
		<tr>
			<td colspan=3>This is a paragraph of text. It can conta…</td>
			<td colspan=3><code>demo.mp4</code></td>
		</tr>
		<tr>
			<td colspan=2>Link to the source code</td>
			<td colspan=2>Documentation</td>
			<td colspan=2>Website using ortfodb</td>
		</tr>
	</table>
</figure>

You can do just about anything that you would think of.

## In `database.json`

The advantage of declaring layouts like this in ortfo/db is that your frontend has almost nothing left to do do render the content appropriately.

The generated database file will contain a two-dimensional array that represents this layout, in a _homogeneous_ way: every row will contain exactly the same number of columns.

The previous example generates the following in database.json:

```jsonc
{
	"my-work": {
    "content": {
      "en": {
        "blocks": [...],
        "layout": [ // [!code focus]
          // id of p1                                   id of m1 // [!code focus]
          [ "1JsYa91YMM", "1JsYa91YMM", "1JsYa91YMM", "GBpC-nYDgw", "GBpC-nYDgw", "GBpC-nYDgw" ], // [!code focus]
          // id of l1                   id of l2                    id of l3 // [!code focus]
          [ "TYxPfqjbPR", "TYxPfqjbPR", "ycmt3306Po", "ycmt3306Po", "FD-ZGJKusV", "FD-ZGJKusV" ] // [!code focus]
        ], // [!code focus]
        ...
      }
    }
  }
}
```

### An example: rendering with grid-template-areas

Here is an example of how you could render this layout using CSS Grid's [`grid-template-areas`](https://developer.mozilla.org/en-US/docs/Web/CSS/grid-template-areas):

```js
const content = myWorkFromDatabase.content[language];

let areasDeclaration = "";

for (const row of content.layout) {
  let rowDeclaration = "";
  for (const cell of row) {
    // need to prefix with an underscore
    // because some content block IDs can start with a dash,
    // this will be fixed in the future...
    rowDeclaration += `_${cell} `;
  }
  areasDeclaration += `'${rowDeclaration}' `;
}

// If you prefer, functional-style:
areasDeclaration = content.layout
  .map((row) => `'${row.map((id) => `_${id}`).join(" ")}'`)
  .join(" ");

// Assuming container refers to a DOM element
// that contains all of the content blocks
container.style.gridTemplateAreas = areasDeclaration;

for (const { id } of content.blocks) {
  // Assuming each individual content block is
  // a DOM element with an id that is
  // the same as the content block's id
  const block = document.getElementById(id);
  block.style.gridArea = `_${id}`;
}
```
