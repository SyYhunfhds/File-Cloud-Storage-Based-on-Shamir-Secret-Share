import 'package:flutter/foundation.dart';
import 'package:hive/hive.dart';
import '../models/share_record_data.dart';

/// 份额本地存储服务（基于 Hive，纯 Dart 零原生编译依赖）
///
/// - 按 `userId_userName` 创建独立 Box：`shares_{userId}_{userName}`
/// - 退登时只关闭内存引用，保留磁盘文件
/// - 重新登录时打开同名 Box，数据自动恢复
/// - Box 懒打开：首次访问时调用 Hive.openBox()，非首次复用缓存
class ShareStorageService {
  /// 已打开的 Box 缓存（boxKey → Box）
  static final Map<String, Box<ShareRecordData>> _boxes = {};

  /// 生成隔离 Box key：`shares_{userId}_{userName}`
  static String _boxKey(String userId) => 'shares_$userId';

  // ===========================================================================
  // Box 获取（懒打开）
  // ===========================================================================

  /// 获取用户 Box，若未打开则调用 Hive.openBox() 懒打开
  Future<Box<ShareRecordData>> _ensureBox(String userId) async {
    final key = _boxKey(userId);
    if (_boxes.containsKey(key)) return _boxes[key]!;
    final box = await Hive.openBox<ShareRecordData>(key);
    _boxes[key] = box;
    debugPrint('[HiveStorage] 已打开 Box: $key');
    return box;
  }

  // ===========================================================================
  // 保存
  // ===========================================================================

  Future<void> save(String userId, ShareRecordData data) async {
    final box = await _ensureBox(userId);
    await box.put('${data.itemId}', data);
    debugPrint('[HiveStorage] 已保存 itemId=${data.itemId} (userId=$userId)');
  }

  // ===========================================================================
  // 查询
  // ===========================================================================

  Future<ShareRecordData?> get(String userId, int itemId) async {
    final box = await _ensureBox(userId);
    return box.get('$itemId');
  }

  // ===========================================================================
  // 列表
  // ===========================================================================

  Future<List<ShareRecordData>> listAll(String userId) async {
    final box = await _ensureBox(userId);
    final values = box.values.toList();
    // 按创建时间降序
    values.sort((a, b) => b.createdAt.compareTo(a.createdAt));
    return values;
  }

  // ===========================================================================
  // 删除
  // ===========================================================================

  Future<void> delete(String userId, int itemId) async {
    final box = await _ensureBox(userId);
    await box.delete('$itemId');
    debugPrint('[HiveStorage] 已删除 itemId=$itemId (userId=$userId)');
  }

  /// 份额总数
  Future<int> count(String userId) async {
    final box = await _ensureBox(userId);
    return box.length;
  }

  // ===========================================================================
  // 生命周期
  // ===========================================================================

  /// 退登时关闭用户 Box 的内存引用（文件保留磁盘）
  Future<void> closeBox(String userId) async {
    final key = _boxKey(userId);
    final box = _boxes.remove(key);
    if (box != null && box.isOpen) {
      await box.close();
      debugPrint('[HiveStorage] 已关闭 Box: $key');
    }
  }
}