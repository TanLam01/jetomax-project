import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import 'bloc/auth_cubit.dart';
import 'bloc/chat_cubit.dart';
import 'bloc/conversations_cubit.dart';
import 'core/dependencies.dart';
import 'core/token_store.dart';
import 'data/chat_repository.dart';
import 'models/models.dart';
import 'realtime/realtime_client.dart';
import 'ui/auth_page.dart';
import 'ui/chat_page.dart';
import 'ui/conversations_page.dart';

class ChatApp extends StatefulWidget {
  const ChatApp({super.key, required this.dependencies});
  final AppDependencies dependencies;
  @override
  State<ChatApp> createState() => _ChatAppState();
}

class _ChatAppState extends State<ChatApp> {
  late final AuthCubit auth;
  late final ConversationsCubit conversations;
  late final GoRouter router;
  late final StreamRefreshNotifier refresh;

  @override
  void initState() {
    super.initState();
    final getIt = widget.dependencies.getIt;
    auth = AuthCubit(
      getIt<ChatRepository>(),
      getIt<TokenStore>(),
      getIt<RealtimeClient>(),
    )..initialize();
    conversations = ConversationsCubit(
      getIt<ChatRepository>(),
      getIt<RealtimeClient>(),
    );
    refresh = StreamRefreshNotifier(auth.stream);
    router = GoRouter(
      initialLocation: '/',
      refreshListenable: refresh,
      redirect: (context, state) {
        final signedIn = auth.state is AuthSignedIn;
        final authRoute =
            state.matchedLocation == '/login' ||
            state.matchedLocation == '/register';
        if (!signedIn && !authRoute) return '/login';
        if (signedIn && authRoute) return '/';
        return null;
      },
      routes: [
        GoRoute(path: '/login', builder: (_, __) => const AuthPage()),
        GoRoute(
          path: '/register',
          builder: (_, __) => const AuthPage(registerMode: true),
        ),
        GoRoute(path: '/', builder: (_, __) => const ConversationsPage()),
        GoRoute(
          path: '/chat/:id',
          builder: (_, state) {
            final conversation = state.extra as Conversation?;
            return BlocProvider(
              create: (_) => ChatCubit(
                conversationId: state.pathParameters['id']!,
                repository: getIt<ChatRepository>(),
                realtime: getIt<RealtimeClient>(),
                tokens: getIt<TokenStore>(),
                peerUserId: conversation?.peerUserId ?? '',
              )..load(),
              child: ChatPage(
                conversationId: state.pathParameters['id']!,
                title: conversation?.name ?? 'Cuộc trò chuyện',
              ),
            );
          },
        ),
      ],
    );
  }

  @override
  void dispose() {
    router.dispose();
    refresh.dispose();
    conversations.close();
    auth.close();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => MultiBlocProvider(
    providers: [
      BlocProvider.value(value: auth),
      BlocProvider.value(value: conversations),
    ],
    child: MaterialApp.router(
      title: 'Jetomax Chat',
      debugShowCheckedModeBanner: false,
      routerConfig: router,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF0866FF),
          brightness: Brightness.light,
        ),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF7F8FA),
      ),
    ),
  );
}

class StreamRefreshNotifier extends ChangeNotifier {
  StreamRefreshNotifier(Stream<Object?> stream) {
    _subscription = stream.asBroadcastStream().listen((_) => notifyListeners());
  }
  late final StreamSubscription<Object?> _subscription;
  @override
  void dispose() {
    _subscription.cancel();
    super.dispose();
  }
}
