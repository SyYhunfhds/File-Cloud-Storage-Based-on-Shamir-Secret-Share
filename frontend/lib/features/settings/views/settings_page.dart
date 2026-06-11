import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api_config_provider.dart';
import '../../../core/auto_refresh_config_provider.dart';
import '../../../core/theme.dart';
import '../../auth/providers/account_cache_provider.dart';
import '../models/about_model.dart';
import '../services/about_api_service.dart';

/// 设置页面
///
/// 包含"开发者模式"子页面（仅在 Debug 模式下可见），
/// 允许修改 API 后端地址（协议/域名/端口）。
class SettingsPage extends ConsumerWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final apiConfig = ref.watch(apiConfigProvider);
    final autoRefreshConfig = ref.watch(autoRefreshConfigProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('设置'),
        centerTitle: false,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // 主题设置
          Text('主题', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Card(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: Row(
                children: [
                  Icon(Icons.palette_outlined, color: colorScheme.primary),
                  const SizedBox(width: 12),
                  const Text('主题模式'),
                  const Spacer(),
                  SegmentedButton<ThemeMode>(
                    segments: const [
                      ButtonSegment(
                        value: ThemeMode.light,
                        label: Text('亮色', style: TextStyle(fontSize: 12)),
                        icon: Icon(Icons.light_mode, size: 16),
                      ),
                      ButtonSegment(
                        value: ThemeMode.dark,
                        label: Text('暗色', style: TextStyle(fontSize: 12)),
                        icon: Icon(Icons.dark_mode, size: 16),
                      ),
                      ButtonSegment(
                        value: ThemeMode.system,
                        label: Text('跟随系统', style: TextStyle(fontSize: 12)),
                        icon: Icon(Icons.settings, size: 16),
                      ),
                    ],
                    selected: {ref.watch(themeModeProvider)},
                    onSelectionChanged: (selection) {
                      ref.read(themeModeProvider.notifier).set(selection.first);
                    },
                    style: ButtonStyle(
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      visualDensity: VisualDensity.compact,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),

          // 账号管理
          _buildAccountSection(context, ref),
          const SizedBox(height: 24),

          // 当前 API 连接信息
          Text('API 连接', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              leading: Icon(Icons.link, color: colorScheme.primary),
              title: const Text('后端地址'),
              subtitle: Text(apiConfig.baseUrl),
            ),
          ),
          const SizedBox(height: 24),

          // 定时刷新设置组
          Text('定时刷新', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Card(
            child: Column(
              children: [
                SwitchListTile(
                  title: const Text('条目列表自动刷新'),
                  subtitle: Text(
                      '间隔: ${autoRefreshConfig.entryIntervalMinutes} 分钟'),
                  secondary: Icon(Icons.description_outlined,
                      color: colorScheme.primary),
                  value: autoRefreshConfig.entryListEnabled,
                  onChanged: (v) => ref
                      .read(autoRefreshConfigProvider.notifier)
                      .toggleEntryList(v),
                ),
                const Divider(height: 1, indent: 72, endIndent: 16),
                SwitchListTile(
                  title: const Text('份额列表自动刷新'),
                  subtitle: Text(
                      '间隔: ${autoRefreshConfig.shareIntervalMinutes} 分钟'),
                  secondary: Icon(Icons.vpn_key_outlined,
                      color: colorScheme.primary),
                  value: autoRefreshConfig.shareListEnabled,
                  onChanged: (v) => ref
                      .read(autoRefreshConfigProvider.notifier)
                      .toggleShareList(v),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // 关于开发者
          Text('关于', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              leading:
                  Icon(Icons.info_outline, color: colorScheme.primary),
              title: const Text('关于开发者'),
              subtitle: const Text('版本信息与开发团队'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => _showAboutDialog(context, ref),
            ),
          ),
          const SizedBox(height: 24),

          // 开发者模式（仅 Debug 模式可见）
          if (kDebugMode) ...[
            Text(
              '开发者模式',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: colorScheme.error,
                  ),
            ),
            const SizedBox(height: 8),
            const DevModeSettings(),
          ],
        ],
      ),
    );
  }
}

/// 账号管理区块
Widget _buildAccountSection(BuildContext context, WidgetRef ref) {
  final colorScheme = Theme.of(context).colorScheme;
  final accountsAsync = ref.watch(accountCacheProvider);

  return Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text('账号管理', style: Theme.of(context).textTheme.titleMedium),
      const SizedBox(height: 8),
      Card(
        child: accountsAsync.when(
          data: (accounts) {
            if (accounts.isEmpty) {
              return const Padding(
                padding: EdgeInsets.all(24),
                child: Center(
                  child: Text('暂无缓存账号',
                      style: TextStyle(color: Colors.grey)),
                ),
              );
            }
            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                ...accounts.map((a) => ListTile(
                      leading: Icon(Icons.person, color: colorScheme.primary),
                      title: Text(a.username),
                      subtitle: Text(
                          '上次登录: ${a.lastUsedAt.year}-${_pad(a.lastUsedAt.month)}-${_pad(a.lastUsedAt.day)}'),
                      trailing: IconButton(
                        icon: const Icon(Icons.delete_outline, size: 20),
                        onPressed: () => _confirmDeleteAccount(
                            context, ref, a.username),
                      ),
                    )),
                const Divider(height: 1),
                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton.icon(
                    onPressed: () => _confirmClearAll(context, ref),
                    icon: const Icon(Icons.delete_sweep, size: 18),
                    label: const Text('清空全部'),
                  ),
                ),
              ],
            );
          },
          loading: () => const Padding(
            padding: EdgeInsets.all(24),
            child: Center(child: CircularProgressIndicator(strokeWidth: 2)),
          ),
          error: (e, _) => Padding(
            padding: const EdgeInsets.all(24),
            child: Center(child: Text('加载失败: $e')),
          ),
        ),
      ),
    ],
  );
}

