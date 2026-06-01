/// 上传流程阶段
enum UploadPhase { idle, fileSelected, uploading, success, error }

/// 条目上传响应 — 仅含 data 子对象字段（code/message 由 ApiResponse 处理）
class ItemSubmitRes {
  final int itemId;
  final String name;
  final String share;
  final String recoveryCode;

  const ItemSubmitRes({
    required this.itemId,
    required this.name,
    required this.share,
    required this.recoveryCode,
  });

  /// 从 data 子对象 JSON 解析（已由 parseApiResponse 剥离 code/message）
  factory ItemSubmitRes.fromJson(Map<String, dynamic> json) {
    return ItemSubmitRes(
      itemId: json['item_id'] as int? ?? 0,
      name: json['name'] as String? ?? '',
      share: json['share'] as String? ?? '',
      recoveryCode: json['recovery_code'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'item_id': itemId,
        'name': name,
        'share': share,
        'recovery_code': recoveryCode,
      };

  @override
  String toString() => 'ItemSubmitRes(itemId=$itemId, name=$name, share=${share.substring(0, min(8, share.length))}...)';

  static int min(int a, int b) => a < b ? a : b;
}

/// 上传页面状态
class UploadState {
  final UploadPhase phase;
  final String? filePath;
  final String? fileName;
  final double? fileSizeMB;
  final int? fileSizeBytes;
  final int recoveryCodeLength;
  final bool isPublic;
  final double uploadProgress;
  final ItemSubmitRes? result;
  final String? errorMessage;

  const UploadState({
    this.phase = UploadPhase.idle,
    this.filePath,
    this.fileName,
    this.fileSizeMB,
    this.fileSizeBytes,
    this.recoveryCodeLength = 32,
    this.isPublic = false,
    this.uploadProgress = 0.0,
    this.result,
    this.errorMessage,
  });

  UploadState copyWith({
    UploadPhase? phase,
    String? filePath,
    String? fileName,
    double? fileSizeMB,
    int? fileSizeBytes,
    int? recoveryCodeLength,
    bool? isPublic,
    double? uploadProgress,
    ItemSubmitRes? result,
    String? errorMessage,
  }) {
    return UploadState(
      phase: phase ?? this.phase,
      filePath: filePath ?? this.filePath,
      fileName: fileName ?? this.fileName,
      fileSizeMB: fileSizeMB ?? this.fileSizeMB,
      fileSizeBytes: fileSizeBytes ?? this.fileSizeBytes,
      recoveryCodeLength: recoveryCodeLength ?? this.recoveryCodeLength,
      isPublic: isPublic ?? this.isPublic,
      uploadProgress: uploadProgress ?? this.uploadProgress,
      result: result ?? this.result,
      errorMessage: errorMessage,
    );
  }
}
