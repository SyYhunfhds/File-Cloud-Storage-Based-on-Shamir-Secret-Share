import 'package:json_annotation/json_annotation.dart';

part 'share.g.dart';

/// 份额模型（用于API响应）
@JsonSerializable()
class Share {
  final int id;
  final int userId;
  final String value;
  final int version;
  final DateTime createdAt;

  Share({
    required this.id,
    required this.userId,
    required this.value,
    required this.version,
    required this.createdAt,
  });

  factory Share.fromJson(Map<String, dynamic> json) => _$ShareFromJson(json);

  Map<String, dynamic> toJson() => _$ShareToJson(this);
}

/// 本地份额存储结构
/// 注意：份额是Base64编码的字符串，不进行任何JSON解析
@JsonSerializable()
class LocalShare {
  final String filename;
  final String shareValue;  // Base64编码的份额字符串
  final DateTime createdAt;

  LocalShare({
    required this.filename,
    required this.shareValue,
    required this.createdAt,
  });

  factory LocalShare.fromJson(Map<String, dynamic> json) => _$LocalShareFromJson(json);

  Map<String, dynamic> toJson() => _$LocalShareToJson(this);
}