void _confirmDeleteAccount(
    BuildContext context, WidgetRef ref, String username) async {
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('删除缓存账号'),
      content: Text('确定要删除「$username」的缓存数据吗？'),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消')),
        TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('删除')),
      ],
    ),
  );
  if (confirmed == true) {
    ref.read(accountCacheProvider.notifier).deleteAccount(username);
  }
}

void _confirmClearAll(BuildContext context, WidgetRef ref) async {
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('清空所有缓存'),
      content: const Text('确定要清空所有缓存的账号数据吗？'),
      actions: [
        TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('取消')),
        TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('清空')),
      ],
    ),
  );
  if (confirmed == true) {
    ref.read(accountCacheProvider.notifier).deleteAll();
  }
}

String _pad(int n) => n.toString().padLeft(2, '0');

/// 弹出"关于开发者"对话框
void _showAboutDialog(BuildContext context, WidgetRef ref) {
  showDialog(
    context: context,
    builder: (_) => const _AboutDeveloperDialog(),
  );
}

/// 关于开发者对话框
///
/// 调用 GET /v1/about 获取版本号和开发团队信息并展示。
class _AboutDeveloperDialog extends ConsumerStatefulWidget {
  const _AboutDeveloperDialog();

  @override
  ConsumerState<_AboutDeveloperDialog> createState() =>
      _AboutDeveloperDialogState();
}

