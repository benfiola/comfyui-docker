FROM python:3.12.11-bookworm

# setup arguments
ARG COMFYUI_VERSION=0.3.50
ARG TORCH_VERSION=2.7.1
ARG TORCH_INDEX_URL=https://download.pytorch.org/whl/cu126
ARG TORCHAUDIO_VERSION=2.7.1
ARG TORCHVISION_VERSION=0.22.1

# entrypoint values
ENV GID=1000
ENV UID=1000

# exposes API port
EXPOSE 8188

COPY comfyui-install.sh /usr/bin/comfyui-install
RUN chmod +x /usr/bin/comfyui-install && comfyui-install

COPY comfyui-install-nodes.sh /usr/bin/comfyui-install-nodes
COPY comfyui-entrypoint.sh /usr/bin/comfyui-entrypoint
RUN chmod +x /usr/bin/comfyui-install-nodes /usr/bin/comfyui-entrypoint

ENTRYPOINT ["comfyui-entrypoint"]