# gpupiped

Check GPU availability and spawn pending process on the issued task queue.

### WARNING

This tool shouldn't be used before bachelor/master/Ph.D thesis submission deadline on sharing machine in your laboratory.
It is because this will occupy GPU resources by your own task.

### Prerequisites

- nvidia-smi

### Examples

1. Run gpiped that will run on target machine with GPU

```
gpiped run
```

2. Define task and publish it into gpiped queue with gpipectl

```
{
  "rootpath": "/path/to/script",
  "command": "<COMMAND>", // It will occupy GPU
  "target_gpu": [0, 1, 2] // Specify target GPU IDs which will be used while executing command. In this example, Usage of GPU 0, 1, 2 will be notified to scheduler.
}
```

```
gpipectl publish --target task.json
```