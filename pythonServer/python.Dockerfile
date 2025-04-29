FROM nvidia/cuda:12.1.1-runtime-ubuntu22.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-pip \
    ffmpeg \
    libsndfile1 \
    gcc \
    g++ \
    ca-certificates \
 && update-ca-certificates \
 && rm -rf /var/lib/apt/lists/*

RUN pip3 install --no-cache-dir --upgrade pip

RUN pip3 install --no-cache-dir \
    torch==2.1.0 \
    torchaudio==2.1.0 \
    --extra-index-url https://download.pytorch.org/whl/cu121

RUN pip3 install --no-cache-dir \
    TTS==0.22.0 \
    grpcio==1.71.0 \
    grpcio-tools==1.71.0

RUN yes | python3 -c "from TTS.utils.manage import ModelManager; manager = ModelManager(); manager.download_model('tts_models/multilingual/multi-dataset/xtts_v2')"

WORKDIR /app
COPY . .

EXPOSE 50051
CMD ["python3", "main.py"]
