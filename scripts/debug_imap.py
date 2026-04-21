"""Debug 126 IMAP - 详细诊断"""
import imaplib
import ssl
import socket

def debug_imap(email, auth_code):
    print(f"Testing: {email}")
    print("=" * 50)

    # Test 1: SSL on 993
    print("\n[1] Testing IMAP SSL (993)...")
    try:
        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE

        mail = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        print(f"    Banner: {mail.send_untagged_response}")

        typ, data = mail.login(email, auth_code)
        print(f"    Login response: {typ} {data}")
    except Exception as e:
        print(f"    Error: {type(e).__name__}: {e}")

    # Test 2: STARTTLS on 143
    print("\n[2] Testing IMAP STARTTLS (143)...")
    try:
        mail2 = imaplib.IMAP4('imap.126.com', 143)
        print(f"    Banner: {mail2.send_untagged_response}")
        typ, data = mail2.starttls(ssl.create_default_context())
        print(f"    STARTTLS: {typ} {data}")
        typ, data = mail2.login(email, auth_code)
        print(f"    Login response: {typ} {data}")
        mail2.logout()
    except Exception as e:
        print(f"    Error: {type(e).__name__}: {e}")

    # Test 3: Check if IMAP capabilities are accessible
    print("\n[3] Testing CAPABILITY command...")
    try:
        context = ssl.create_default_context()
        context.check_hostname = False
        context.verify_mode = ssl.CERT_NONE
        mail3 = imaplib.IMAP4_SSL('imap.126.com', 993, ssl_context=context)
        # Before login, check capabilities
        typ, data = mail3.capability()
        print(f"    CAPABILITY (pre-login): {data}")
        typ, data = mail3.login(email, auth_code)
        print(f"    Login: {typ} {data}")
        if typ == 'OK':
            typ, data = mail3.capability()
            print(f"    CAPABILITY (post-login): {data}")
            mail3.logout()
    except Exception as e:
        print(f"    Error: {type(e).__name__}: {e}")

if __name__ == "__main__":
    import sys
    if len(sys.argv) < 3:
        print("Usage: python debug_imap.py <email> <auth_code>")
        sys.exit(1)
    debug_imap(sys.argv[1], sys.argv[2])