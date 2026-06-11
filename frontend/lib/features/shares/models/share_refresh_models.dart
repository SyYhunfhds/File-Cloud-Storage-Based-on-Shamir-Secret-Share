import 'package:json_annotation/json_annotation.dart';

part 'share_refresh_models.g.dart';

/// 份额刷新请求
@JsonSerializable(includeIfNull: false)
class ShareRefreshReq {
  @JsonKey(name: 'item_id')
  final int itemId;

  @JsonKey(name: 'recovery_code')
  final String? recoveryCode;

  @JsonKey(name: 'share')
  final String? deviceShare;

  const ShareRefreshReq({
    required this.itemId,
    this.recoveryCode,
    this.deviceShare,
  });

  factory ShareRefreshReq.fromJson(Map<String, dynamic> json) =>
      _$ShareRefreshReqFromJson(json);

  Map<String, dynamic> toJson() {
    final map = _$ShareRefreshReqToJson(this);
    // 排除空字符串（includeIfNull 只处理 null）
    map.removeWhere((key, value) => value is String && value.isEmpty);
    return map;
  }
}

/// SSE 进度消息（单条）
@JsonSerializable(explicitToJson: true)
class ShareRefreshProgressMessage {
  final int progress; // 0-100
  final String message; // 阶段描述
  final ShareRefreshRes? data; // 仅 progress=100 时非 null

  const ShareRefreshProgressMessage({
    required this.progress,
    required this.message,
    this.data,
  });

  factory ShareRefreshProgressMessage.fromJson(Map<String, dynamic> json) =>
      _$ShareRefreshProgressMessageFromJson(json);
}

/// 份额刷新响应（progress=100 时 data 字段）
@JsonSerializable()
class ShareRefreshRes {
  @JsonKey(name: 'device_share')
  final String deviceShare;

  @JsonKey(name: 'is_recovery_code_re_generated')
  final bool isRecoveryCodeReGenerated;

  @JsonKey(name: 'recovery_code')
  final String recoveryCode;

  @JsonKey(name: 'recovery_share', defaultValue: '')
  final String recoveryShare;

  const ShareRefreshRes({
    required this.deviceShare,
    required this.isRecoveryCodeReGenerated,
    required this.recoveryCode,
    this.recoveryShare = '',
  });

  factory ShareRefreshRes.fromJson(Map<String, dynamic> json) =>
      _$ShareRefreshResFromJson(json);

  Map<String, dynamic> toJson() => _$ShareRefreshResToJson(this);
}
