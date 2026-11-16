import hashlib
import json
import os
import subprocess
import sys
from pathlib import Path

MODEL = "intfloat/multilingual-e5-small"
REVISION = "614241f622f53c4eeff9890bdc4f31cfecc418b3"
MODEL_FILE = "onnx/model.onnx"
MODEL_SHA256 = "ca456c06b3a9505ddfd9131408916dd79290368331e7d76bb621f1cba6bc8665"
FILES = [
    MODEL_FILE,
    "config.json",
    "tokenizer.json",
    "tokenizer_config.json",
    "special_tokens_map.json",
]


def available(path):
    path = Path(path)
    try:
        marker = json.loads((path / "ready.json").read_text())
        return marker == {"revision": REVISION, "sha256": MODEL_SHA256} and all(
            (path / name).is_file() for name in FILES
        )
    except (OSError, ValueError):
        return False


def ensure_model(path, timeout=120):
    if available(path):
        return None
    try:
        result = subprocess.run(
            [sys.executable, str(Path(__file__).resolve()), "--download", str(path)],
            timeout=timeout,
            capture_output=True,
            text=True,
            env=dict(
                os.environ,
                HF_HUB_DISABLE_TELEMETRY="1",
                HF_HUB_DISABLE_XET="1",
                HF_HUB_DOWNLOAD_TIMEOUT="20",
                HF_HUB_ETAG_TIMEOUT="10",
            ),
        )
        if result.returncode:
            return "model_download_failed"
    except subprocess.TimeoutExpired:
        return "model_download_timeout"
    return None if available(path) else "model_files_missing"


def encoder(path):
    from fastembed import TextEmbedding
    from fastembed.common.model_description import ModelSource, PoolingType

    if not available(path):
        raise ValueError("E5 model is not prepared")
    TextEmbedding.add_custom_model(
        model=MODEL,
        pooling=PoolingType.MEAN,
        normalization=True,
        sources=ModelSource(hf=MODEL),
        dim=384,
        model_file=MODEL_FILE,
    )
    return TextEmbedding(
        model_name=MODEL,
        specific_model_path=str(Path(path).resolve()),
        local_files_only=True,
        threads=4,
        providers=["CPUExecutionProvider"],
    )


def download(path):
    from huggingface_hub import snapshot_download

    path = Path(path)
    path.mkdir(parents=True, exist_ok=True)
    snapshot_download(MODEL, revision=REVISION, local_dir=path, allow_patterns=FILES)
    digest = hashlib.sha256()
    with (path / MODEL_FILE).open("rb") as source:
        for chunk in iter(lambda: source.read(8 << 20), b""):
            digest.update(chunk)
    if digest.hexdigest() != MODEL_SHA256:
        raise ValueError("E5 checksum mismatch")
    (path / "ready.json").write_text(
        json.dumps({"revision": REVISION, "sha256": MODEL_SHA256})
    )


if __name__ == "__main__":
    if len(sys.argv) != 3 or sys.argv[1] != "--download":
        raise SystemExit("usage: semantic.py --download DIRECTORY")
    download(sys.argv[2])
