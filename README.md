# comfyui-docker

This repo produces a simple [ComfyUI](https://www.comfy.org/) base docker image

Currently, it:

- Installs pytorch
- Downloads ComfyUI
- Installs ComfyUI's dependencies

Thus, you'll probably want to extend this docker image to further personalize it.

## Usage

This docker image expects you to bind-mount a folder to `/data`. Ensure that the contents of this directory match the structure defined by [ComfyUI](https://github.com/comfyanonymous/ComfyUI/tree/master/models). Finally, supply the arguments you need to the end of your `docker run` command.

This means that if you're running ComfyUI with only CPUs:

```shell
docker run -p 8188:8188 -v "[some-dir]:/data" benfiola/comfyui-docker:[version]-cpu --listen --cpu
```

In the above example:

- `-p 8188:8188` instructs docker to publish the UI/API port for ComfyUI
- `-v [some-dir]:/data` mounts your local folder `[some-dir]` into the container at `/data`
- `--listen --cpu` are the arguments passed directly to ComfyUI - instructing it to listen on `0.0.0.0` and run in cpu-only mode.
