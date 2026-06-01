import 'package:flutter_test/flutter_test.dart';
import 'package:frontendv2/models/item.dart';

void main() {
  group('ItemInfo JSON序列化测试', () {
    test('API实际响应格式应该能正确解析', () {
      final json = {
        'filename': 'test.txt',
        'owner': 'Christ',
        'uploader': 'Christ',
        'uploaded_at': '2026-05-30T11:43:46.279526Z',
        'changed_at': '2026-05-30T11:43:46.279526Z',
      };

      final item = ItemInfo.fromJson(json);
      expect(item.filename, 'test.txt');
      expect(item.owner, 'Christ');
      expect(item.uploader, 'Christ');
    });

    test('ItemListResponse应该能正确解析', () {
      final json = {
        'count': 5,
        'items': [
          {
            'filename': 'test.txt',
            'owner': 'Christ',
            'uploader': 'Christ',
            'uploaded_at': '2026-05-30T11:43:46.279526Z',
            'changed_at': '2026-05-30T11:43:46.279526Z',
          }
        ]
      };

      final response = ItemListResponse.fromJson(json);
      expect(response.count, 5);
      expect(response.items.length, 1);
      expect(response.items[0].filename, 'test.txt');
    });

    test('空items数组应该能正确解析', () {
      final json = {
        'count': 0,
        'items': [],
      };

      final response = ItemListResponse.fromJson(json);
      expect(response.count, 0);
      expect(response.items.isEmpty, true);
    });
  });
}