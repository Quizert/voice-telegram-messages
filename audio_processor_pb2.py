# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: audio_processor.proto
# Protobuf Python Version: 5.29.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    0,
    '',
    'audio_processor.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x15\x61udio_processor.proto\x12\x13\x61udio_processing.v1\"M\n\x0e\x43ontentRequest\x12\x0c\n\x04text\x18\x01 \x01(\t\x12-\n\x05\x61udio\x18\x02 \x01(\x0b\x32\x1e.audio_processing.v1.AudioFile\"\x19\n\tAudioFile\x12\x0c\n\x04\x64\x61ta\x18\x01 \x01(\x0c\"V\n\x12ProcessingResponse\x12\x0e\n\x06status\x18\x01 \x01(\t\x12\x30\n\x06result\x18\x02 \x01(\x0b\x32 .audio_processing.v1.AudioResult\"&\n\x0b\x41udioResult\x12\x17\n\x0fprocessed_audio\x18\x01 \x01(\x0c\x32p\n\x0e\x41udioProcessor\x12^\n\x0eProcessContent\x12#.audio_processing.v1.ContentRequest\x1a\'.audio_processing.v1.ProcessingResponseB\x15Z\x13kursach/proto;audiob\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'audio_processor_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z\023kursach/proto;audio'
  _globals['_CONTENTREQUEST']._serialized_start=46
  _globals['_CONTENTREQUEST']._serialized_end=123
  _globals['_AUDIOFILE']._serialized_start=125
  _globals['_AUDIOFILE']._serialized_end=150
  _globals['_PROCESSINGRESPONSE']._serialized_start=152
  _globals['_PROCESSINGRESPONSE']._serialized_end=238
  _globals['_AUDIORESULT']._serialized_start=240
  _globals['_AUDIORESULT']._serialized_end=278
  _globals['_AUDIOPROCESSOR']._serialized_start=280
  _globals['_AUDIOPROCESSOR']._serialized_end=392
# @@protoc_insertion_point(module_scope)
