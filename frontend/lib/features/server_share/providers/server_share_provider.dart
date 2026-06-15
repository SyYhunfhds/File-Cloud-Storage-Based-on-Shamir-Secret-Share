import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../../shares/providers/share_providers.dart';
import '../../../core/api_config_provider.dart';
import '../models/server_share_models.dart';
import '../services/share_api_service.dart';

/// 服务器份额列表状态
class ServerShareState {
  final List<ServerShareInfo> shares;
  final int total;
  final int currentPage;
  final int pageSize;
  final bool isLoading;
  final String? errorMessage;
  final Set<int> selectedShareIds;
  final bool isPulling;
  final List<PulledShare> pulledShares;
  final String? pullError;

  const ServerShareState({
    this.shares = const [],
    this.total = 0,
    this.currentPage = 1,
    this.pageSize = 20,
    this.isLoading = false,
    this.errorMessage,
    this.selectedShareIds = const {},
    this.isPulling = false,
    this.pulledShares = const [],
    this.pullError,
  });

  ServerShareState copyWith({
    List<ServerShareInfo>? shares,
    int? total,
    int? currentPage,
    int? pageSize,
    bool? isLoading,
    String? errorMessage,
    Set<int>? selectedShareIds,
    bool? isPulling,
    List<PulledShare>? pulledShares,
    String? pullError,
  }) {
    return ServerShareState(
      shares: shares ?? this.shares,
      total: total ?? this.total,
      currentPage: currentPage ?? this.currentPage,
      pageSize: pageSize ?? this.pageSize,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
      selectedShareIds: selectedShareIds ?? this.selectedShareIds,
      isPulling: isPulling ?? this.isPulling,
      pulledShares: pulledShares ?? this.pulledShares,
      pullError: pullError,
    );
  }
}

/// 服务器份额 Notifier
class ServerShareNotifier extends Notifier<ServerShareState> {
  late ShareApiService _apiService;

  @override
  ServerShareState build() {
    _apiService = ShareApiService(ref.read(apiConfigProvider).baseUrl);
    return const ServerShareState();
  }

  // ===========================================================================
  // 列表加载
  // ===========================================================================

  Future<void> fetch() => loadPage(1);

  Future<void> loadPage(int page) async {
    final authState = ref.read(authProvider);
    if (!authState.isLoggedIn || authState.token.isEmpty) {
      state = state.copyWith(
        errorMessage: '请先登录',
        isLoading: false,
      );
      return;
    }

    state = state.copyWith(isLoading: true, errorMessage: null);

    try {
      final resp = await _apiService.list(
        token: authState.token,
        page: page,
        size: state.pageSize,
      );

      if (resp.isSuccess && resp.data != null) {
        state = state.copyWith(
          shares: resp.data!.list,
          total: resp.data!.total,
          currentPage: page,
          isLoading: false,
          selectedShareIds: {},
        );
      } else {
        state = state.copyWith(
          errorMessage: resp.message.isNotEmpty ? resp.message : '加载份额列表失败',
          isLoading: false,
        );
      }
    } catch (e) {
      debugPrint('[ServerShare] 加载列表异常: $e');
      state = state.copyWith(
        errorMessage: '网络请求失败，请检查网络连接',
        isLoading: false,
      );
    }
  }

  // ===========================================================================
  // 多选
  // ===========================================================================

  void toggleSelection(int shareId) {
    final selected = Set<int>.from(state.selectedShareIds);
    if (selected.contains(shareId)) {
      selected.remove(shareId);
    } else {
      selected.add(shareId);
    }
    state = state.copyWith(selectedShareIds: selected);
  }

  void selectAll() {
    if (state.shares.isEmpty) return;
    final allIds = state.shares.map((s) => s.shareId).toSet();
    state = state.copyWith(selectedShareIds: allIds);
  }

  void clearSelection() {
    state = state.copyWith(selectedShareIds: {});
  }

  // ===========================================================================
  // 拉取份额
  // ===========================================================================

  Future<void> pullSelected() async {
    if (state.selectedShareIds.isEmpty) return;

    final authState = ref.read(authProvider);
    if (!authState.isLoggedIn || authState.token.isEmpty) {
      state = state.copyWith(pullError: '请先登录');
      return;
    }

    state = state.copyWith(isPulling: true, pullError: null);

    try {
      final resp = await _apiService.pull(
        token: authState.token,
        shareIds: state.selectedShareIds.toList(),
      );

      if (resp.isSuccess && resp.data != null) {
        // 拉取成功 → 遍历结果存入 Hive
        final shareService = ref.read(shareServiceProvider);
        for (final pulled in resp.data!.shares) {
          try {
            await shareService.saveShare(
              itemId: pulled.itemId,
              originalFilename: pulled.filename,
              serverFilename: pulled.filename,
              share: pulled.deviceShare,
              recoveryCode: '',
            );
          } catch (e) {
            debugPrint('[ServerShare] 保存份额到Hive失败 itemId=${pulled.itemId}: $e');
          }
        }

        state = state.copyWith(
          isPulling: false,
          pulledShares: resp.data!.shares,
          pullError: null,
        );
      } else {
        state = state.copyWith(
          isPulling: false,
          pulledShares: [],
          pullError: resp.message.isNotEmpty ? resp.message : '拉取份额失败',
        );
      }
    } catch (e) {
      debugPrint('[ServerShare] 拉取异常: $e');
      state = state.copyWith(
        isPulling: false,
        pulledShares: [],
        pullError: '网络请求失败，请检查网络连接',
      );
    }
  }

  void resetPulled() {
    state = state.copyWith(
      pulledShares: [],
      pullError: null,
    );
  }
}

/// 服务器份额 Provider
final serverShareProvider =
    NotifierProvider<ServerShareNotifier, ServerShareState>(
  ServerShareNotifier.new,
);
