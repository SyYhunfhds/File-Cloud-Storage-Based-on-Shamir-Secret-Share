import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../services/share_storage_service.dart';
import '../services/share_service.dart';

// =============================================================================
// Provider 层
// =============================================================================

/// ShareStorageService（单例）
final shareStorageServiceProvider = Provider<ShareStorageService>((ref) {
  return ShareStorageService();
});

/// ShareService（按当前登录用户的 userId 注入）
final shareServiceProvider = Provider<ShareService>((ref) {
  final auth = ref.watch(authProvider);
  final userId = auth.userName.isNotEmpty ? auth.userName : 'anonymous';
  return ShareService(
    ref.read(shareStorageServiceProvider),
    userId: userId,
  );
});
