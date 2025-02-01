## 🐧 robin

`robin` is a tool for easy job managment on HPC like referencing by name, logging, logging into nodes.

Works on Slurm, PBSPro, and tmux.


### Binary installation

0. In case you don't have internet access there, you can use `mitten` (https://github.com/unkaktus/mitten).

1. Download and install the `robin` binary:
```shell
curl -L -o robin https://github.com/unkaktus/robin/releases/latest/download/robin-linux-amd64
mkdir -p ~/bin
mv robin ~/bin/
chmod +x ~/bin/robin
```

3. Add `$HOME/bin` into your `$PATH` into your `.bashrc`:
```shell
export PATH="$HOME/bin:$PATH"
```

### Manual building

0. Install Go (https://go.dev)

1. Build `robin` for Linux:
```shell
git clone https://github.com/unkaktus/robin
cd robin/cmd/robin
env GOOS=linux GOARCH=amd64 go build
```
2. `scp` the `robin` binary to your favorite supercomp and add it to your `$PATH`.


### Example uses

#### List jobs

```shell
$ robin list
╭───────────────────────┬─────────┬───────┬─────────────────────────┬───────╮
│         NAME          │  STATE  │ QUEUE │          TIME           │ NODES │
├───────────────────────┼─────────┼───────┼─────────────────────────┼───────┤
│ Compare_Apples        │ R [0]   │ small │ [8%] 2h0m41s/24h0m0s    │     8 │
│ Compare_Oranges       │ Q [0]   │ small │ [0%] 0s/20h0m0s         │     2 │
│ Compare_Bananas       │ F [9]   │ small │ [0%] 0s/20h0m0s         │    16 │
╰───────────────────────┴─────────┴───────┴─────────────────────────┴───────╯
```

#### Logs

Open full logs in `$EDITOR` (defauts to `vim`):

```shell
$ robin logs Compare_Apples
```

Follow the log tail of a job:

```shell
$ robin logs -f Compare_Apples
```

#### Shell

To connect to the shell on the job nodes, you first need
to start your job binary via `robin nest`:

```shell
export MPI_SHEPHERD=true # Needed for PBSPro
[mpirun -n 16] robin nest ./exe/binary
```

Then, to connect to the shell of the node 1 (starting from 0) of running job `Compare_Apples`:

```shell
$ robin shell Compare_Apples/1
node123$
```

#### Exec

Using `robin exec`, you can execute a single command on the job nodes:
```shell
$ robin exec Compare_Apples/4 "ps aux"
  PID TTY           TIME CMD
 9993 ttys000    0:00.05 compare_apples
```

#### Stopping jobs

Cancel job `Compare_Apples`:

```shell
$ robin cancel Compare_Apples
```

#### Port forwarding

Forward a port to the node of a job:

```shell
$ robin port-forward -p 11111 -m supercomp compare_apples
```

#### Changing directory to the job root

You might want to add the following function to your `.bashrc`,
that will bring you to the directory from which the job was submitted:

```shell
cdj () {
        cd $(robin list --json | jq -r '.[] | select(.Name=="'$1'").WorkingDirectory')
}
```

Note that it requires `jq`.
