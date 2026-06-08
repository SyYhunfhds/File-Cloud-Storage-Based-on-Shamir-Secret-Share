/// 审计筛选范围
///
/// 对应后端 scope 参数: 0=已拒绝, 1=已通过, 2=待审查
enum AuditFilterScope {
  all, // [0, 1, 2]
  pending, // [2]
  approved, // [1]
  rejected; // [0]

  /// 转换为后端 API 需要的 scope 数组
  List<int> toScopeList() {
    return switch (this) {
      AuditFilterScope.all => [0, 1, 2],
      AuditFilterScope.pending => [2],
      AuditFilterScope.approved => [1],
      AuditFilterScope.rejected => [0],
    };
  }

  String get label {
    return switch (this) {
      AuditFilterScope.all => '全部',
      AuditFilterScope.pending => '待审查',
      AuditFilterScope.approved => '已通过',
      AuditFilterScope.rejected => '已拒绝',
    };
  }
}

/// 审计条目的聚合状态
enum AggregateStatus {
  pending, // 全部待审查
  approved, // 全部已通过
  rejected, // 全部已拒绝
  mixed, // 混合状态
}

/// 审计条目模型
///
/// 映射后端 `LessDetailedAudit` DTO，后端已按更新时间降序聚合排列。
class LessDetailedAudit {
  final List<int> auditIds;
  final int itemId;
  final String itemName;
  final List<int> rejected;
  final List<int> approved;
  final List<int> memberConfirmed;
  final List<String> applicant;
  final List<String> createdAt;
  final List<String> updatedAt;
  final List<String> joinedAt;
  final List<int> canDownload;
  final List<int> allowDownload;

  const LessDetailedAudit({
    required this.auditIds,
    required this.itemId,
    required this.itemName,
    required this.rejected,
    required this.approved,
    required this.memberConfirmed,
    required this.applicant,
    required this.createdAt,
    required this.updatedAt,
    required this.joinedAt,
    required this.canDownload,
    required this.allowDownload,
  });

  factory LessDetailedAudit.fromJson(Map<String, dynamic> json) {
    return LessDetailedAudit(
      auditIds: (json['audit_id'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
      itemId: (json['item_id'] as num).toInt(),
      itemName: json['item_name'] as String? ?? '',
      rejected: (json['rejected'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
      approved: (json['approved'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
      memberConfirmed: (json['member_confirmed'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
      applicant: (json['applicant'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      createdAt: (json['created_at'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      updatedAt: (json['updated_at'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      joinedAt: (json['joined_at'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      canDownload: (json['can_download'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
      allowDownload: (json['allow_download'] as List<dynamic>?)
              ?.map((e) => (e as num).toInt())
              .toList() ??
          [],
    );
  }

  /// 计算当前聚合状态
  AggregateStatus get status {
    final allRejected =
        rejected.isNotEmpty && rejected.every((r) => r == 1);
    final allApproved =
        approved.isNotEmpty && approved.every((a) => a == 1);

    if (allRejected && !allApproved) return AggregateStatus.rejected;
    if (allApproved && !allRejected) return AggregateStatus.approved;
    if (!allRejected && !allApproved) return AggregateStatus.pending;
    return AggregateStatus.mixed;
  }

  /// 最新的申请人
  String get latestApplicant =>
      applicant.isNotEmpty ? applicant.first : '-';

  /// 最新的创建时间（格式化显示）
  String get latestCreatedAt {
    if (createdAt.isEmpty) return '-';
    final raw = createdAt.first;
    // 截取到分钟级别显示，处理 ISO 8601 格式
    try {
      final dt = DateTime.parse(raw);
      return '${dt.year}-${_pad(dt.month)}-${_pad(dt.day)} '
          '${_pad(dt.hour)}:${_pad(dt.minute)}';
    } catch (_) {
      return raw.length > 16 ? raw.substring(0, 16) : raw;
    }
  }

  /// 审计记录数量
  int get auditCount => createdAt.length;

  /// 待审查的 audit_id 列表（rejected[i]==0 && approved[i]==0）
  List<int> get pendingAuditIds {
    final ids = <int>[];
    for (int i = 0; i < auditIds.length; i++) {
      final rej = i < rejected.length ? rejected[i] : 0;
      final app = i < approved.length ? approved[i] : 0;
      if (rej == 0 && app == 0) {
        ids.add(auditIds[i]);
      }
    }
    return ids;
  }

  static String _pad(int n) => n.toString().padLeft(2, '0');
}

/// 审计列表响应
class AuditListRes {
  final int total;
  final List<LessDetailedAudit> list;

  const AuditListRes({required this.total, required this.list});

  factory AuditListRes.fromJson(Map<String, dynamic> json) {
    return AuditListRes(
      total: (json['total'] as num?)?.toInt() ?? 0,
      list: (json['list'] as List<dynamic>?)
              ?.map(
                  (e) => LessDetailedAudit.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

/// 审计操作响应
class AuditOperationRes {
  final int affected;

  const AuditOperationRes({required this.affected});

  factory AuditOperationRes.fromJson(Map<String, dynamic> json) {
    return AuditOperationRes(
      affected: (json['affected'] as num?)?.toInt() ?? 0,
    );
  }
}
