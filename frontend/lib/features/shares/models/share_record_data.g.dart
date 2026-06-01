// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'share_record_data.dart';

// ignore_for_file: type=lint

class ShareRecordDataAdapter extends TypeAdapter<ShareRecordData> {
  @override
  final int typeId = 0;

  @override
  ShareRecordData read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return ShareRecordData(
      itemId: fields[0] as int,
      originalFilename: fields[1] as String,
      serverFilename: fields[2] as String,
      encryptedShare: fields[3] as String,
      encryptedRecoveryCode: fields[4] as String,
      createdAt: fields[5] as DateTime,
    );
  }

  @override
  void write(BinaryWriter writer, ShareRecordData obj) {
    writer
      ..writeByte(6)
      ..writeByte(0)
      ..write(obj.itemId)
      ..writeByte(1)
      ..write(obj.originalFilename)
      ..writeByte(2)
      ..write(obj.serverFilename)
      ..writeByte(3)
      ..write(obj.encryptedShare)
      ..writeByte(4)
      ..write(obj.encryptedRecoveryCode)
      ..writeByte(5)
      ..write(obj.createdAt);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ShareRecordDataAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
