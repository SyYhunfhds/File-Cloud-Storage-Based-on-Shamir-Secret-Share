// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'about.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

About _$AboutFromJson(Map<String, dynamic> json) => About(
  version: json['version'] as String,
  leader: json['leader'] as String,
  developers: (json['developers'] as List<dynamic>)
      .map((e) => e as String)
      .toList(),
);

Map<String, dynamic> _$AboutToJson(About instance) => <String, dynamic>{
  'version': instance.version,
  'leader': instance.leader,
  'developers': instance.developers,
};
