# Need for Sync
### Sync files to pods

![nfs logo](assets/images/need-for-sync.png "") {width=150px}

## Usage:

Place a `nfs.yaml` at every root you want to run `nfs` in.

```yaml
manifest: "v1"

pod:
  cwd: "/app"

files:
  - pattern: "*.{js,css,img,html}"
    hooks: 
      - "yarn run build"
  - pattern: "**/*.php"
    hooks:
      - "yarn run build"
```


# Interesting kubectl / bash foo

```bash

tar cf - website/**/*.vue -P | cat

# outputs:

website/pages/support/faq/index.vue0000664000175000017500000000073414677171254021476 0ustar  niklas-hanftniklas-hanft<script lang="ts">
  import Vue from "vue"
  import { allowedCategories } from "@marbis/common/config/faq"

  export default Vue.extend({
    layout: "empty",
    nuxtI18n: {
      paths: {
        "es-ES": "/soporte/faq",
        "it-IT": "/supporto/faq",
        "nl-NL": "/ondersteuning/faq",
        "pl-PL": "/wsparcie/faq/"
      }
    },
    created(): void {
      this.$router.push(this.localePath(`${this.$route.path}/${allowedCategories[0]}`))
    }
  })
</script>

```


# exec foo

Original
`tar cf - /tmp/foo | kubectl exec -i -n <some-namespace> <some-pod> -- tar xf - -C /tmp/bar`

1. Read pipe
``tar cf - file1 file2 file3``
2. Write pipe
``kubectl exec -i -n fe-nihanft frontend-6cc49dfb67-vnz4k -- tar xf - -C /home/frontend``

One could connect a read / progress pipe in between and measure progress that way
