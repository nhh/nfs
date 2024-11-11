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