import 'package:json_annotation/json_annotation.dart';

part 'share.g.dart';

/// 份额模型
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
