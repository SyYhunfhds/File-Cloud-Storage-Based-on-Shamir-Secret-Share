import 'package:flutter_riverpod/flutter_riverpod.dart';

/// 自动刷新配置
///
/// 分别控制条目列表和份额列表的定时刷新开关与间隔。
class AutoRefreshConfig {
  final bool entryListEnabled;
  final bool shareListEnabled;
  final int entryIntervalMinutes;
  final int shareIntervalMinutes;

  const AutoRefreshConfig({
    this.entryListEnabled = false,
    this.shareListEnabled = false,
    this.entryIntervalMinutes = 5,
    this.shareIntervalMinutes = 5,
  });

  AutoRefreshConfig copyWith({
    bool? entryListEnabled,
    bool? shareListEnabled,
    int? entryIntervalMinutes,
    int? shareIntervalMinutes,
  }) {
    return AutoRefreshConfig(
      entryListEnabled: entryListEnabled ?? this.entryListEnabled,
      shareListEnabled: shareListEnabled ?? this.shareListEnabled,
      entryIntervalMinutes:
          entryIntervalMinutes ?? this.entryIntervalMinutes,
      shareIntervalMinutes:
          shareIntervalMinutes ?? this.shareIntervalMinutes,
    );
  }
}

/// 自动刷新配置状态管理
class AutoRefreshConfigNotifier extends Notifier<AutoRefreshConfig> {
  @override
  AutoRefreshConfig build() => const AutoRefreshConfig();

  void toggleEntryList(bool enabled) {
    state = state.copyWith(entryListEnabled: enabled);
  }

  void toggleShareList(bool enabled) {
    state = state.copyWith(shareListEnabled: enabled);
  }

  void setEntryInterval(int minutes) {
    state = state.copyWith(entryIntervalMinutes: minutes);
  }

  void setShareInterval(int minutes) {
    state = state.copyWith(shareIntervalMinutes: minutes);
  }
}

/// 自动刷新配置 Provider
final autoRefreshConfigProvider =
    NotifierProvider<AutoRefreshConfigNotifier, AutoRefreshConfig>(
  AutoRefreshConfigNotifier.new,
);
