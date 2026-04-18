# 真实邮箱无法获取邮件问题总结

> 日期：2026-04-17
> 涉及模块：email-backend/server/pkg/email/provider/net126.go

## 问题描述

使用126邮箱真实账户进行邮件同步时，无法成功获取邮件数据。系统使用IMAP协议连接`imap.126.com:993`。

## 已知问题点

### 1. 126邮箱安全策略限制

**现象**：
- 126邮箱可能对第三方IMAP客户端有严格的访问限制
- 即使使用正确的授权码，也可能被拒绝访问

**原因分析**：
- 网易邮箱要求使用"授权码"而非密码登录
- 部分账户可能未开启IMAP服务
- 安全策略可能限制非官方客户端

### 2. 代码层面的兼容性处理

已尝试的改进（见net126.go）：

```go
// 126邮箱安全策略可能阻止SELECT，先尝试SELECT再尝试EXAMINE
mailbox, err := p.client.Select("INBOX", true) // true = readonly
if err != nil {
    // SELECT失败时尝试只读方式
    mailbox, err = p.client.Select("INBOX", false)
    ...
}
```

### 3. 认证方式

当前使用：
- 服务器：`imap.126.com:993`
- 认证：`LOGIN` 命令 + 授权码
- TLS：强制SSL连接

## 可能的解决方案

### 方案A：OAuth2认证（推荐但复杂）
- 126邮箱可能需要OAuth2认证
- 需要申请开发者API
- 实现成本较高

### 方案B：调整IMAP参数
```go
// 尝试不同的端口或模式
server: "imap.126.com"
port: 143  // 非SSL端口，再STARTTLS
// 或
port: 993  // 当前SSL端口
```

### 方案C：使用POP3协议
- 126邮箱对POP3限制可能较少
- 需要实现POP3 Provider
- 缺点：无法管理邮件状态

### 方案D：验证账户设置
用户需要确认：
1. 登录126邮箱网页版 → 设置 → POP3/SMTP/IMAP
2. 确认"IMAP/SMTP服务"已开启
3. 生成新的授权码（非登录密码）
4. 可能需要在"客户端授权密码"中单独设置

## 当前状态

由于真实邮箱获取困难，已实现 **Mock Provider** 作为临时方案：

```go
// mock.go - 模拟8封不同类别的邮件
p.MockEmails = []*Email{
    // 紧急工作、普通工作、个人、订阅、通知、营销、会议、任务
}
```

## 测试建议

1. 使用Wireshark抓包分析IMAP交互过程
2. 尝试其他邮箱服务（如Gmail、Outlook）验证代码通用性
3. 联系网易开发者支持确认IMAP访问要求

## 相关文件

- `email-backend/server/pkg/email/provider/net126.go` - 126邮箱Provider实现
- `email-backend/server/pkg/email/provider/mock.go` - Mock Provider（当前使用）
- `email-backend/server/service/sync_service.go` - 同步服务

## 参考资料

- [网易邮箱帮助 - 客户端设置](https://help.mail.163.com/)
- [RFC 3501 - IMAP协议规范](https://tools.ietf.org/html/rfc3501)
- [go-imap库文档](https://github.com/emersion/go-imap)
