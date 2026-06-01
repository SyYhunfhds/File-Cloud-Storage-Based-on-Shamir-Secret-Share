/// 将字节数格式化为人类可读的文件大小字符串
///
/// Examples:
///   500     → "500 B"
///   2048    → "2.00 KB"
///   1572864 → "1.50 MB"
String formatFileSize(int bytes) {
  if (bytes < 1024) {
    return '$bytes B';
  } else if (bytes < 1024 * 1024) {
    final kb = bytes / 1024;
    return '${kb.toStringAsFixed(2)} KB';
  } else {
    final mb = bytes / (1024 * 1024);
    return '${mb.toStringAsFixed(2)} MB';
  }
}
