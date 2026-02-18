# General guidelines for running Python jobs on GPU farms

# First check: what job system exists

Different GPU farms use different schedulers. That determines how jobs run.

Run:

```bash
command -v qsub
command -v sbatch
command -v systemctl
```

Interpretation:

| Output                               | Meaning                  | Recommended method          |
|--------------------------------------|--------------------------|-----------------------------|
| `qsub` exists                        | PBS / Torque cluster     | Use PBS job                 |
| `sbatch` exists                      | Slurm cluster            | Use Slurm job               |
| neither but `systemctl --user` works | persistent Linux session | Use systemd user service    |
| none of the above                    | basic server only        | fallback: `tmux` or `nohup` |

The Huma-Num cluster doesn't have qsub, but it does have sbatch, so we will use Slurm for this example. If you have a different setup, adapt accordingly.

# First-time job setup

## Create a persistent workspace for the job

```bash
mkdir -p ~/jobs/example_python_job
cd ~/jobs/example_python_job
mkdir logs artifacts
```

As an example, to test the server, you can use the following simple script that simulates work and saves a result file.

Copy the job's Python script to the job folder:

```bash
vim script.py
```

```python
from pathlib import Path
import os
import torch

print("Hello from Slurm job")
print("torch:", torch.__version__)
print("CUDA_VISIBLE_DEVICES:", os.environ.get("CUDA_VISIBLE_DEVICES"))
print("cuda available:", torch.cuda.is_available())

if torch.cuda.is_available():
    print("device count:", torch.cuda.device_count())
    print("device 0:", torch.cuda.get_device_name(0))

Path("artifacts").mkdir(exist_ok=True)
(Path("artifacts") / "result.txt").write_text(
    f"cuda_available={torch.cuda.is_available()}\n"
)
print("Wrote artifacts/result.txt")
```

You can also add the requirements file if you have dependencies:

```bash
vim requirements.txt
```

```txt
torch
torchvision
ultralytics
```

## Python environment setup

For each job, create a virtual environment

```bash
cd ~/jobs/example_python_job
python3 -m venv .venv
```

Activate it:

```bash
source .venv/bin/activate
```

Upgrade pip:

```bash
pip install -U pip wheel
```

Install dependencies if needed:

```bash
[ -f requirements.txt ] && pip install -r requirements.txt
```

Deactivate afterward:

```bash
deactivate
```

This environment will later be used by the job automatically.

# Running the job

Chose your partition, you can understand the available ones with:

```bash
sinfo -o "%P %a %l %D %t %G %f" | head -n 50
```

For the Hume-Num cluster, Choose either:
* Batch training (disconnect, let it run): `gpu_v100` or `gpu_h100`
* Interactive debugging (quick tests): `gpu_v100_interactive` or `gpu_h100_interactive`

Rule of thumb:
* Start with `gpu_v100` for normal training
* Use `gpu_h100` if faster training is needed and access exists

Create a job script:

```bash
vim job.sbatch
```

Change the `--partition` and `--gres` lines as needed:

```bash
#!/bin/bash
#SBATCH --job-name=hello_gpu
#SBATCH --partition=gpu_v100
#SBATCH --gres=gpu:v100:1
#SBATCH --cpus-per-task=2
#SBATCH --mem=4G
#SBATCH --time=00:05:00
#SBATCH --output=logs/%x_%j.out
#SBATCH --error=logs/%x_%j.err

set -euo pipefail

cd "$SLURM_SUBMIT_DIR"
echo "host=$(hostname)"
nvidia-smi || true

source .venv/bin/activate
python script.py
```

Submit:

```bash
sbatch job.sbatch
```

You can exit the server or start another job, the job will run in the background.

# Monitoring jobs

Queue:

```bash
squeue -u $USER
```

Logs:

```bash
tail -f logs/<jobname>_<jobid>.out
```
And 
```bash
tail -f logs/<jobname>_<jobid>.err
```

For the example above, it would be something like:

```bash
tail -n 200 logs/hello_gpu_<JOBID>.out
tail -n 200 logs/hello_gpu_<JOBID>.err
```

When it finishes, you can check the output file(s).

For the example above:

```bash
cat artifacts/result.txt
```

# Stopping a job

```bash
scancel JOBID
```

Then verify it stopped:

```bash
squeue -u $USER
```