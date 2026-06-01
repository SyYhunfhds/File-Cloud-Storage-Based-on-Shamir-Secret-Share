import 'package:json_annotation/json_annotation.dart';

part 'audit_item.g.dart';

/// 审计条目模型
@JsonSerializable()
class AuditItem {
  final int id;
  final String title;
  final String status;
  final DateTime createdAt;

  AuditItem({
    required this.id,
    required this.title,
    required this.status,
    required this.createdAt,
  });

  factory AuditItem.fromJson(Map<String, dynamic> json) => _$AuditItemFromJson(json);

  Map<String, dynamic> toJson() => _$AuditItemToJson(this);

  /// 获取状态显示文本
  String get statusText {
    switch (status) {
      case 'encrypted':
        return '已加密';
      case 'decrypted':
        return '已解密';
      case 'pending':
        return '待处理';
      default:
        return status;
    }
  }

  /// 获取状态颜色
  String get statusColor {
    switch (status) {
      case 'encrypted':
        return '#ef4444'; // red
      case 'decrypted':
        return '#22c55e'; // green
      case 'pending':
        return '#eab308'; // yellow
      default:
        return '#6b7280'; // gray
    }
  }
}
