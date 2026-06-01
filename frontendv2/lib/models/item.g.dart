// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'item.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ItemSubmitResponse _$ItemSubmitResponseFromJson(Map<String, dynamic> json) =>
    ItemSubmitResponse(
      name: json['name'] as String,
      authShare: json['share'] as String,
      recoveryCode: json['recovery_code'] as String,
    );

Map<String, dynamic> _$ItemSubmitResponseToJson(ItemSubmitResponse instance) =>
    <String, dynamic>{
      'name': instance.name,
      'share': instance.authShare,
      'recovery_code': instance.recoveryCode,
    };

ItemInfo _$ItemInfoFromJson(Map<String, dynamic> json) => ItemInfo(
  filename: json['filename'] as String,
  owner: json['owner'] as String,
  uploader: json['uploader'] as String,
  uploadedAt: DateTime.parse(json['uploaded_at'] as String),
  changedAt: DateTime.parse(json['changed_at'] as String),
);

Map<String, dynamic> _$ItemInfoToJson(ItemInfo instance) => <String, dynamic>{
  'filename': instance.filename,
  'owner': instance.owner,
  'uploader': instance.uploader,
  'uploaded_at': instance.uploadedAt.toIso8601String(),
  'changed_at': instance.changedAt.toIso8601String(),
};

ItemListResponse _$ItemListResponseFromJson(Map<String, dynamic> json) =>
    ItemListResponse(
      count: (json['count'] as num).toInt(),
      items: (json['items'] as List<dynamic>)
          .map((e) => ItemInfo.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$ItemListResponseToJson(ItemListResponse instance) =>
    <String, dynamic>{'count': instance.count, 'items': instance.items};
