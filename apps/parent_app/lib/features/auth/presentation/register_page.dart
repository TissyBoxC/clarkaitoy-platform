import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../application/auth_controller.dart';

/// Guardian registration screen with an explicit consent requirement.
class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _formKey = GlobalKey<FormState>();
  final _displayNameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _hasGuardianConsent = false;
  bool _isSubmitting = false;
  String? _errorMessage;

  @override
  void dispose() {
    _displayNameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }
    if (!_hasGuardianConsent) {
      setState(() => _errorMessage = '请先确认你是孩子的监护人');
      return;
    }
    setState(() {
      _isSubmitting = true;
      _errorMessage = null;
    });
    try {
      await ref
          .read(authControllerProvider.notifier)
          .register(
            email: _emailController.text.trim(),
            password: _passwordController.text,
            displayName: _displayNameController.text.trim(),
          );
      if (mounted) {
        context.go('/family');
      }
    } on Object catch (error) {
      if (mounted) {
        setState(() => _errorMessage = authErrorMessage(error));
      }
    } finally {
      if (mounted) {
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          tooltip: '返回登录',
          onPressed: () => context.go('/login'),
          icon: const Icon(Icons.arrow_back),
        ),
        title: const Text('创建家长账号'),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 440),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    TextFormField(
                      controller: _displayNameController,
                      decoration: const InputDecoration(labelText: '家长称呼'),
                      validator: (value) {
                        if (value == null || value.trim().isEmpty) {
                          return '请输入家长称呼';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _emailController,
                      keyboardType: TextInputType.emailAddress,
                      autofillHints: const [AutofillHints.newUsername],
                      decoration: const InputDecoration(labelText: '邮箱'),
                      validator: (value) {
                        if (value == null || value.trim().isEmpty) {
                          return '请输入邮箱';
                        }
                        if (!value.contains('@')) {
                          return '请输入有效的邮箱';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _passwordController,
                      obscureText: true,
                      autofillHints: const [AutofillHints.newPassword],
                      decoration: const InputDecoration(
                        labelText: '密码',
                        helperText: '至少 8 位，并同时包含字母和数字',
                      ),
                      validator: (value) {
                        if (value == null || value.length < 8) {
                          return '密码至少 8 位';
                        }
                        if (!RegExp(r'[A-Za-z]').hasMatch(value) ||
                            !RegExp(r'[0-9]').hasMatch(value)) {
                          return '密码需要同时包含字母和数字';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    CheckboxListTile(
                      value: _hasGuardianConsent,
                      onChanged: _isSubmitting
                          ? null
                          : (value) => setState(
                              () => _hasGuardianConsent = value ?? false,
                            ),
                      contentPadding: EdgeInsets.zero,
                      controlAffinity: ListTileControlAffinity.leading,
                      title: const Text('我是孩子的监护人，并同意创建家长账号'),
                      subtitle: const Text('孩子使用设备前，请先由家长完成账号和绑定设置。'),
                    ),
                    if (_errorMessage != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _errorMessage!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
                        ),
                      ),
                    ],
                    const SizedBox(height: 20),
                    FilledButton(
                      onPressed: _isSubmitting ? null : _submit,
                      child: Text(_isSubmitting ? '正在创建…' : '创建账号'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
