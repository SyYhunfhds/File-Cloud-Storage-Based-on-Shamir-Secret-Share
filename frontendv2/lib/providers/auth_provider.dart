import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/user.dart';
import '../utils/constants.dart';

typedef TokenExpiredCallback = void Function();

class AuthState {
  final bool isAuthenticated;
  final User? user;
  final bool isLoading;
  final String? error;

  const AuthState({
    this.isAuthenticated = false,
    this.user,
    this.isLoading = false,
    this.error,
  });

  AuthState copyWith({
    bool? isAuthenticated,
    User? user,
    bool? isLoading,
    String? error,
  }) {
    return AuthState(
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      user: user ?? this.user,
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier() : super(const AuthState()) {
    _initFromStorage();
  }

  static TokenExpiredCallback? onTokenExpired;

  Future<void> _initFromStorage() async {
    final prefs = await SharedPreferences.getInstance();
    final userId = prefs.getInt(Constants.storageUserIdKey);
    final username = prefs.getString(Constants.storageUsernameKey);
    final email = prefs.getString('${Constants.storageUsernameKey}_email');
    final job = prefs.getString('${Constants.storageUsernameKey}_job');
    final privilege = prefs.getInt('${Constants.storageUsernameKey}_privilege');

    if (username != null) {
      final user = User(
        id: userId,
        username: username,
        email: email ?? '',
        job: job,
        privilege: privilege,
      );
      state = state.copyWith(
        isAuthenticated: true,
        user: user,
      );
    }
  }

  Future<void> _saveUserToStorage(User user) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt(Constants.storageUserIdKey, user.id ?? 0);
    await prefs.setString(Constants.storageUsernameKey, user.username);
    await prefs.setString('${Constants.storageUsernameKey}_email', user.email);
    if (user.job != null) {
      await prefs.setString('${Constants.storageUsernameKey}_job', user.job!);
    }
    if (user.privilege != null) {
      await prefs.setInt('${Constants.storageUsernameKey}_privilege', user.privilege!);
    }
  }

  Future<void> _clearUserFromStorage() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(Constants.storageUserIdKey);
    await prefs.remove(Constants.storageUsernameKey);
    await prefs.remove('${Constants.storageUsernameKey}_email');
    await prefs.remove('${Constants.storageUsernameKey}_job');
    await prefs.remove('${Constants.storageUsernameKey}_privilege');
  }

  void setLoading(bool loading) {
    state = state.copyWith(isLoading: loading);
  }

  void setError(String? error) {
    state = state.copyWith(error: error, isLoading: false);
  }

  void clearError() {
    state = state.copyWith(error: null);
  }

  Future<void> loginSuccess(User user) async {
    await _saveUserToStorage(user);
    state = state.copyWith(
      isAuthenticated: true,
      user: user,
      isLoading: false,
      error: null,
    );
  }

  Future<void> logout() async {
    await _clearUserFromStorage();
    state = const AuthState();
    onTokenExpired?.call();
  }

  void updateUser(User user) {
    _saveUserToStorage(user);
    state = state.copyWith(user: user);
  }

  void handleTokenExpired() {
    logout();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier();
});