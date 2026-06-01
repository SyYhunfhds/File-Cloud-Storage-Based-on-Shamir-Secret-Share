import 'package:json_annotation/json_annotation.dart';

part 'item.g.dart';

@JsonSerializable()
class ItemSubmitResponse {
  final String name;
  @JsonKey(name: 'share')
  final String authShare;
  @JsonKey(name: 'recovery_code')
  final String recoveryCode;

  ItemSubmitResponse({
    required this.name,
    required this.authShare,
    required this.recoveryCode,
  });

  factory ItemSubmitResponse.fromJson(Map<String, dynamic> json) =>
      _$ItemSubmitResponseFromJson(json);
  Map<String, dynamic> toJson() => _$ItemSubmitResponseToJson(this);
}

@JsonSerializable()
class ItemInfo {
  final String filename;
  final String owner;
  final String uploader;
  @JsonKey(name: 'uploaded_at')
  final DateTime uploadedAt;
  @JsonKey(name: 'changed_at')
  final DateTime changedAt;

  ItemInfo({
    required this.filename,
    required this.owner,
    required this.uploader,
    required this.uploadedAt,
    required this.changedAt,
  });

  factory ItemInfo.fromJson(Map<String, dynamic> json) =>
      _$ItemInfoFromJson(json);
  Map<String, dynamic> toJson() => _$ItemInfoToJson(this);
}

@JsonSerializable()
class ItemListResponse {
  final int count;
  final List<ItemInfo> items;

  ItemListResponse({
    required this.count,
    required this.items,
  });

  factory ItemListResponse.fromJson(Map<String, dynamic> json) =>
      _$ItemListResponseFromJson(json);
  Map<String, dynamic> toJson() => _$ItemListResponseToJson(this);
}