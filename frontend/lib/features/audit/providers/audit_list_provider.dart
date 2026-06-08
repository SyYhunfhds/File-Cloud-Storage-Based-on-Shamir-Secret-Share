import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api_config_provider.dart';
import '../../auth/providers/auth_provider.dart';
import '../models/audit_models.dart';
import '../services/audit_api_service.dart';

/// 审计列表不可变状态
class AuditListState {
  final List<LessDetailedAudit> audits;
  final bool isLoading;
  final String? errorMessage;
  final AuditFilterScope filterScope;
  final int totalCount;
  final int currentPage;

  static const int pageSize = 20;

  const AuditListState({
    this.audits = const [],
    this.isLoading = false,
    this.errorMessage,
    this.filterScope = AuditFilterScope.all,
    this.totalCount = 0,
    this.currentPage = 1,
  });

  int get totalPages =>
      totalCount > 0 ? (totalCount + pageSize - 1) ~/ pageSize : 1;

  AuditListState copyWith({
    List<LessDetailedAudit>? audits,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
    AuditFilterScope? filterScope,
    int? totalCount,
    int? currentPage,
  }) {
    return AuditListState(
      audits: audits ?? this.audits,
      isLoading: isLoading ?? this.isLoading,
      errorMessage:
          clearError ? null : (errorMessage ?? this.errorMessage),
      filterScope: filterScope ?? this.filterScope,
      totalCount: totalCount ?? this.totalCount,
      currentPage: currentPage ?? this.currentPage,
    );
  }
}

/// 审计列表状态管理
class AuditListNotifier extends Notifier<AuditListState> {
  @override
  AuditListState build() => const AuditListState();

  /// 切换筛选范围并重新加载第一页
  void setFilter(AuditFilterScope scope) {
    if (scope != state.filterScope) {
      state = state.copyWith(filterScope: scope, currentPage: 1);
      fetch(silent: true);
    }
  }

  /// 翻页
  Future<void> goToPage(int page) async {
    if (page < 1 || page > state.totalPages || page == state.currentPage) {
      return;
    }
    state = state.copyWith(currentPage: page);
    await fetch(silent: true);
  }

  /// 核心 fetch 流水线
  ///
  /// [silent] 为 true 时跳过 `isLoading = true`，适用于筛选切换和翻页，
  /// 避免全屏重建导致的闪烁。首次加载和手动刷新时建议传 false。
  Future<void> fetch({bool silent = false}) async {
    final auth = ref.read(authProvider);
    if (!auth.isLoggedIn) {
      state = state.copyWith(errorMessage: '请先登录', clearError: false);
      return;
    }

    if (!silent) {
      state = state.copyWith(isLoading: true, clearError: true);
    }

    try {
      final config = ref.read(apiConfigProvider);
      final service = AuditApiService(config.baseUrl);
      final resp = await service.listAudits(
        token: auth.token,
        page: state.currentPage,
        size: AuditListState.pageSize,
        scope: state.filterScope.toScopeList(),
      );

      if (resp.isSuccess && resp.data != null) {
        final data = resp.data;
        if (data == null) return;
        state = state.copyWith(
          audits: data.list,
          totalCount: data.total,
          isLoading: false,
        );
      } else {
        state = state.copyWith(
          audits: silent ? state.audits : [],
          isLoading: silent ? state.isLoading : false,
          errorMessage:
              resp.message.isNotEmpty ? resp.message : '获取审计列表失败',
        );
      }
    } catch (e) {
        state = state.copyWith(
          audits: silent ? state.audits : [],
          isLoading: silent ? state.isLoading : false,
          errorMessage: '连接失败',
        );
    }
  }

  /// 执行审计操作（通过/拒绝条目下的待审核记录）
  Future<void> performOperation({
    required int itemId,
    required List<int> auditIds,
    required String action,
  }) async {
    final auth = ref.read(authProvider);
    if (!auth.isLoggedIn) {
      state = state.copyWith(errorMessage: '请先登录');
      return;
    }
    if (auditIds.isEmpty) return;

    try {
      final config = ref.read(apiConfigProvider);
      final service = AuditApiService(config.baseUrl);
      final operations = {for (final id in auditIds) id: action};
      final resp = await service.operateAudit(
        token: auth.token,
        operations: operations,
      );

      if (resp.isSuccess) {
        await fetch();
      } else {
        state = state.copyWith(
          errorMessage: resp.message.isNotEmpty ? resp.message : '操作失败',
        );
      }
    } catch (e) {
      state = state.copyWith(errorMessage: '连接失败');
    }
  }
}

/// 审计列表 Provider
final auditListProvider =
    NotifierProvider<AuditListNotifier, AuditListState>(
  AuditListNotifier.new,
);
