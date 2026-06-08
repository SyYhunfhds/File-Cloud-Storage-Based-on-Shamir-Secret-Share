import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/audit_models.dart';
import '../providers/audit_list_provider.dart';

/// 审计筛选栏 — FilterChip 行
class AuditFilterBar extends ConsumerWidget {
  const AuditFilterBar({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(auditListProvider);

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Wrap(
        spacing: 8,
        children: AuditFilterScope.values.map((scope) {
          final selected = scope == state.filterScope;
          return FilterChip(
            label: Text(scope.label),
            selected: selected,
            onSelected: (_) =>
                ref.read(auditListProvider.notifier).setFilter(scope),
            selectedColor:
                Theme.of(context).colorScheme.primaryContainer,
            checkmarkColor: Theme.of(context).colorScheme.primary,
          );
        }).toList(),
      ),
    );
  }
}
