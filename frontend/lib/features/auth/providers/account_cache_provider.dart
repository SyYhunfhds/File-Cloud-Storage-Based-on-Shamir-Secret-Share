import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/account_cache_model.dart';
import '../services/account_cache_service.dart';

/// 账号缓存 Provider
final accountCacheProvider = AsyncNotifierProvider<AccountCacheNotifier,
    List<CachedAccount>>(AccountCacheNotifier.new);

class AccountCacheNotifier extends AsyncNotifier<List<CachedAccount>> {
  AccountCacheService get _service => AccountCacheService();

  @override
  Future<List<CachedAccount>> build() async => _service.loadAccounts();

  Future<void> refresh() async {
    state = AsyncData(await _service.loadAccounts());
  }

  Future<void> deleteAccount(String username) async {
    await _service.deleteAccount(username);
    await refresh();
  }

  Future<void> deleteAll() async {
    await _service.deleteAll();
    await refresh();
  }
}
