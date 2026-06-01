import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../shares/models/share_record_data.dart';
import '../../shares/providers/share_providers.dart';
import '../../auth/providers/auth_provider.dart';

/// 份额列表分页数量
const sharePageSize = 15;

/// 份额列表状态
class ShareListState {
  final List<ShareRecordData> shares;
  final int currentPage;
  final int totalCount;
  final int? selectedItemId;
  final bool isLoading;

  const ShareListState({
    this.shares = const [],
    this.currentPage = 1,
    this.totalCount = 0,
    this.selectedItemId,
    this.isLoading = false,
  });

  ShareListState copyWith({
    List<ShareRecordData>? shares,
    int? currentPage,
    int? totalCount,
    int? selectedItemId,
    bool? isLoading,
    bool clearSelection = false,
  }) {
    return ShareListState(
      shares: shares ?? this.shares,
      currentPage: currentPage ?? this.currentPage,
      totalCount: totalCount ?? this.totalCount,
      selectedItemId:
          clearSelection ? null : (selectedItemId ?? this.selectedItemId),
      isLoading: isLoading ?? this.isLoading,
    );
  }

  int get totalPages =>
      totalCount == 0 ? 0 : ((totalCount - 1) ~/ sharePageSize) + 1;

  List<ShareRecordData> get currentPageShares {
    final start = (currentPage - 1) * sharePageSize;
    final end = start + sharePageSize;
    if (start >= shares.length) return [];
    return shares.sublist(start, end.clamp(0, shares.length));
  }
}

/// 份额列表状态管理
class ShareListNotifier extends Notifier<ShareListState> {
  @override
  ShareListState build() {
    ref.listen(authProvider, (prev, next) {
      if (prev?.isLoggedIn == false && next.isLoggedIn) {
        fetch();
      } else if (prev?.isLoggedIn == true && !next.isLoggedIn) {
        state = const ShareListState();
      }
    });
    return const ShareListState();
  }

  Future<void> fetch() async {
    state = state.copyWith(isLoading: true);
    try {
      final service = ref.read(shareServiceProvider);
      final all = await service.listAll();
      state = state.copyWith(
        shares: all,
        totalCount: all.length,
        currentPage: 1,
        isLoading: false,
      );
    } catch (e) {
      debugPrint('[ShareListNotifier] fetch error: $e');
      state = state.copyWith(isLoading: false);
    }
  }

  void goToPage(int page) {
    final totalPages = state.totalPages;
    if (page < 1 || page > totalPages) return;
    state = state.copyWith(currentPage: page, clearSelection: true);
  }

  void select(int? itemId) {
    state = state.copyWith(selectedItemId: itemId);
  }

  Future<void> delete(int itemId) async {
    try {
      final svc = ref.read(shareStorageServiceProvider);
      final auth = ref.read(authProvider);
      final userId = auth.userName.isNotEmpty ? auth.userName : 'anonymous';
      await svc.delete(userId, itemId);
      // 刷新列表
      await fetch();
    } catch (e) {
      debugPrint('[ShareListNotifier] delete error: $e');
    }
  }
}

/// Provider
final shareListProvider =
    NotifierProvider<ShareListNotifier, ShareListState>(
  ShareListNotifier.new,
);
