import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/item.dart';

/// 审计条目状态
class AuditState {
  final List<ItemInfo> items;
  final bool isLoading;
  final String? error;
  final ItemInfo? selectedItem;

  const AuditState({
    this.items = const [],
    this.isLoading = false,
    this.error,
    this.selectedItem,
  });

  AuditState copyWith({
    List<ItemInfo>? items,
    bool? isLoading,
    String? error,
    ItemInfo? selectedItem,
  }) {
    return AuditState(
      items: items ?? this.items,
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
      selectedItem: selectedItem ?? this.selectedItem,
    );
  }
}

/// 审计条目状态管理
class AuditNotifier extends StateNotifier<AuditState> {
  AuditNotifier() : super(const AuditState());

  void setLoading(bool loading) {
    state = state.copyWith(isLoading: loading);
  }

  void setItems(List<ItemInfo> items) {
    state = state.copyWith(items: items, isLoading: false, error: null);
  }

  void setError(String error) {
    state = state.copyWith(error: error, isLoading: false);
  }

  void selectItem(ItemInfo item) {
    state = state.copyWith(selectedItem: item);
  }

  void clearSelection() {
    state = state.copyWith(selectedItem: null);
  }
}

final auditProvider = StateNotifierProvider<AuditNotifier, AuditState>((ref) {
  return AuditNotifier();
});
