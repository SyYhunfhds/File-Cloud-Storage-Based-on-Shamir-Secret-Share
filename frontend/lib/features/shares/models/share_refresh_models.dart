/// 份额刷新请求
class ShareRefreshReq {
  final int itemId;
  final String? recoveryCode;
  final String? deviceShare;

  const ShareRefreshReq({
    required this.itemId,
    this.recoveryCode,
    this.deviceShare,
  });

  Map<String, dynamic> toJson() {
    return {
      'item_id': itemId,
      if (recoveryCode != null && recoveryCode!.isNotEmpty)
        'recovery_code': recoveryCode,
      if (deviceShare != null && deviceShare!.isNotEmpty) 'share': deviceShare,
    };
  }
}

/// SSE 进度消息（单条）
class ShareRefreshProgressMessage {
  final int progress; // 0-100
  final String message; // 阶段描述
  final ShareRefreshRes? data; // 仅 progress=100 时非 null

  const ShareRefreshProgressMessage({
    required this.progress,
    required this.message,
    this.data,
  });

  factory ShareRefreshProgressMessage.fromJson(Map<String, dynamic> json) {
    return ShareRefreshProgressMessage(
      progress: (json['progress'] as num?)?.toInt() ?? 0,
      message: json['message'] as String? ?? '',
      data: json['data'] != null
          ? ShareRefreshRes.fromJson(json['data'] as Map<String, dynamic>)
          : null,
    );
  }
}

/// 份额刷新响应（progress=100 时 data 字段）
class ShareRefreshRes {
  final String deviceShare;
  final bool isRecoveryCodeReGenerated;
  final String recoveryCode;
  final String recoveryShare;

  const ShareRefreshRes({
    required this.deviceShare,
    required this.isRecoveryCodeReGenerated,
    required this.recoveryCode,
    required this.recoveryShare,
  });

  factory ShareRefreshRes.fromJson(Map<String, dynamic> json) {
    return ShareRefreshRes(
      deviceShare: json['device_share'] as String? ?? '',
      isRecoveryCodeReGenerated:
          (json['is_recovery_code_re_generated'] as num?)?.toInt() == 1,
      recoveryCode: json['recovery_code'] as String? ?? '',
      recoveryShare: json['recovery_share'] as String? ?? '',
    );
  }
}