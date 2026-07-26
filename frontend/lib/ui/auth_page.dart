import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../bloc/auth_cubit.dart';

class AuthPage extends StatefulWidget {
  const AuthPage({super.key, this.registerMode = false});
  final bool registerMode;
  @override
  State<AuthPage> createState() => _AuthPageState();
}

class _AuthPageState extends State<AuthPage> {
  final formKey = GlobalKey<FormState>();
  final name = TextEditingController();
  final email = TextEditingController();
  final password = TextEditingController();
  bool obscure = true;
  @override
  void dispose() {
    name.dispose();
    email.dispose();
    password.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    body: Center(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 420),
          child: Card(
            elevation: 0,
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Form(
                key: formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const CircleAvatar(
                      radius: 32,
                      child: Icon(Icons.forum_rounded, size: 34),
                    ),
                    const SizedBox(height: 18),
                    Text(
                      widget.registerMode ? 'Tạo tài khoản' : 'Đăng nhập',
                      textAlign: TextAlign.center,
                      style: Theme.of(context).textTheme.headlineSmall
                          ?.copyWith(fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 28),
                    if (widget.registerMode) ...[
                      TextFormField(
                        controller: name,
                        decoration: const InputDecoration(
                          labelText: 'Tên hiển thị',
                          prefixIcon: Icon(Icons.person_outline),
                        ),
                        validator: required,
                      ),
                      const SizedBox(height: 14),
                    ],
                    TextFormField(
                      controller: email,
                      keyboardType: TextInputType.emailAddress,
                      decoration: const InputDecoration(
                        labelText: 'Email',
                        prefixIcon: Icon(Icons.mail_outline),
                      ),
                      validator: required,
                    ),
                    const SizedBox(height: 14),
                    TextFormField(
                      controller: password,
                      obscureText: obscure,
                      decoration: InputDecoration(
                        labelText: 'Mật khẩu',
                        prefixIcon: const Icon(Icons.lock_outline),
                        suffixIcon: IconButton(
                          onPressed: () => setState(() => obscure = !obscure),
                          icon: Icon(
                            obscure
                                ? Icons.visibility_outlined
                                : Icons.visibility_off_outlined,
                          ),
                        ),
                      ),
                      validator: (value) => (value?.length ?? 0) < 8
                          ? 'Mật khẩu cần ít nhất 8 ký tự'
                          : null,
                    ),
                    const SizedBox(height: 22),
                    BlocConsumer<AuthCubit, AuthState>(
                      listener: (context, state) {
                        if (state is AuthFailure) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(content: Text(state.message)),
                          );
                        }
                      },
                      builder: (context, state) => FilledButton(
                        onPressed: state is AuthLoading ? null : submit,
                        child: Padding(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          child: state is AuthLoading
                              ? const SizedBox.square(
                                  dimension: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : Text(
                                  widget.registerMode ? 'Đăng ký' : 'Đăng nhập',
                                ),
                        ),
                      ),
                    ),
                    TextButton(
                      onPressed: () => context.go(
                        widget.registerMode ? '/login' : '/register',
                      ),
                      child: Text(
                        widget.registerMode
                            ? 'Đã có tài khoản? Đăng nhập'
                            : 'Chưa có tài khoản? Đăng ký',
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    ),
  );

  String? required(String? value) =>
      value == null || value.trim().isEmpty ? 'Không được để trống' : null;
  void submit() {
    if (!formKey.currentState!.validate()) return;
    final cubit = context.read<AuthCubit>();
    widget.registerMode
        ? cubit.register(name.text, email.text, password.text)
        : cubit.login(email.text, password.text);
  }
}
