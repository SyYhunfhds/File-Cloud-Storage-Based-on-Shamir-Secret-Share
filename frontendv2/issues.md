# 以下结果均在Windows热重载模式下测试得到

## 份额提交功能
- 预期功能已恢复

## 份额查看功能
```
[ERROR:flutter/runtime/dart_vm_initializer.cc(40)] Unhandled Exception: Fo
rmatException: Unexpected character (at character 1)
Vd9}*0n5z09n95gw06}};n5g454gs+}4"`w~o9;9dw5x0@5w09n90nsz69FnIo...
^

#0      _ChunkedJsonParser.fail (dart:convert-patch/convert_patch.dart:146
7:5)
#1      _ChunkedJsonParser.parseNumber (dart:convert-patch/convert_patch.d
art:1333:9)
#2      _ChunkedJsonParser.parse (dart:convert-patch/convert_patch.dart:93
5:22)
#3      _parseJson (dart:convert-patch/convert_patch.dart:35:10)
#4      JsonDecoder.convert (dart:convert/json.dart:641:36)
#5      JsonCodec.decode (dart:convert/json.dart:223:41)
#6      jsonDecode (dart:convert/json.dart:160:12)
#7      StorageService.getShares (package:frontendv2/services/storage_serv
ice.dart:84:32)
#8      _SharePageState._loadShares (package:frontendv2/pages/share_page.d
art:37:42)
<asynchronous suspension>
```
- 附上报错提示
怎么这又给我json解析了？我都说了份额是**Base64编码**的字符串, 而桌面端你的实现是异或加密, 你异或解密回去就行了, 你为什么要json解析啊？！