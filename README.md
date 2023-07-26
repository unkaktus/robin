## 🔧 spanner 🔧

`spanner` is a tool for easy job managment on HPC like referencing by name, logging,  logging into nodes.

Works across Slurm and PBSPro.

### Easy installation

0. Install MambaForge on your cluster. In case you don't
have internet access there, you can use `mitten` (https://github.com/unkaktus/mitten).

1. Install Go
```shell
mamba install go
```

2. Install `spanner`:
```shell
go install https://github.com/unkaktus/spanner/cmd/spanner@latest
```

3. Add `$HOME/go/bin` into your `$PATH` into your `.bashrc`:
```shell
export PATH="$HOME/go/bin:$PATH"
```

### Manual building

0. Install Go (https://go.dev)

1. Build `spanner` for Linux:
```shell
git clone https://github.com/unkaktus/spanner
cd spanner/cmd/spanner
env GOOS=linux GOARCH=amd64 go build
```
2. `scp` the `spanner` binary to your favorite supercomp and add it to your `$PATH`.


### Example uses

List jobs:

```shell
$ spanner list
╭───────────────────────┬─────────┬───────┬─────────────────────────┬───────┬────────╮
│         NAME          │  STATE  │ QUEUE │          TIME           │ NODES │  MPI   │
├───────────────────────┼─────────┼───────┼─────────────────────────┼───────┼────────┤
│ Compare_Apples        │ R [0]   │ small │ 21m30s/23h30m0s (1%)    │     8 │  6/48  │
│ Compare_Oranges       │ Q [0]   │ small │ 0s/23h30m0s (0%)        │     8 │  6/48  │
│ Compare_Bananas       │ F [9]   │ small │ 0s/23h30m0s (0%)        │     8 │  6/48  │
╰───────────────────────┴─────────┴───────┴─────────────────────────┴───────┴────────╯
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
$ spanner logs -f Compare_Apples
```

Begin a job using `begin.toml` file and configuration file `bucket.dat` for the run:

```shell
$ spanner begin bucket.dat
```

Connect to the shell of the node 1 of running job `Compare_Apples`.
For that, you first need to start your job binary via `spanner tent`:
```shell
[mpirun -n 16] spanner tent ./exe/binary
```
Then, connect to the node:
```shell
$ spanner shell Compare_Apples 1
node123$
```

Cancel job `Compare_Apples`:

```shell
$ spanner cancel Compare_Apples
```