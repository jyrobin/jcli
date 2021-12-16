# jcli

A simple concurrency-friendly CLI package that employs go-flags under the hood.

To summarize, to assist concurrent executions, each `jcli` execution first rebuilds the
`go-flags` command tree with a provided context, whereas each command can decide
whether to keep the context or not to be used in subsequent execution.

Note that although it is desirable to be able to build the command tree only once and
allowing concurrent executions with different contexts, it requires
extensive modifications to `go-flags` and thus not considered.
