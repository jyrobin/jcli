# jcli

A simple concurrency-friendly CLI package.

Previous `jcli` version employs go-flags under the hood. Each execution rebuilds a command tree
with a provided context to make it more thread-safe. But the design is still unnatural. 

Current implementation derives from a simpler library [clir](https://github.com/leaanthony/clir), by
adding support for context.Context and factoring out some side-effecting code, to allow concurrent
execution on the same command tree.
