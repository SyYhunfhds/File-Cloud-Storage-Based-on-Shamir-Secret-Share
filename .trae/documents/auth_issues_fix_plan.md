# 认证和路由问题修复计划

## 问题分析

1. **路由没有正确响应认证状态变化：
   - `routerProvider` 使用普通 Provider 创建 GoRouter，在创建时只读取一次 authState
   - 之后认证状态变化时，路由器不会重新初始化
   - 导致登录成功后无法正确重定向逻辑失效

## 修复方案

1. 修改 `router.dart`：
   - 使用 `refreshListenable` 让路由器在认证状态变化时更新
   - 或者使用 `Listenable 来正确处理认证状态

2. 确保认证成功后添加路由跳转显式跳转逻辑

## 需要修改的文件

- `f:\GolangWorkspace\CryptographyDesign\front4largescreen\lib\router.dart`
