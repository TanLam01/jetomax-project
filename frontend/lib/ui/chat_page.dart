import 'dart:async';

import 'package:cached_network_image/cached_network_image.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../bloc/chat_cubit.dart';
import '../data/chat_repository.dart';
import '../models/models.dart';
import 'conversations_page.dart';

class ChatPage extends StatelessWidget {
  const ChatPage({
    super.key,
    required this.conversationId,
    required this.title,
  });
  final String conversationId;
  final String title;

  @override
  Widget build(BuildContext context) {
    final desktop = MediaQuery.sizeOf(context).width >= 800;
    return Scaffold(
      body: SafeArea(
        child: Row(
          children: [
            if (desktop)
              SizedBox(
                width: 380,
                child: ConversationPane(selectedId: conversationId),
              ),
            Expanded(child: ChatPanel(title: title)),
          ],
        ),
      ),
    );
  }
}

class ChatPanel extends StatefulWidget {
  const ChatPanel({super.key, required this.title});
  final String title;
  @override
  State<ChatPanel> createState() => _ChatPanelState();
}

class _ChatPanelState extends State<ChatPanel> {
  final input = TextEditingController();
  final scroll = ScrollController();
  int lastCount = 0;

  @override
  void initState() {
    super.initState();
    scroll.addListener(() {
      if (scroll.hasClients && scroll.position.pixels < 100) {
        context.read<ChatCubit>().loadOlder();
      }
    });
  }

  @override
  void dispose() {
    input.dispose();
    scroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Material(
    color: const Color(0xFFF7F8FA),
    child: Column(
      children: [
        BlocBuilder<ChatCubit, ChatState>(
          builder: (context, state) => Container(
            color: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
            child: Row(
              children: [
                if (MediaQuery.sizeOf(context).width < 800) const BackButton(),
                CircleAvatar(
                  child: Text(
                    widget.title.isEmpty
                        ? '?'
                        : widget.title.characters.first.toUpperCase(),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        widget.title,
                        style: const TextStyle(
                          fontWeight: FontWeight.w700,
                          fontSize: 16,
                        ),
                      ),
                      Text(
                        state.typingUsers.isNotEmpty
                            ? 'Đang nhập...'
                            : state.peerOnline
                            ? 'Đang hoạt động'
                            : 'Ngoại tuyến',
                        style: TextStyle(
                          fontSize: 12,
                          color: state.peerOnline ? Colors.green : Colors.grey,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  onPressed: () {},
                  icon: const Icon(Icons.info_outline),
                ),
              ],
            ),
          ),
        ),
        Expanded(
          child: BlocConsumer<ChatCubit, ChatState>(
            listener: (context, state) {
              if (state.error != null) {
                ScaffoldMessenger.of(
                  context,
                ).showSnackBar(SnackBar(content: Text(state.error!)));
              }
              if (state.messages.length > lastCount) {
                lastCount = state.messages.length;
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  if (scroll.hasClients) {
                    scroll.animateTo(
                      scroll.position.maxScrollExtent,
                      duration: const Duration(milliseconds: 250),
                      curve: Curves.easeOut,
                    );
                  }
                });
              }
            },
            builder: (context, state) {
              if (state.loading && state.messages.isEmpty) {
                return const Center(child: CircularProgressIndicator());
              }
              return ListView.builder(
                controller: scroll,
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 18,
                ),
                itemCount: state.messages.length + (state.loadingOlder ? 1 : 0),
                itemBuilder: (context, index) {
                  if (state.loadingOlder && index == 0) {
                    return const Center(
                      child: Padding(
                        padding: EdgeInsets.all(8),
                        child: CircularProgressIndicator(),
                      ),
                    );
                  }
                  final offset = state.loadingOlder ? index - 1 : index;
                  return MessageBubble(message: state.messages[offset]);
                },
              );
            },
          ),
        ),
        _Composer(controller: input),
      ],
    ),
  );
}

class MessageBubble extends StatelessWidget {
  const MessageBubble({super.key, required this.message});
  final Message message;

