import 'package:json_annotation/json_annotation.dart';

part 'user.g.dart';

/// 用户模型
@JsonSerializable()
class User {
  final int? id;
  final String username;
  final String email;
  @JsonKey(name: 'registered_at')
  final DateTime? registeredAt;
  final String? job;
  final int? privilege;

  User({
    this.id,
    required this.username,
    required this.email,
    this.registeredAt,
    this.job,
    this.privilege,
  });

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
  Map<String, dynamic> toJson() => _$UserToJson(this);

  String get privilegeText {
    switch (privilege) {
      case 1:
        return '普通员工';
      case 2:
        return '主管';
      case 3:
        return '经理';
      case 4:
        return '总监';
      case 5:
        return '管理员';
      default:
        return '未知';
    }
  }
}
