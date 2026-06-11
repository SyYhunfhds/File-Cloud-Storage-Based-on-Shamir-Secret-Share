// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'share_refresh_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShareRefreshReq _$ShareRefreshReqFromJson(Map<String, dynamic> json) =>
    ShareRefreshReq(
      itemId: (json['item_id'] as num).toInt(),
      recoveryCode: json['recovery_code'] as String?,
      deviceShare: json['share'] as String?,
    );

Map<String, dynamic> _$ShareRefreshReqToJson(ShareRefreshReq instance) =>
    <String, dynamic>{
      'item_id': instance.itemId,
      'recovery_code': ?instance.recoveryCode,
      'share': ?instance.deviceShare,
    };

ShareRefreshProgressMessage _$ShareRefreshProgressMessageFromJson(
  Map<String, dynamic> json,
) => ShareRefreshProgressMessage(
  progress: (json['progress'] as num).toInt(),
  message: json['message'] as String,
  data: json['data'] == null
      ? null
      : ShareRefreshRes.fromJson(json['data'] as Map<String, dynamic>),
);

Map<String, dynamic> _$ShareRefreshProgressMessageToJson(
  ShareRefreshProgressMessage instance,
) => <String, dynamic>{
  'progress': instance.progress,
  'message': instance.message,
  'data': instance.data?.toJson(),
};

ShareRefreshRes _$ShareRefreshResFromJson(Map<String, dynamic> json) =>
    ShareRefreshRes(
      deviceShare: json['device_share'] as String,
      isRecoveryCodeReGenerated: json['is_recovery_code_re_generated'] as bool,
      recoveryCode: json['recovery_code'] as String,
      recoveryShare: json['recovery_share'] as String? ?? '',
    );

Map<String, dynamic> _$ShareRefreshResToJson(ShareRefreshRes instance) =>
    <String, dynamic>{
      'device_share': instance.deviceShare,
      'is_recovery_code_re_generated': instance.isRecoveryCodeReGenerated,
      'recovery_code': instance.recoveryCode,
      'recovery_share': instance.recoveryShare,
    };
