"""
126邮箱Provider测试说明

由于126邮箱需要真实的授权码才能连接IMAP服务器，
本测试脚本用于验证Provider的基本功能，不实际连接邮箱。

要测试实际连接，请按以下步骤操作：
1. 登录126邮箱网页版
2. 进入 设置 → POP3/SMTP/IMAP
3. 开启 IMAP/SMTP 服务
4. 生成"授权码"（16位字母数字）
5. 使用授权码替换下面的MOCK_CREDENTIAL

测试命令:
cd email-backend
go run test_126_provider.go
"""

# 测试配置示例
TEST_CONFIG = {
    "email": "your_email@126.com",
    "credential": "your_authorization_code",  # 授权码，非登录密码
    "server": "imap.126.com",
    "port": 993,
}

# 注意事项
NOTES = """
126邮箱IMAP连接注意事项：

1. 授权码 vs 登录密码
   - 126邮箱必须使用"授权码"而非登录密码
   - 授权码在邮箱设置中生成，16位字母数字

2. 客户端标识 (IMAP ID)
   - 126邮箱要求发送IMAP ID命令
   - 否则会被标记为"Unsafe Login"
   - 已在Net126Provider中实现

3. SSL/TLS
   - 必须使用SSL连接 (端口993)
   - 已在Provider中配置

4. 错误处理
   - Provider内置重试机制 (默认3次)
   - 区分认证错误和网络错误

5. 已实现功能
   - Connect: 连接邮箱服务器
   - FetchEmailList: 获取邮件列表
   - FetchEmailDetail: 获取邮件详情
   - FetchEmails: 批量获取邮件
   - TestConnection: 测试连接状态
"""

if __name__ == "__main__":
    print(__doc__)
    print(NOTES)
