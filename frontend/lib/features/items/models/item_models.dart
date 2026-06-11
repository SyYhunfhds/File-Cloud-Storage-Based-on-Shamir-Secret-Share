// =============================================================================
// 条目数据模型 — 对齐后端 api.md 定义
// =============================================================================

/// 列表渲染用条目信息（对应后端 `ItemInfo`）
class ItemInfo {
  final int itemId;
  final String filename;
  final String owner;
  final String uploader;
  final bool canDownload;
  final DateTime uploadedAt;
  final DateTime changedAt;

  const ItemInfo({
    required this.itemId,
    required this.filename,
    required this.owner,
    required this.uploader,
    required this.canDownload,
    required this.uploadedAt,
    required this.changedAt,
  });

  factory ItemInfo.fromJson(Map<String, dynamic> json) {
    return switch (json) {
      {
        'item_id': int itemId,
        'filename': String filename,
        'owner': String owner,
        'uploader': String uploader,
        'can_download': bool canDownload,
        'uploaded_at': String uploadedAt,
        'changed_at': String changedAt,
      } =>
        ItemInfo(
          itemId: itemId,
          filename: filename,
          owner: owner,
          uploader: uploader,
          canDownload: canDownload,
          uploadedAt: DateTime.parse(uploadedAt),
          changedAt: DateTime.parse(changedAt),
        ),
      _ => throw const FormatException('Failed to parse ItemInfo'),
    };
  }

  Map<String, dynamic> toJson() => {
        'item_id': itemId,
        'filename': filename,
        'owner': owner,
        'uploader': uploader,
        'can_download': canDownload,
        'uploaded_at': uploadedAt.toIso8601String(),
        'changed_at': changedAt.toIso8601String(),
      };
}

/// 条目列表响应
class ItemListRes {
  final int total;
  final List<ItemInfo> items;

  const ItemListRes({required this.total, required this.items});

  factory ItemListRes.fromJson(Map<String, dynamic> json) {
    final itemsRaw = json['items'] as List<dynamic>;
    return ItemListRes(
      total: json['total'] as int? ?? 0,
      items: itemsRaw
          .map((e) => ItemInfo.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}

/// 条目详情/修改页面用（对应后端 `DetailItemInfo`）
class DetailItemInfo {
  final int itemId;
  final String filename;
  final String owner;
  final String uploader;
  final int minimumPrivilege;
  final bool isPublic;
  final DateTime uploadedAt;
  final DateTime changedAt;

  const DetailItemInfo({
    required this.itemId,
    required this.filename,
    required this.owner,
    required this.uploader,
    required this.minimumPrivilege,
    required this.isPublic,
    required this.uploadedAt,
    required this.changedAt,
  });

  factory DetailItemInfo.fromJson(Map<String, dynamic> json) {
    return switch (json) {
      {
        'item_id': int itemId,
        'filename': String filename,
        'owner': String owner,
        'uploader': String uploader,
        'minimum_privilege': int minimumPrivilege,
        'is_public': bool isPublic,
        'uploaded_at': String uploadedAt,
        'changed_at': String changedAt,
      } =>
        DetailItemInfo(
          itemId: itemId,
          filename: filename,
          owner: owner,
          uploader: uploader,
          minimumPrivilege: minimumPrivilege,
          isPublic: isPublic,
          uploadedAt: DateTime.parse(uploadedAt),
          changedAt: DateTime.parse(changedAt),
        ),
      _ => throw const FormatException('Failed to parse DetailItemInfo'),
    };
  }

  Map<String, dynamic> toJson() => {
        'item_id': itemId,
        'filename': filename,
        'owner': owner,
        'uploader': uploader,
        'minimum_privilege': minimumPrivilege,
        'is_public': isPublic,
        'uploaded_at': uploadedAt.toIso8601String(),
        'changed_at': changedAt.toIso8601String(),
      };
}

/// 条目修改请求（对应后端 `ItemUpdateReq`）
class ItemUpdateReq {
  final int itemId;
  final String? newFilename;
  final int? minimumPrivilege;
  final bool? enablePublic;

  const ItemUpdateReq({
    required this.itemId,
    this.newFilename,
    this.minimumPrivilege,
    this.enablePublic,
  });

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{'item_id': itemId};
    if (newFilename != null) map['new_filename'] = newFilename;
    if (minimumPrivilege != null) map['minimum_privilege'] = minimumPrivilege;
    if (enablePublic != null) map['enable_public'] = enablePublic;
    return map;
  }
}

/// 条目删除响应（对应后端 `ItemDeleteRes`）
class ItemDeleteRes {
  final int totalDeleted;

  const ItemDeleteRes({required this.totalDeleted});

  factory ItemDeleteRes.fromJson(Map<String, dynamic> json) =>
      ItemDeleteRes(
          totalDeleted: (json['total_deleted'] as num?)?.toInt() ?? 0);
}

/// 申请查看权限请求（对应后端 `ApplyForViewingReq`）
class ApplyForViewingReq {
  final List<int> itemIds;

  const ApplyForViewingReq({required this.itemIds});

  Map<String, dynamic> toJson() => {'item_ids': itemIds};
}

/// 申请查看权限响应（对应后端 `ApplyForViewingRes`）
class ApplyForViewingRes {
  final int totalApplied;

  const ApplyForViewingRes({required this.totalApplied});

  factory ApplyForViewingRes.fromJson(Map<String, dynamic> json) =>
      ApplyForViewingRes(
          totalApplied: (json['total_applied'] as num?)?.toInt() ?? 0);
}

/// 条目详情响应包装（对应后端 `GetOneItemRes`）
class GetOneItemRes {
  final DetailItemInfo item;

  const GetOneItemRes({required this.item});

  factory GetOneItemRes.fromJson(Map<String, dynamic> json) =>
      GetOneItemRes(
          item: DetailItemInfo.fromJson(json['item'] as Map<String, dynamic>));
}
