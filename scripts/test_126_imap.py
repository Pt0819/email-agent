"""
测试126邮箱IMAP连接
用于验证授权码是否正确配置

使用方法：
1. 登录 mail.126.com
2. 设置 → POP3/SMTP/IMAP → 开启IMAP服务
3. 生成"客户端授权码"
4. 运行此脚本测试
"""
import imaplib
import ssl

def test_126_connection(email: str, auth_code: str):
    """
    测试126邮箱IMAP连接

    Args:
        email: 126邮箱地址
        auth_code: 客户端授权码（不是登录密码！）
    """
    print(f"测试连接: {email}")
    print(f"服务器: imap.126.com:993")
    print("-" * 40)

    try:
        # 创建SSL连接
        context = ssl.create_default_context()
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        print("✓ SSL连接成功")

        # 登录（使用授权码）
        mail.login(email, auth_code)
        print("✓ 登录成功！授权码有效")

        # 选择收件箱
        status, messages = mail.select('INBOX')
        print(f"✓ 选择收件箱成功，邮件数: {messages[0].decode()}")

        # 搜索最近5封邮件
        status, email_ids = mail.search(None, 'ALL')
        if status == 'OK':
            ids = email_ids[0].split()
            recent_ids = ids[-5:] if len(ids) >= 5 else ids
            print(f"✓ 找到 {len(ids)} 封邮件，最近 {len(recent_ids)} 封")

            # 获取主题
            for email_id in recent_ids:
                status, msg_data = mail.fetch(email_id, '(BODY.PEEK[HEADER.FIELDS (SUBJECT FROM DATE)])')
                if status == 'OK':
                    header = msg_data[0][1].decode('utf-8', errors='ignore')
                    print(f"  - {header[:80]}...")

        mail.logout()
        print("-" * 40)
        print("✅ 连接测试完全成功！")
        return True

    except imaplib.IMAP4.error as e:
        print(f"❌ IMAP错误: {e}")
        if "authentication failed" in str(e).lower() or "login" in str(e).lower():
            print("\n可能的原因：")
            print("1. 未使用'客户端授权码'（不是登录密码）")
            print("2. 授权码输入错误")
            print("3. IMAP服务未开启")
        return False

    except Exception as e:
        print(f"❌ 连接错误: {e}")
        return False


if __name__ == "__main__":
    print("=" * 50)
    print("126邮箱IMAP连接测试工具")
    print("=" * 50)
    print()
    print("⚠️  重要提示：")
    print("   密码字段请填写'客户端授权码'，不是邮箱登录密码！")
    print("   获取方式：mail.126.com → 设置 → POP3/SMTP/IMAP")
    print()

    email = input("请输入126邮箱地址: ").strip()
    auth_code = input("请输入客户端授权码: ").strip()

    print()
    test_126_connection(email, auth_code)
