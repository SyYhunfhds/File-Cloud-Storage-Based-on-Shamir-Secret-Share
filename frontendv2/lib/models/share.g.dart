// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'share.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Share _$ShareFromJson(Map<String, dynamic> json) => Share(
  id: (json['id'] as num).toInt(),
  userId: (json['userId'] as num).toInt(),
  value: json['value'] as String,
  version: (json['version'] as num).toInt(),
  createdAt: DateTime.parse(json['createdAt'] as String),
);

Map<String, dynamic> _$ShareToJson(Share instance) => <String, dynamic>{
  'id': instance.id,
  'userId': instance.userId,
  'value': instance.value,
  'version': instance.version,
  'createdAt': instance.createdAt.toIso8601String(),
};

LocalShare _$LocalShareFromJson(Map<String, dynamic> json) => LocalShare(
  filename: json['filename'] as String,
  shareValue: json['shareValue'] as String,
  createdAt: DateTime.parse(json['createdAt'] as String),
);

Map<String, dynamic> _$LocalShareToJson(LocalShare instance) =>
    <String, dynamic>{
      'filename': instance.filename,
      'shareValue': instance.shareValue,
      'createdAt': instance.createdAt.toIso8601String(),
    };
