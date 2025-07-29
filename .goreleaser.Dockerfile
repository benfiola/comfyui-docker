FROM python:3.12.11-bookworm

# setup arguments
ARG COMFYUI_VERSION=0.3.45
ARG TORCH_VERSION=2.7.1
ARG TORCH_INDEX_URL=https://download.pytorch.org/whl/cu126
ARG TORCHAUDIO_VERSION=2.7.1
ARG TORCHVISION_VERSION=0.22.1

# entrypoint values
ENV DATA_DIR=/data 
ENV GID=1000
ENV UID=1000

# exposes API port
EXPOSE 8188

COPY comfyui-docker /usr/bin/comfyui-docker
RUN comfyui-docker setup
ENTRYPOINT ["comfyui-docker", "entrypoint", "--"]