class _AboutDeveloperDialogState
    extends ConsumerState<_AboutDeveloperDialog> {
  AboutInfo? _aboutInfo;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _fetchAbout();
  }

  Future<void> _fetchAbout() async {
    try {
      final config = ref.read(apiConfigProvider);
      final service = AboutApiService(config.baseUrl);
      final resp = await service.getAbout();
      if (mounted) {
        setState(() {
          if (resp.isSuccess && resp.data != null) {
            _aboutInfo = resp.data;
          } else {
            _error = resp.message.isNotEmpty ? resp.message : '获取信息失败';
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      debugPrint('[TestConnection] $e');
      if (mounted) {
        setState(() {
          _error = '连接失败';
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return AlertDialog(
      title: const Text('关于开发者'),
      content: SizedBox(
        width: 360,
        child: _isLoading
            ? const Center(
                child: Padding(
                  padding: EdgeInsets.all(24),
                  child: CircularProgressIndicator(),
                ),
              )
            : _error != null
                ? Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.error_outline,
                          size: 40, color: colorScheme.error),
                      const SizedBox(height: 12),
                      Text(_error!,
                          style: TextStyle(color: colorScheme.error)),
                      const SizedBox(height: 16),
                      FilledButton.tonal(
                        onPressed: () {
                          setState(() {
                            _isLoading = true;
                            _error = null;
                          });
                          _fetchAbout();
                        },
                        child: const Text('重试'),
                      ),
                    ],
                  )
                : _buildInfo(colorScheme, textTheme),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('关闭'),
        ),
      ],
    );
  }

  Widget _buildInfo(ColorScheme colorScheme, TextTheme textTheme) {
    final info = _aboutInfo!;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 版本号
        Row(
          children: [
            Icon(Icons.tag, size: 18, color: colorScheme.primary),
            const SizedBox(width: 8),
            Text('版本', style: textTheme.labelMedium),
            const Spacer(),
            Text(info.version,
                style: textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: colorScheme.primary,
                )),
          ],
        ),
        const Divider(height: 24),
        // 项目领导人
        Row(
          children: [
            Icon(Icons.star, size: 18, color: colorScheme.primary),
            const SizedBox(width: 8),
            Text('项目领导人', style: textTheme.labelMedium),
            const Spacer(),
            Text(info.leader, style: textTheme.bodyMedium),
          ],
        ),
        const SizedBox(height: 16),
        // 开发者团队
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Icon(Icons.people, size: 18,
                  color: colorScheme.primary),
            ),
            const SizedBox(width: 8),
            Text('开发团队', style: textTheme.labelMedium),
          ],
        ),
        const SizedBox(height: 8),
        ...info.developers.map(
          (dev) => Padding(
            padding: const EdgeInsets.only(left: 26, bottom: 4),
            child: Text('• $dev',
                style: textTheme.bodyMedium?.copyWith(
                  color: colorScheme.onSurfaceVariant,
                )),
          ),
        ),
      ],
    );
  }
}

/// 开发者模式设置卡片
///
/// 允许用户动态修改 API 后端地址（协议、域名、端口）。
/// 此组件仅在 `kDebugMode` 为 `true` 时被渲染。
class DevModeSettings extends ConsumerWidget {
  const DevModeSettings({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colorScheme = Theme.of(context).colorScheme;
    final apiConfig = ref.watch(apiConfigProvider);

    final hostController =
        TextEditingController(text: apiConfig.host);
    final portController =
        TextEditingController(text: apiConfig.port.toString());

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.developer_mode, size: 20, color: colorScheme.error),
                const SizedBox(width: 8),
                Text(
                  '修改后端地址',
                  style: TextStyle(
                    fontWeight: FontWeight.w600,
                    color: colorScheme.error,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              '修改将在保存后立即对所有 API 请求生效。',
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 16),

            // 协议选择
            DropdownButtonFormField<String>(
              initialValue: apiConfig.protocol,
              decoration: const InputDecoration(
                labelText: '协议',
                border: OutlineInputBorder(),
                isDense: true,
              ),
              items: const [
                DropdownMenuItem(value: 'http', child: Text('http')),
                DropdownMenuItem(value: 'https', child: Text('https')),
              ],
              onChanged: (value) {
                if (value != null) {
                  ref.read(apiConfigProvider.notifier).update(
                        protocol: value,
                        host: apiConfig.host,
                        port: apiConfig.port,
                      );
                }
              },
            ),
            const SizedBox(height: 12),

            // 域名输入
            TextField(
              controller: hostController,
              decoration: const InputDecoration(
                labelText: '域名',
                hintText: 'localhost',
                border: OutlineInputBorder(),
                isDense: true,
              ),
            ),
            const SizedBox(height: 12),

            // 端口输入
            TextField(
              controller: portController,
              decoration: const InputDecoration(
                labelText: '端口',
                hintText: '8000',
                border: OutlineInputBorder(),
                isDense: true,
              ),
              keyboardType: TextInputType.number,
            ),
            const SizedBox(height: 16),

            // 保存按钮
            Align(
              alignment: Alignment.centerRight,
              child: FilledButton.tonal(
                onPressed: () {
                  final host = hostController.text.trim();
                  final portStr = portController.text.trim();
                  final port = int.tryParse(portStr) ?? apiConfig.port;

                  if (host.isEmpty) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('域名不能为空')),
                    );
                    return;
                  }

                  ref.read(apiConfigProvider.notifier).update(
                        protocol: apiConfig.protocol,
                        host: host,
                        port: port,
                      );

                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('API 地址已更新'),
                      duration: Duration(seconds: 1),
                    ),
                  );
                },
                child: const Text('保存'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
