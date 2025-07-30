#!/bin/bash -ex
COMFYUI_URL="https://github.com/comfyanonymous/ComfyUI/archive/refs/tags/v${COMFYUI_VERSION}.tar.gz"

# update apt cache
apt -y update

# install utilities
apt -y install gosu

#/ create comfyui user and group
groupadd -g 1000 comfyui
useradd -u 1000 -g 1000 comfyui

# install torch
pip install --no-cache-dir "torch==${TORCH_VERSION}" "torchaudio==${TORCHAUDIO_VERSION}" "torchvision==${TORCHVISION_VERSION}" --index-url "${TORCH_INDEX_URL}"

# download and extract comfyui
mkdir -p /comfyui /data
curl -o /archive.tar.gz -fsSL "${COMFYUI_URL}"
tar xvzf /archive.tar.gz -C /comfyui --strip-components 1
rm -f /archive.tar.gz

# install comfyui dependencies
pip install --no-cache-dir -r /comfyui/requirements.txt

# ensure directories are owned by non-root user
chown -R comfyui:comfyui /comfyui /data
