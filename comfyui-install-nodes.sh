#!/bin/bash -ex
for repo in "${@}"
do
  name="$(basename "${repo}")"

  # clone custom node
  git clone "${repo}" "/comfyui/custom_nodes/${name}"

  reqs="/comfyui/custom_nodes/${name}/requirements.txt"
  if [ -f "${reqs}" ]; then
    # install custom node dependencies if they exist
    pip install -r "${reqs}"
  fi
done
