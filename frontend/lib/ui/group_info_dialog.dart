import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../bloc/chat_cubit.dart';
import '../bloc/conversations_cubit.dart';
import '../core/api_client.dart';
import '../data/chat_repository.dart';
import '../models/models.dart';

class GroupInfoDialog extends StatefulWidget {
  const GroupInfoDialog({
    super.key,
    required this.conversationId,
    required this.groupName,
    required this.initialRole,
  });

  final String conversationId;
  final String groupName;
  final String initialRole;

  @override
  State<GroupInfoDialog> createState() => _GroupInfoDialogState();
}

class _GroupInfoDialogState extends State<GroupInfoDialog> {
  List<GroupMember> members = const [];
  bool loading = true;
  String? error;

  ChatRepository get repository => context.read<ChatCubit>().repository;
  String get currentUserID => context.read<ChatCubit>().tokens.user!.id;
  GroupMember? get currentMember {
    for (final member in members) {
      if (member.userId == currentUserID) return member;
    }
    return null;
  }

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    setState(() {
      loading = true;
      error = null;
    });
    try {
      members = await repository.groupMembers(widget.conversationId);
    } catch (exception) {
      error = readableApiError(exception);
    }
    if (mounted) setState(() => loading = false);
  }

  @override
  Widget build(BuildContext context) {
    final myRole = currentMember?.role ?? widget.initialRole;
    final canAdd = myRole == 'owner' || myRole == 'admin';
    return AlertDialog(
      title: Row(
        children: [
          CircleAvatar(
            child: Text(widget.groupName.characters.first.toUpperCase()),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(widget.groupName, overflow: TextOverflow.ellipsis),
          ),
        ],
      ),
      content: SizedBox(
        width: 480,
        height: 520,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              children: [
                Text(
                  '${members.length} thành viên',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const Spacer(),
                if (canAdd)
                  FilledButton.tonalIcon(
                    onPressed: addMembers,
                    icon: const Icon(Icons.person_add_alt_1),
                    label: const Text('Thêm'),
                  ),
              ],
            ),
            const SizedBox(height: 10),
            if (loading) const LinearProgressIndicator(),
            if (error != null)
              Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    Text(error!),
                    TextButton(onPressed: load, child: const Text('Thử lại')),
                  ],
                ),
              ),
            if (!loading && error == null)
              Expanded(
                child: ListView.builder(
                  itemCount: members.length,
                  itemBuilder: (_, index) => memberTile(members[index], myRole),
                ),
              ),
            if (myRole != 'owner') ...[
              const Divider(),
              OutlinedButton.icon(
                style: OutlinedButton.styleFrom(
                  foregroundColor: Theme.of(context).colorScheme.error,
                ),
                onPressed: leaveGroup,
                icon: const Icon(Icons.logout),
                label: const Text('Rời nhóm'),
              ),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Đóng'),
        ),
      ],
    );
  }

  Widget memberTile(GroupMember member, String myRole) {
    final isMe = member.userId == currentUserID;
    final actions = <PopupMenuEntry<String>>[];
    if (myRole == 'owner' && !isMe && member.role != 'owner') {
      actions.add(
        PopupMenuItem(
          value: member.role == 'admin' ? 'member' : 'admin',
          child: Text(
            member.role == 'admin'
                ? 'Hạ xuống thành viên'
                : 'Cấp quyền quản trị viên',
          ),
        ),
      );
      actions.add(
        const PopupMenuItem(value: 'owner', child: Text('Chuyển quyền sở hữu')),
      );
      actions.add(const PopupMenuDivider());
      actions.add(
        const PopupMenuItem(value: 'remove', child: Text('Xóa khỏi nhóm')),
      );
    } else if (myRole == 'admin' && member.role == 'member' && !isMe) {
      actions.add(
        const PopupMenuItem(value: 'remove', child: Text('Xóa khỏi nhóm')),
      );
    }
    return ListTile(
      leading: CircleAvatar(
        child: Text(member.displayName.characters.first.toUpperCase()),
      ),
      title: Text('${member.displayName}${isMe ? ' (Bạn)' : ''}'),
      subtitle: Text(member.email),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          _RoleBadge(role: member.role),
          if (actions.isNotEmpty)
            PopupMenuButton<String>(
              onSelected: (action) => handleAction(member, action),
              itemBuilder: (_) => actions,
            ),
        ],
      ),
    );
  }

  Future<void> handleAction(GroupMember member, String action) async {
    try {
      if (action == 'admin' || action == 'member') {
        await repository.updateGroupMemberRole(
          widget.conversationId,
          member.userId,
          action,
        );
      } else if (action == 'owner') {
        if (!await confirm(
          'Chuyển quyền sở hữu cho ${member.displayName}? Bạn sẽ trở thành quản trị viên.',
        )) {
          return;
        }
        await repository.transferGroupOwnership(
          widget.conversationId,
          member.userId,
        );
      } else if (action == 'remove') {
        if (!await confirm('Xóa ${member.displayName} khỏi nhóm?')) return;
        await repository.removeGroupMember(
          widget.conversationId,
          member.userId,
        );
      }
      await load();
      if (mounted) await context.read<ConversationsCubit>().load(silent: true);
    } catch (exception) {
      showError(exception);
    }
  }

  Future<void> addMembers() async {
    final selected = await showDialog<List<User>>(
      context: context,
      builder: (_) => _AddGroupMembersDialog(
        repository: repository,
        existingIDs: members.map((member) => member.userId).toSet(),
      ),
    );
    if (selected == null || selected.isEmpty) return;
    try {
      await repository.addGroupMembers(
        widget.conversationId,
        selected.map((user) => user.id).toList(),
      );
      await load();
    } catch (exception) {
      showError(exception);
    }
  }

  Future<void> leaveGroup() async {
    if (!await confirm('Bạn có chắc muốn rời nhóm?')) return;
    try {
      await repository.removeGroupMember(widget.conversationId, currentUserID);
      if (!mounted) return;
      await context.read<ConversationsCubit>().load(silent: true);
      if (!mounted) return;
      final router = GoRouter.of(context);
      Navigator.pop(context);
      router.go('/');
    } catch (exception) {
      showError(exception);
    }
  }

  Future<bool> confirm(String message) async =>
      await showDialog<bool>(
        context: context,
        builder: (_) => AlertDialog(
          title: const Text('Xác nhận'),
          content: Text(message),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Hủy'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Xác nhận'),
            ),
          ],
        ),
      ) ??
      false;

  void showError(Object exception) {
    if (!mounted) return;
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(readableApiError(exception))));
  }
}

