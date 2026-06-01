// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'audit_item.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

AuditItem _$AuditItemFromJson(Map<String, dynamic> json) => AuditItem(
  id: (json['id'] as num).toInt(),
  title: json['title'] as String,
  status: json['status'] as String,
  createdAt: DateTime.parse(json['createdAt'] as String),
);

Map<String, dynamic> _$AuditItemToJson(AuditItem instance) => <String, dynamic>{
  'id': instance.id,
  'title': instance.title,
  'status': instance.status,
  'createdAt': instance.createdAt.toIso8601String(),
};
