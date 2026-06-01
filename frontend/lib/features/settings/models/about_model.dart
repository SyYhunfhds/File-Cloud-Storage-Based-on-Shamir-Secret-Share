/// `/v1/about` 接口返回的项目信息模型
class AboutInfo {
  final String version;
  final String leader;
  final List<String> developers;

  const AboutInfo({
    required this.version,
    required this.leader,
    required this.developers,
  });

  factory AboutInfo.fromJson(Map<String, dynamic> json) {
    return AboutInfo(
      version: json['version'] as String? ?? '未知',
      leader: json['leader'] as String? ?? '未知',
      developers: (json['developers'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
    );
  }
}