class _RoleBadge extends StatelessWidget {
  const _RoleBadge({required this.role});
  final String role;
  @override
  Widget build(BuildContext context) {
    final label = role == 'owner'
        ? 'Chủ nhóm'
        : role == 'admin'
        ? 'Quản trị viên'
        : 'Thành viên';
    return Chip(label: Text(label), visualDensity: VisualDensity.compact);
  }
}

class _AddGroupMembersDialog extends StatefulWidget {
  const _AddGroupMembersDialog({
    required this.repository,
    required this.existingIDs,
  });
  final ChatRepository repository;
  final Set<String> existingIDs;
  @override
  State<_AddGroupMembersDialog> createState() => _AddGroupMembersDialogState();
}

class _AddGroupMembersDialogState extends State<_AddGroupMembersDialog> {
  final query = TextEditingController();
  final Map<String, User> selected = {};
  List<User> results = const [];
  bool loading = false;

  @override
  void dispose() {
    query.dispose();
    super.dispose();
  }

  Future<void> search() async {
    if (query.text.trim().isEmpty) return;
    setState(() => loading = true);
    try {
      final users = await widget.repository.searchUsers(query.text.trim());
      results = users
          .where((user) => !widget.existingIDs.contains(user.id))
          .toList();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Thêm thành viên'),
    content: SizedBox(
      width: 420,
      height: 420,
      child: Column(
        children: [
          TextField(
            controller: query,
            autofocus: true,
            onSubmitted: (_) => search(),
            decoration: InputDecoration(
              hintText: 'Tên hoặc email',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: IconButton(
                onPressed: search,
                icon: const Icon(Icons.arrow_forward),
              ),
            ),
          ),
          if (loading) const LinearProgressIndicator(),
          Expanded(
            child: ListView.builder(
              itemCount: results.length,
              itemBuilder: (_, index) {
                final user = results[index];
                return CheckboxListTile(
                  value: selected.containsKey(user.id),
                  title: Text(user.displayName),
                  subtitle: Text(user.email),
                  onChanged: (checked) => setState(() {
                    checked == true
                        ? selected[user.id] = user
                        : selected.remove(user.id);
                  }),
                );
              },
            ),
          ),
        ],
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.pop(context),
        child: const Text('Hủy'),
      ),
      FilledButton(
        onPressed: selected.isEmpty
            ? null
            : () => Navigator.pop(context, selected.values.toList()),
        child: Text('Thêm (${selected.length})'),
      ),
    ],
  );
}
