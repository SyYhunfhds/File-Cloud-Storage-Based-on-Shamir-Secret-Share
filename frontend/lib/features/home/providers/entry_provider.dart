import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../items/models/item_models.dart';
import '../../items/services/item_api_service.dart';
import '../../auth/providers/auth_provider.dart';
import '../../../core/api_config_provider.dart';

/// 每页条目数量（对齐 API 默认的 size）
const int pageSize = 20;

// =============================================================================
// 条目来源筛选模式
// =============================================================================

enum EntryFilterMode { my, public, all }

// =============================================================================
// 条目列表状态
// =============================================================================

class EntryListState {
  final List<ItemInfo> entries;
  final bool isLoading;
  final String? errorMessage;
  final EntryFilterMode filterMode;
  final int totalCount;
  final int currentPage;

  const EntryListState({
    this.entries = const [],
    this.isLoading = false,
    this.errorMessage,
    this.filterMode = EntryFilterMode.all,
    this.totalCount = 0,
    this.currentPage = 1,
  });

  EntryListState copyWith({
    List<ItemInfo>? entries,
    bool? isLoading,
    String? errorMessage,
    EntryFilterMode? filterMode,
    int? totalCount,
    int? currentPage,
  }) {
    return EntryListState(
      entries: entries ?? this.entries,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      filterMode: filterMode ?? this.filterMode,
      totalCount: totalCount ?? this.totalCount,
      currentPage: currentPage ?? this.currentPage,
    );
  }
}

// =============================================================================
// 条目列表 Notifier（Riverpod 3.x）
// =============================================================================

class EntryListNotifier extends Notifier<EntryListState> {
  @override
  EntryListState build() => const EntryListState();

  /// 总页数（根据 totalCount 和 pageSize 计算）
  int get totalPages {
    if (state.totalCount == 0) return 1;
    return (state.totalCount + pageSize - 1) ~/ pageSize;
  }

  // ---------------------------------------------------------------------------
  // 三个条目来源（映射到 API scope 参数: 1=我的, 2=公开, 3=所有）
  // 切换来源时重置到第 1 页
  // ---------------------------------------------------------------------------

  Future<void> fetchMyEntries() {
    state = state.copyWith(currentPage: 1);
    return _fetchByScope(EntryFilterMode.my, 1);
  }

  Future<void> fetchPublicEntries() {
    state = state.copyWith(currentPage: 1);
    return _fetchByScope(EntryFilterMode.public, 2);
  }

  Future<void> fetchAllEntries() {
    state = state.copyWith(currentPage: 1);
    return _fetchByScope(EntryFilterMode.all, 3);
  }

  /// 跳到指定页（不改变 filterMode）
  Future<void> goToPage(int page) async {
    if (page < 1 || page > totalPages) return;

    final scope = switch (state.filterMode) {
      EntryFilterMode.my => 1,
      EntryFilterMode.public => 2,
      EntryFilterMode.all => 3,
    };
    state = state.copyWith(currentPage: page);
    await _fetchByScope(state.filterMode, scope);
  }

  Future<void> _fetchByScope(EntryFilterMode mode, int scope) async {
    state = state.copyWith(
      isLoading: true,
      errorMessage: null,
      filterMode: mode,
    );

    try {
      final authState = ref.read(authProvider);
      final apiConfig = ref.read(apiConfigProvider);

      if (!authState.isLoggedIn || authState.token.isEmpty) {
        state = state.copyWith(
          entries: const [],
          isLoading: false,
          errorMessage: '请先登录',
        );
        return;
      }

      final service = ItemApiService(apiConfig.baseUrl);
      final apiResp = await service.listItems(
        token: authState.token,
        page: state.currentPage,
        size: pageSize,
        scope: scope,
      );

      if (apiResp.isSuccess && apiResp.data != null) {
        debugPrint('[INFO] 获取条目列表成功: total=${apiResp.data!.total}, items=${apiResp.data!.items.length}');

        state = state.copyWith(
          entries: apiResp.data!.items,
          isLoading: false,
          totalCount: apiResp.data!.total,
        );
      } else {
        debugPrint('[ERROR] 获取条目列表失败: code=${apiResp.code}, message=${apiResp.message}');

        state = state.copyWith(
          entries: const [],
          isLoading: false,
          errorMessage: apiResp.message.isNotEmpty ? apiResp.message : '获取条目列表失败',
        );
      }
    } catch (e, stack) {
      debugPrint('[ERROR] EntryListNotifier._fetchByScope 异常: $e');
      debugPrint('[ERROR] 堆栈: $stack');

      state = state.copyWith(
        entries: const [],
        isLoading: false,
        errorMessage: '网络请求失败，请检查网络连接',
      );
    }
  }

  // ---------------------------------------------------------------------------
  // 条目操作 API 桩（仅预留实现，不实际调用）
  // ---------------------------------------------------------------------------

  Future<DetailItemInfo?> fetchItemDetail(int itemId) async {
    debugPrint('[DEBUG] fetchItemDetail called: itemId=$itemId');
    debugPrint('[TODO] GET /v1/protected/item/detail?item_id=$itemId');
    return null;
  }

  Future<void> updateItem(ItemUpdateReq req) async {
    debugPrint('[DEBUG] updateItem called: filename=${req.filename}');
    debugPrint('[TODO] POST /v1/protected/item/update — body=${jsonEncode(req.toJson())}');
  }

  // ---------------------------------------------------------------------------
  // 本地搜索（按文件名筛选）
  // ---------------------------------------------------------------------------

  List<ItemInfo> searchByFilename(String query) {
    if (query.isEmpty) return state.entries;
    return state.entries
        .where((e) => e.filename.toLowerCase().contains(query.toLowerCase()))
        .toList();
  }
}

/// 条目列表 Provider
final entryListProvider = NotifierProvider<EntryListNotifier, EntryListState>(
  EntryListNotifier.new,
);
