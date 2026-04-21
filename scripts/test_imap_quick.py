"""快速测试126邮箱IMAP连接 - 非交互式"""
import imaplib
import ssl
import sys

def test(email: str, auth_code: str):
    print(f"Email: {email}")
    print(f"Server: imap.126.com:993 (SSL)")
    print("-" * 40)

    try:
        context = ssl.create_default_context()
        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        print("[OK] TCP+SSL connected")

        mail.login(email, auth_code)
        print("[OK] Login success")

        status, data = mail.select('INBOX')
        print(f"[OK] INBOX selected, emails: {data[0].decode()}")

        status, ids = mail.search(None, 'ALL')
        count = len(ids[0].split()) if ids[0] else 0
        print(f"[OK] Total emails: {count}")

        mail.logout()
        return True
    except imaplib.IMAP4.error as e:
        print(f"[FAIL] IMAP error: {e}")
        return False
    except Exception as e:
        print(f"[FAIL] {type(e).__name__}: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python test_imap.py <email> <auth_code>")
        sys.exit(1)
    test(sys.argv[1], sys.argv[2])
