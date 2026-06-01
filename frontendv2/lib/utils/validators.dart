/// 表单验证工具类
class Validators {
  /// 验证用户名
  /// - 长度3-24个字符
  /// - 只能包含字母、数字、下划线
  static String? validateUsername(String? value) {
    if (value == null || value.isEmpty) {
      return '用户名不能为空';
    }
    if (value.length < 3 || value.length > 24) {
      return '用户名长度必须在3-24个字符之间';
    }
    final regex = RegExp(r'^[a-zA-Z0-9_]+$');
    if (!regex.hasMatch(value)) {
      return '用户名只能包含字母、数字和下划线';
    }
    return null;
  }

  /// 验证邮箱
  static String? validateEmail(String? value) {
    if (value == null || value.isEmpty) {
      return '邮箱不能为空';
    }
    final regex = RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$');
    if (!regex.hasMatch(value)) {
      return '请输入有效的邮箱地址';
    }
    return null;
  }

  /// 验证密码
  /// - 长度6-24个字符
  /// - 必须包含大小写字母和数字
  static String? validatePassword(String? value) {
    if (value == null || value.isEmpty) {
      return '密码不能为空';
    }
    if (value.length < 6 || value.length > 24) {
      return '密码长度必须在6-24个字符之间';
    }
    // 检查是否包含大小写字母和数字
    final hasUpper = RegExp(r'[A-Z]').hasMatch(value);
    final hasLower = RegExp(r'[a-z]').hasMatch(value);
    final hasNumber = RegExp(r'[0-9]').hasMatch(value);
    if (!hasUpper || !hasLower || !hasNumber) {
      return '密码必须包含大小写字母和数字';
    }
    return null;
  }

  /// 验证确认密码
  static String? validateConfirmPassword(String? value, String? original) {
    if (value != original) {
      return '两次输入的密码不一致';
    }
    return null;
  }

  /// 验证非空
  static String? validateRequired(String? value, [String fieldName = '']) {
    if (value == null || value.isEmpty) {
      return '${fieldName.isNotEmpty ? fieldName : '该项'}不能为空';
    }
    return null;
  }
}
