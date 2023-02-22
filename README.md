## 🔧 spanner 🔧

`spanner` is a tool for easy managment of job on HPC like refencing by name, logging, ssh to nodes.

Works across Slurm and PBSPro.

### Building

0. Install Go (https://go.dev)

1. Build `spanner` for Linux:
```shell
git clone https://github.com/unkaktus/spanner
cd spanner/cmd/spanner
env GOOS=linux GOARCH=amd64 go build
```
2. `scp` the `spanner` binary to your favorite supercomp and add it to `PATH`.


### Example uses

List jobs:

```shell
$ spanner list
+--------------------+-------+-------+----------------------+-------+------+
|        NAME        | STATE | QUEUE |         TIME         | NODES | MPI  |
+--------------------+-------+-------+----------------------+-------+------+
| Compare_Apples     | R     | small | 21m30s/23h30m0s (1%) |     8 | 6/48 |
| Compare_Oranges    | Q     | small | 0s/23h30m0s (0%)     |     8 | 6/48 |
| Compare_Bananas    | F     | small | 0s/23h30m0s (0%)     |     8 | 6/48 |
+--------------------+-------+-------+----------------------+-------+------+
```

Open full logs in `vim`, even for a finished job:

```shell
$ spanner logs Compare_Apples
```

Or, for `stderr`:

```shell
$ spanner logs Compare_Apples err
```


Similarly, follow the log tail of a job:

```shell
$ spanner logtail Compare_Apples
```


SSH to the node 1 of running job `Compare_Apples`:

```shell
$ spanner ssh Compare_Apples 1
node123$
```

Cancel job `Compare_Apples`:

```shell
$ spanner cancel Compare_Apples
```

Clear the entire job history:

```shell
$ spanner clear history
```