  @override
  Widget build(BuildContext context) {
    final mine =
        message.senderId == AppDependenciesHolder.of(context).tokensUserId;
    final bubbleColor = mine
        ? Theme.of(context).colorScheme.primary
        : Colors.white;
    final textColor = mine ? Colors.white : Colors.black87;
    return Align(
      alignment: mine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 420),
        margin: const EdgeInsets.symmetric(vertical: 3),
        padding: message.type == 'image'
            ? const EdgeInsets.all(4)
            : const EdgeInsets.symmetric(horizontal: 14, vertical: 9),
        decoration: BoxDecoration(
          color: bubbleColor,
          borderRadius: BorderRadius.circular(18),
          boxShadow: mine
              ? null
              : const [BoxShadow(color: Color(0x11000000), blurRadius: 4)],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            if (message.type == 'image' && message.attachment != null)
              _AttachmentImage(attachment: message.attachment!),
            if (message.text.isNotEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                child: Text(message.text, style: TextStyle(color: textColor)),
              ),
            if (mine)
              Padding(
                padding: const EdgeInsets.only(right: 4, bottom: 2),
                child: Icon(
                  message.status == MessageStatus.pending
                      ? Icons.schedule
                      : message.status == MessageStatus.failed
                      ? Icons.error_outline
                      : Icons.done,
                  size: 13,
                  color: Colors.white70,
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _AttachmentImage extends StatelessWidget {
  const _AttachmentImage({required this.attachment});
  final Attachment attachment;
  @override
  Widget build(BuildContext context) => FutureBuilder<String>(
    future: AppDependenciesHolder.of(
      context,
    ).repository.attachmentUrl(attachment.id),
    builder: (_, snapshot) {
      if (!snapshot.hasData) {
        return const SizedBox(
          width: 240,
          height: 160,
          child: Center(child: CircularProgressIndicator()),
        );
      }
      return ClipRRect(
        borderRadius: BorderRadius.circular(15),
        child: CachedNetworkImage(
          imageUrl: snapshot.data!,
          width: 280,
          height: 220,
          fit: BoxFit.cover,
          errorWidget: (_, __, ___) => const SizedBox(
            width: 240,
            height: 160,
            child: Icon(Icons.broken_image_outlined),
          ),
        ),
      );
    },
  );
}

class _Composer extends StatelessWidget {
  const _Composer({required this.controller});
  final TextEditingController controller;

  @override
  Widget build(BuildContext context) => Container(
    color: Colors.white,
    padding: const EdgeInsets.fromLTRB(8, 8, 12, 12),
    child: SafeArea(
      top: false,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          IconButton(
            tooltip: 'Gửi ảnh',
            onPressed: () => pickImage(context),
            icon: const Icon(Icons.image_outlined),
          ),
          Expanded(
            child: TextField(
              controller: controller,
              minLines: 1,
              maxLines: 5,
              onChanged: (_) => context.read<ChatCubit>().typing(),
              onSubmitted: (_) => send(context),
              decoration: const InputDecoration(
                hintText: 'Aa',
                filled: true,
                border: OutlineInputBorder(
                  borderSide: BorderSide.none,
                  borderRadius: BorderRadius.all(Radius.circular(22)),
                ),
              ),
            ),
          ),
          IconButton(
            onPressed: () => send(context),
            icon: Icon(
              Icons.send_rounded,
              color: Theme.of(context).colorScheme.primary,
            ),
          ),
        ],
      ),
    ),
  );

  void send(BuildContext context) {
    context.read<ChatCubit>().sendText(controller.text);
    controller.clear();
  }

  Future<void> pickImage(BuildContext context) async {
    final result = await FilePicker.pickFiles(
      type: FileType.custom,
      allowedExtensions: const ['jpg', 'jpeg', 'png', 'webp'],
      withData: true,
    );
    if (result == null || !context.mounted) return;
    final file = result.files.single;
    if (file.bytes == null) return;
    final extension = file.extension?.toLowerCase();
    final mime = extension == 'png'
        ? 'image/png'
        : extension == 'webp'
        ? 'image/webp'
        : 'image/jpeg';
    await context.read<ChatCubit>().sendImage(file.name, mime, file.bytes!);
  }
}

// Keeps UI widgets independent from the global service locator while remaining lightweight.
class AppDependenciesHolder {
  AppDependenciesHolder._(this.repository, this.tokensUserId);
  final ChatRepository repository;
  final String tokensUserId;
  static AppDependenciesHolder of(BuildContext context) {
    final cubit = context.read<ChatCubit>();
    return AppDependenciesHolder._(
      cubit.repository,
      cubit.tokens.user?.id ?? '',
    );
  }
}
