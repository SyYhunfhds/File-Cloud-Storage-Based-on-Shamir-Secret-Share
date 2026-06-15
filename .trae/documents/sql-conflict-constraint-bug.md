# SQL ON CONFLICT 约束不匹配问题分析

## 问题定位

### 数据库唯一约束
从截图可以看到，`shares` 表的唯一约束 `unique_share` 包含 **4 列**：
- `item_id`
- `user_id`
- `share_type`
- `status`

### SQL 语句中的 ON CONFLICT
代码生成的 SQL 只指定了 **3 列**：
```sql
ON CONFLICT (item_id,user_id,share_type)
```

### 代码来源
```go
// item_v1_item_submit.go
.OnConflict("item_id", "user_id", "share_type").Save()
```

## 问题后果

当执行这条 SQL 时，PostgreSQL 会报错：
```
pq: there is no unique or exclusion constraint matching the ON CONFLICT specification
```

因为 `(item_id, user_id, share_type)` 这个列组合没有对应的唯一约束，唯一约束是 `(item_id, user_id, share_type, status)`。

## 修复方案

### 方案 A：修改代码（推荐）
将 `OnConflict` 的列参数改为包含 `status`：
```go
.OnConflict("item_id", "user_id", "share_type", "status").Save()
```

### 方案 B：修改数据库约束
如果业务逻辑上不需要 `status` 作为唯一约束的一部分，可以修改数据库表定义，移除 `status` 列：
```sql
DROP INDEX IF EXISTS unique_share;
CREATE UNIQUE INDEX unique_share ON shares (item_id, user_id, share_type);
```

## 业务逻辑分析

从业务角度看：
- `item_id + user_id + share_type` 应该唯一标识一条份额记录
- `status` 表示份额状态（active/inactive），不应该参与唯一约束
- 同一个用户对同一个条目的同一种份额类型应该只有一条记录，不管状态如何

**推荐方案 A**，因为业务逻辑上 `status` 不应该作为唯一约束的一部分。
