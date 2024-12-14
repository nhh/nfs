# <img src="./assets/images/need-for-sync.png" style="width:40px;padding-right:10px;margin-bottom:-8px;"> Need for Sync
### Sync files to pods

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

123/a/c/faq/index.vue0000664000175000017500000000073414677171254021476 0ustar  niklas-hanftniklas-hanft<script lang="ts">
  import Vue from "vue"
  import { allowedCategories } from ""/faq"

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

For later... found this script and dont quite understand it. Might be able to use rsync instead of tar
```bash
#! /bin/bash

set -eo pipefail

if [[ -z "$KUBECTL_RSYNC_RSH" ]]; then
    [[ -n "$KUBE_CONTEXT" ]] && echo >&2 "* Found \$KUBE_CONTEXT = $KUBE_CONTEXT"
    [[ -n "$POD_NAMESPACE" ]] && echo >&2 "* Found \$POD_NAMESPACE = $POD_NAMESPACE"
    [[ -n "$POD_NAME" ]] && echo >&2 "* Found \$POD_NAME = $POD_NAME"

    while [[ $# -gt 0 ]]; do
        case "$1" in
        --context)
            KUBE_CONTEXT="$2"
            shift 2
            ;;
        --context=*)
            KUBE_CONTEXT="${1#*=}"
            shift
            ;;
        -c | --container)
            POD_CONTAINER="$2"
            shift 2
            ;;
        --container=*)
            POD_CONTAINER="${1#*=}"
            shift
            ;;
        -n | --namespace)
            POD_NAMESPACE="$2"
            shift 2
            ;;
        --namespace=*)
            POD_NAMESPACE="${1#*=}"
            shift
            ;;
        -h | --help)
            echo "Rsync file and directories from/to Kubernetes pod"
            echo ""
            echo "IMPORTANT:"
            echo "'rsync' must be installed on both the local machine and the target container for this script to work."
            echo ""
            echo "Usage:"
            echo "  $(basename "$0") [options] [--] [rsync-options] SRC DST"
            echo ""
            echo "Options:"
            echo "  -n, --namespace=''         Namespace of the pod"
            echo "      --context=''           The name of the kubeconfig context to use."
            echo "                             Has precedence over KUBE_CONTEXT variable."
            echo "  -c, --container=''         Container name. If omitted, the first container in the pod will be chosen"
            echo "      --help                 Display this help and exit"
            echo ""

            exit
            ;;
        --)
            shift
            break
            ;;
        *)
            break
            ;;
        esac
    done

    export KUBECTL_RSYNC_RSH=true
    export KUBE_CONTEXT POD_NAMESPACE POD_CONTAINER

    set -x
    exec rsync --blocking-io --rsh="$0" "$@"
fi

# Running under --rsh

# If user uses pod@namespace, rsync passes args as `-l pod namespace`
if [[ x"$1" == x"-l" ]]; then
    POD_NAME="$2"
    POD_NAMESPACE="$3"
    shift 3
else
    POD_NAME="$1"
    shift
fi

export KUBE_CONTEXT POD_NAMESPACE POD_CONTAINER POD_NAME
echo >&2 "* Connect to pod $POD_NAME in ${POD_NAMESPACE:-current namespace}"

set -x
exec kubectl exec \
    ${KUBE_CONTEXT:+--context=${KUBE_CONTEXT}} \
    ${POD_NAMESPACE:+--namespace=${POD_NAMESPACE}} \
    ${POD_CONTAINER:+--container=${POD_CONTAINER}} \
    "${POD_NAME}" -i \
    -- \
    "$@"

```
