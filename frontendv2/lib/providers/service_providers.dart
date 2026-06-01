import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/api_service.dart';
import '../services/storage_service.dart';

final apiServiceProvider = Provider<ApiService>((ref) {
  final apiService = ApiService();
  apiService.init();
  return apiService;
});

final storageServiceProvider = Provider<StorageService>((ref) {
  final storageService = StorageService();
  storageService.init();
  return storageService;
});