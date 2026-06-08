/// 用户信息（对齐后端 `MeRes`）
///
/// 从 `GET /v1/protected/user/me` 返回的 `data` 字段解析。
class UserInfo {
  final int userId;
  final String username;
  final String email;
  final DateTime registeredAt;
  final String job;
  final int privilege;

  const UserInfo({
    required this.userId,
    required this.username,
    required this.email,
    required this.registeredAt,
    required this.job,
    required this.privilege,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    // userId 可选：兼容 id 或 user_id 字段名
    final int? uid;
    if (json.containsKey('user_id')) {
      uid = (json['user_id'] as num?)?.toInt();
    } else {
      uid = (json['id'] as num?)?.toInt();
    }
    return switch (json) {
      {
        'username': String n,
        'email': String e,
        'registered_at': String r,
        'job': String j,
        'privilege': int p,
      } =>
        UserInfo(
          userId: uid ?? 0,
          username: n,
          email: e,
          registeredAt: DateTime.parse(r),
          job: j,
          privilege: p,
        ),
      _ => throw const FormatException('Invalid UserInfo JSON'),
    };
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'username': username,
      'email': email,
      'registered_at': registeredAt.toIso8601String(),
      'job': job,
      'privilege': privilege,
    };
  }
}
