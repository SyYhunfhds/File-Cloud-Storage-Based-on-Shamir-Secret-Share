/// 用户信息（对齐后端 `MeRes`）
///
/// 从 `GET /v1/protected/user/me` 返回的 `data` 字段解析。
class UserInfo {
  final String username;
  final String email;
  final DateTime registeredAt;
  final String job;
  final int privilege;

  const UserInfo({
    required this.username,
    required this.email,
    required this.registeredAt,
    required this.job,
    required this.privilege,
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return switch (json) {
      {
        'username': String u,
        'email': String e,
        'registered_at': String r,
        'job': String j,
        'privilege': int p,
      } =>
        UserInfo(
          username: u,
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
      'username': username,
      'email': email,
      'registered_at': registeredAt.toIso8601String(),
      'job': job,
      'privilege': privilege,
    };
  }
}
