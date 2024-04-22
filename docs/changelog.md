---
editLink: false
---


<script setup>
  import { data } from './changelog.data.js'
  console.log(data)
</script>

<main v-html="data[0].html"></main>
