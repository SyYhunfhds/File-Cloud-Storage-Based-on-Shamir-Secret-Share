import 'package:hive/hive.dart';

part 'share_record_data.g.dart';

/// 本地份额存储数据类（Hive TypeAdapter 版本）
@HiveType(typeId: 0)
class ShareRecordData extends HiveObject {
  @HiveField(0)
  final int itemId;

  @HiveField(1)
  final String originalFilename;

  @HiveField(2)
  final String serverFilename;

  @HiveField(3)
  final String encryptedShare;

  @HiveField(4)
  final String encryptedRecoveryCode;

  @HiveField(5)
  final DateTime createdAt;

  ShareRecordData({
    required this.itemId,
    required this.originalFilename,
    required this.serverFilename,
    required this.encryptedShare,
    required this.encryptedRecoveryCode,
    required this.createdAt,
  });
}
