"""126 IMAP 多种认证方式测试"""
import imaplib
import ssl
import base64

def test_auth_methods(email, auth_code):
    print(f"Email: {email}")
    print(f"Auth Code: {auth_code}")
    print("=" * 50)

    context = ssl.create_default_context()
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE

    # Test 1: 用用户名部分（不含@126.com）
    print("\n[1] 尝试用户名: pt0819 (不含@126.com)...")
    try:
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        typ, data = mail.login('pt0819', auth_code)
        print(f"    成功! {typ} {data}")
        mail.logout()
    except Exception as e:
        print(f"    失败: {e}")

    # Test 2: 完整邮箱地址
    print("\n[2] 尝试用户名: pt0819@126.com...")
    try:
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        typ, data = mail.login(email, auth_code)
        print(f"    成功! {typ} {data}")
        mail.logout()
    except Exception as e:
        print(f"    失败: {e}")

    # Test 3: XOAUTH2 方式
    print("\n[3] 尝试 XOAUTH2 认证...")
    try:
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        # XOAUTH2 格式
        auth_string = f"user={email}\x01auth=Bearer {auth_code}\x01\x01"
        auth_bytes = base64.b64encode(auth_string.encode()).decode()
        typ, data = mail.authenticate('XOAUTH2', lambda x: auth_bytes)
        print(f"    成功! {typ} {data}")
        mail.logout()
    except Exception as e:
        print(f"    失败: {e}")

    # Test 4: PLAIN SASL
    print("\n[4] 尝试 SASL PLAIN 认证...")
    try:
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        auth_string = f"\x00{email}\x00{auth_code}"
        auth_bytes = base64.b64encode(auth_string.encode()).decode()
        typ, data = mail.authenticate('PLAIN', lambda x: auth_bytes)
        print(f"    成功! {typ} {data}")
        mail.logout()
    except Exception as e:
        print(f"    失败: {e}")

    # Test 5: 尝试 POP3
    print("\n[5] 尝试 POP3 (pop.126.com:995)...")
    try:
        import poplib
        pop = poplib.POP3_SSL('pop.126.com', 995, context=context)
        typ, data = pop.user(email)
        print(f"    USER: {typ} {data}")
        typ, data = pop.pass_(auth_code)
        print(f"    PASS: 成功! 邮件数: {data}")
        pop.quit()
    except Exception as e:
        print(f"    失败: {e}")

if __name__ == "__main__":
    import sys
    if len(sys.argv) < 3:
        print("Usage: python test_auth.py <email> <auth_code>")
        sys.exit(1)
    test_auth_methods(sys.argv[1], sys.argv[2])