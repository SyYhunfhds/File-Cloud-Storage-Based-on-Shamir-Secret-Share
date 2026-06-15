import 'package:json_annotation/json_annotation.dart';

part 'server_share_models.g.dart';

// =============================================================================
// share/list — 份额列表
// =============================================================================

/// 份额列表请求
@JsonSerializable(includeIfNull: false)
class ShareListReq {
  @JsonKey(name: 'page')
  final int page;

  @JsonKey(name: 'size')
  final int size;

  const ShareListReq({this.page = 1, this.size = 20});

  factory ShareListReq.fromJson(Map<String, dynamic> json) =>
      _$ShareListReqFromJson(json);

  Map<String, dynamic> toJson() => _$ShareListReqToJson(this);
}

/// 远端可拉取份额（share/list 响应 list 元素）
@JsonSerializable(includeIfNull: false, createToJson: false)
class ServerShareInfo {
  @JsonKey(name: 'share_id')
  final int shareId;

  @JsonKey(name: 'item_id')
  final int itemId;

  final String filename;
  final String owner;

  @JsonKey(name: 'is_expired')
  final bool isExpired;

  @JsonKey(name: 'updated_at')
  final String updatedAt;

  @JsonKey(name: 'expire_at')
  final String expireAt;

  const ServerShareInfo({
    required this.shareId,
    required this.itemId,
    required this.filename,
    required this.owner,
    required this.isExpired,
    required this.updatedAt,
    required this.expireAt,
  });

  factory ServerShareInfo.fromJson(Map<String, dynamic> json) =>
      _$ServerShareInfoFromJson(json);
}

/// 份额列表响应
@JsonSerializable(includeIfNull: false, createToJson: false)
class ShareListRes {
  final int total;
  final List<ServerShareInfo> list;

  const ShareListRes({required this.total, required this.list});

  factory ShareListRes.fromJson(Map<String, dynamic> json) =>
      _$ShareListResFromJson(json);
}

// =============================================================================
// share/pull — 拉取份额（取后即焚）
// =============================================================================

/// 拉取请求
@JsonSerializable(includeIfNull: false)
class SharePullReq {
  @JsonKey(name: 'share_ids')
  final List<int> shareIds;

  const SharePullReq({required this.shareIds});

  factory SharePullReq.fromJson(Map<String, dynamic> json) =>
      _$SharePullReqFromJson(json);

  Map<String, dynamic> toJson() => _$SharePullReqToJson(this);
}

/// 拉取的单个份额（share/pull 响应 shares 元素）
@JsonSerializable(includeIfNull: false, createToJson: false)
class PulledShare {
  @JsonKey(name: 'item_id')
  final int itemId;

  final String filename;

  @JsonKey(name: 'device_share')
  final String deviceShare;

  const PulledShare({
    required this.itemId,
    required this.filename,
    required this.deviceShare,
  });

  factory PulledShare.fromJson(Map<String, dynamic> json) =>
      _$PulledShareFromJson(json);
}

/// 拉取响应
@JsonSerializable(includeIfNull: false, createToJson: false)
class SharePullRes {
  final List<PulledShare> shares;

  const SharePullRes({required this.shares});

  factory SharePullRes.fromJson(Map<String, dynamic> json) =>
      _$SharePullResFromJson(json);
}
