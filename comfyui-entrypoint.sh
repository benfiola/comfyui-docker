#!/bin/bash -ex
current_uid="$(id -u)"
comfy_uid="$(id -u comfyui)"
comfy_gid="$(id -g comfyui)"

if [ "${current_uid}" != "0" ]; then
  if [ "${comfy_uid}" != "${UID}" ] || [ "${comfy_gid}" != "${GID}"]; then
    1>&2 echo "error: cannot update permissions"
    exit 1;
  fi
  python /comfyui/main.py "$@"
  exit $?
fi

if [ "${comfy_uid}" != "${UID}" ]; then
  usermod -u "${UID}" comfyui
fi
if [ "${comfy_gid}" != "${GID}" ]; then
  groupmod -g "${GID}" comfyui
fi

mv /comfyui/models /comfyui/models.bak
mkdir -p /data/models
rm -rf /data/models/configs
cp -R /comfyui/models.bak/configs /data/models/configs
ln -s /data/models /comfyui/models

mkdir -p /data/user
ln -s /data/user /comfyui/user

chown -R comfyui:comfyui /comfyui /data
gosu comfyui:comfyui python /comfyui/main.py "$@"
