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

/// ShareService（按当前登录用户的 userId 和 userName 双标识注入）
final shareServiceProvider = Provider<ShareService>((ref) {
  final auth = ref.watch(authProvider);
  final userId = auth.isLoggedIn
      ? '${auth.userId}_${auth.userName}'
      : 'anonymous_guest';
  return ShareService(
    ref.read(shareStorageServiceProvider),
    userId: userId,
  );
});
