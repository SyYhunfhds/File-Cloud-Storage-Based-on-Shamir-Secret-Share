/// 本地缓存的账号元数据（不含密码）
///
/// 密码独立存储在 [flutter_secure_storage] 中，key = `acct_pwd_$username`。
class CachedAccount {
  final String username;
  final String email;
  final DateTime lastUsedAt;

  const CachedAccount({
    required this.username,
    required this.email,
    required this.lastUsedAt,
  });

  factory CachedAccount.fromJson(Map<String, dynamic> json) => CachedAccount(
        username: json['username'] as String,
        email: json['email'] as String? ?? '',
        lastUsedAt: DateTime.parse(json['last_used_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'username': username,
        'email': email,
        'last_used_at': lastUsedAt.toIso8601String(),
      };
}
