# comfyui-docker

This repo produces a simple [ComfyUI](https://www.comfy.org/) base docker image.

Currently, it:

- Installs pytorch
- Downloads ComfyUI
- Installs ComfyUI's dependencies
- Provides ability to install custom nodes

## Usage

Extend the base image by writing a `Dockerfile`:

```docker
FROM benfiola/comfyui-docker:1.1.6-cuda12

RUN <<EOF
# install custom nodes
comfyui-install-nodes <repo url> ...
>>

ENTRYPOINT ["comfyui-entrypoint", <comfyui-args>...]
```

You can then build and run this docker file.

## Data

This docker image expects you to bind-mount a folder to `/data`. Ensure that the contents of this directory match the structure defined by [ComfyUI](https://github.com/comfyanonymous/ComfyUI/tree/master/models). 
