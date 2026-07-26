import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../bloc/auth_cubit.dart';
import '../bloc/conversations_cubit.dart';
import '../models/models.dart';

class ConversationsPage extends StatefulWidget {
  const ConversationsPage({super.key});
  @override
  State<ConversationsPage> createState() => _ConversationsPageState();
}

class _ConversationsPageState extends State<ConversationsPage> {
  @override
  void initState() {
    super.initState();
    context.read<ConversationsCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    final desktop = MediaQuery.sizeOf(context).width >= 800;
    return Scaffold(
      body: SafeArea(
        child: Row(
          children: [
            SizedBox(
              width: desktop ? 380 : MediaQuery.sizeOf(context).width,
              child: const ConversationPane(),
            ),
            if (desktop) const Expanded(child: _EmptyChat()),
          ],
        ),
      ),
    );
  }
}

class ConversationPane extends StatelessWidget {
  const ConversationPane({super.key, this.selectedId});
  final String? selectedId;

  @override
  Widget build(BuildContext context) => Material(
    color: Colors.white,
    child: Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(18, 12, 10, 8),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  'Đoạn chat',
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              IconButton(
                tooltip: 'Tin nhắn mới',
                onPressed: () => showDialog(
                  context: context,
                  builder: (_) => BlocProvider.value(
                    value: context.read<ConversationsCubit>(),
                    child: const NewChatDialog(),
                  ),
                ),
                icon: const Icon(Icons.edit_square),
              ),
              PopupMenuButton<String>(
                onSelected: (value) {
                  if (value == 'logout') context.read<AuthCubit>().logout();
                },
                itemBuilder: (_) => const [
                  PopupMenuItem(value: 'logout', child: Text('Đăng xuất')),
                ],
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          child: TextField(
            readOnly: true,
            onTap: () => showDialog(
              context: context,
              builder: (_) => BlocProvider.value(
                value: context.read<ConversationsCubit>(),
                child: const NewChatDialog(),
              ),
            ),
            decoration: const InputDecoration(
              hintText: 'Tìm kiếm người dùng',
              prefixIcon: Icon(Icons.search),
              filled: true,
              border: OutlineInputBorder(
                borderSide: BorderSide.none,
                borderRadius: BorderRadius.all(Radius.circular(24)),
              ),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Expanded(
          child: BlocConsumer<ConversationsCubit, ConversationsState>(
            listener: (context, state) {
              if (state.error != null) {
                ScaffoldMessenger.of(
                  context,
                ).showSnackBar(SnackBar(content: Text(state.error!)));
              }
            },
            builder: (context, state) {
              if (state.loading && state.items.isEmpty) {
                return const Center(child: CircularProgressIndicator());
              }
              if (state.items.isEmpty) {
                return const Center(
                  child: Padding(
                    padding: EdgeInsets.all(24),
                    child: Text(
                      'Chưa có cuộc trò chuyện.\nHãy tìm một người để bắt đầu.',
                      textAlign: TextAlign.center,
                    ),
                  ),
                );
              }
              return RefreshIndicator(
                onRefresh: () => context.read<ConversationsCubit>().load(),
                child: ListView.builder(
                  itemCount: state.items.length,
                  itemBuilder: (context, index) {
                    final item = state.items[index];
                    return ListTile(
                      selected: item.id == selectedId,
                      selectedTileColor: Theme.of(
                        context,
                      ).colorScheme.primaryContainer.withValues(alpha: .45),
                      leading: CircleAvatar(
                        radius: 25,
                        child: Text(
                          item.name.isEmpty
                              ? '?'
                              : item.name.characters.first.toUpperCase(),
                        ),
                      ),
                      title: Text(
                        item.name,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                      subtitle: Text(
                        item.lastMessage?.type == 'image'
                            ? '📷 Hình ảnh'
                            : item.lastMessage?.text ?? 'Bắt đầu trò chuyện',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      trailing: item.unreadCount > 0
                          ? Badge(label: Text('${item.unreadCount}'))
                          : null,
                      onTap: () {
                        final location = '/chat/${item.id}';
                        if (MediaQuery.sizeOf(context).width < 800) {
                          context.push(location, extra: item);
                        } else {
                          context.go(location, extra: item);
                        }
                      },
                    );
                  },
                ),
              );
            },
          ),
        ),
      ],
    ),
  );
}

class NewChatDialog extends StatefulWidget {
  const NewChatDialog({super.key});
  @override
  State<NewChatDialog> createState() => _NewChatDialogState();
}

class _NewChatDialogState extends State<NewChatDialog> {
  final query = TextEditingController();
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
      results = await context.read<ConversationsCubit>().repository.searchUsers(
        query.text.trim(),
      );
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: const Text('Tin nhắn mới'),
    content: SizedBox(
      width: 420,
      height: 360,
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
          const SizedBox(height: 10),
          if (loading) const LinearProgressIndicator(),
          Expanded(
            child: ListView.builder(
              itemCount: results.length,
              itemBuilder: (_, index) {
                final user = results[index];
                return ListTile(
                  leading: CircleAvatar(
                    child: Text(
                      user.displayName.characters.first.toUpperCase(),
                    ),
                  ),
                  title: Text(user.displayName),
                  subtitle: Text(user.email),
                  onTap: () async {
                    final conversation = await context
                        .read<ConversationsCubit>()
                        .createDirect(user);
                    if (!context.mounted || conversation == null) return;
                    Navigator.pop(context);
                    context.push(
                      '/chat/${conversation.id}',
                      extra: conversation,
                    );
                  },
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
        child: const Text('Đóng'),
      ),
    ],
  );
}

class _EmptyChat extends StatelessWidget {
  const _EmptyChat();
  @override
  Widget build(BuildContext context) => const Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        CircleAvatar(
          radius: 38,
          child: Icon(Icons.chat_bubble_outline, size: 38),
        ),
        SizedBox(height: 16),
        Text(
          'Chọn một cuộc trò chuyện để bắt đầu',
          style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
        ),
      ],
    ),
  );
}
