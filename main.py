# main.py
import grpc
from concurrent import futures
import torch
import os
from TTS.api import TTS
import audio_processor_pb2
import audio_processor_pb2_grpc

os.environ["TF_ENABLE_ONEDNN_OPTS"] = "0"

class AudioProcessorServicer(audio_processor_pb2_grpc.AudioProcessorServicer):
    def __init__(self):
        self.model_name = "tts_models/multilingual/multi-dataset/xtts_v2"
        self.device = "cuda" if torch.cuda.is_available() else "cpu"
        self.tts = TTS(self.model_name).to(self.device)
        self.user_voice_sample = "1.wav"

    def ProcessContent(self, request, context):
        try:
            input_audio_path = "temp_input.wav"
            with open(input_audio_path, "wb") as f:
                f.write(request.audio.data)

            output_audio_path = "generated_voice.ogg"
            self.tts.tts_to_file(
                text=request.text,
                language="ru",
                speaker_wav=input_audio_path,
                file_path=output_audio_path,
            )

            with open(output_audio_path, "rb") as f:
                return audio_processor_pb2.ProcessingResponse(
                    status="OK",
                    result=audio_processor_pb2.AudioResult(
                        processed_audio=f.read()
                    )
                )

        except Exception as e:
            return audio_processor_pb2.ProcessingResponse(
                status=f"ERROR: {str(e)}",
                result=audio_processor_pb2.AudioResult()
            )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    audio_processor_pb2_grpc.add_AudioProcessorServicer_to_server(
        AudioProcessorServicer(), server
    )
    server.add_insecure_port("[::]:50051")
    server.start()
    print("gRPC сервер запущен на порту 50051")
    server.wait_for_termination()

if __name__ == "__main__":
    serve()