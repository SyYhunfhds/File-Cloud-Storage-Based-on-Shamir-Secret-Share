// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'server_share_models.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

ShareListReq _$ShareListReqFromJson(Map<String, dynamic> json) => ShareListReq(
  page: (json['page'] as num?)?.toInt() ?? 1,
  size: (json['size'] as num?)?.toInt() ?? 20,
);

Map<String, dynamic> _$ShareListReqToJson(ShareListReq instance) =>
    <String, dynamic>{'page': instance.page, 'size': instance.size};

ServerShareInfo _$ServerShareInfoFromJson(Map<String, dynamic> json) =>
    ServerShareInfo(
      shareId: (json['share_id'] as num).toInt(),
      itemId: (json['item_id'] as num).toInt(),
      filename: json['filename'] as String,
      owner: json['owner'] as String,
      isExpired: json['is_expired'] as bool,
      updatedAt: json['updated_at'] as String,
      expireAt: json['expire_at'] as String,
    );

ShareListRes _$ShareListResFromJson(Map<String, dynamic> json) => ShareListRes(
  total: (json['total'] as num).toInt(),
  list: (json['list'] as List<dynamic>)
      .map((e) => ServerShareInfo.fromJson(e as Map<String, dynamic>))
      .toList(),
);

SharePullReq _$SharePullReqFromJson(Map<String, dynamic> json) => SharePullReq(
  shareIds: (json['share_ids'] as List<dynamic>)
      .map((e) => (e as num).toInt())
      .toList(),
);

Map<String, dynamic> _$SharePullReqToJson(SharePullReq instance) =>
    <String, dynamic>{'share_ids': instance.shareIds};

PulledShare _$PulledShareFromJson(Map<String, dynamic> json) => PulledShare(
  itemId: (json['item_id'] as num).toInt(),
  filename: json['filename'] as String,
  deviceShare: json['device_share'] as String,
);

SharePullRes _$SharePullResFromJson(Map<String, dynamic> json) => SharePullRes(
  shares: (json['shares'] as List<dynamic>)
      .map((e) => PulledShare.fromJson(e as Map<String, dynamic>))
      .toList(),
);
