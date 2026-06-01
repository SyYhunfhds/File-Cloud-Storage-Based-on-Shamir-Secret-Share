import 'package:json_annotation/json_annotation.dart';

part 'about.g.dart';

@JsonSerializable()
class About {
  final String version;
  final String leader;
  final List<String> developers;

  About({
    required this.version,
    required this.leader,
    required this.developers,
  });

  factory About.fromJson(Map<String, dynamic> json) => _$AboutFromJson(json);
  Map<String, dynamic> toJson() => _$AboutToJson(this);
}
