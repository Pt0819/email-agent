"""
集成测试脚本 - 验证Agent服务和LLM连接
运行方式: python scripts/test_integration.py
"""
import sys
import asyncio
from pathlib import Path

# Windows控制台编码处理
if sys.platform == 'win32':
    sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    sys.stderr.reconfigure(encoding='utf-8', errors='replace')

# 添加项目根目录到路径
sys.path.insert(0, str(Path(__file__).parent.parent / "email-agent"))

import httpx
from app.llm import get_llm_manager
from app.core import get_config


async def test_llm_connection():
    """测试LLM连接"""
    print("\n" + "="*50)
    print("测试1: LLM连接")
    print("="*50)

    manager = get_llm_manager()
    providers = manager.get_available_providers()
    print(f"可用Provider: {providers}")

    if not providers:
        print("[FAIL] 没有可用的LLM Provider")
        return False

    # 测试智谱GLM
    if "zhipu" in providers:
        print("\n测试智谱GLM连接...")
        is_valid = await manager.validate_provider("zhipu")
        if is_valid:
            print("[PASS] 智谱GLM连接成功")
        else:
            print("[FAIL] 智谱GLM连接失败")
            return False
    else:
        print("[FAIL] 智谱GLM Provider未初始化")
        return False

    return True


async def test_llm_classification():
    """测试LLM分类功能"""
    print("\n" + "="*50)
    print("测试2: 邮件分类")
    print("="*50)

    manager = get_llm_manager()

    test_email = {
        "subject": "【紧急】项目进度汇报 - 需要今天下班前回复",
        "sender_name": "张经理",
        "sender_email": "zhangmanager@company.com",
        "content": "各位同事好，请在今天下班前提交本周项目进度汇报，谢谢配合。",
        "received_at": "2026-04-21 10:00:00"
    }

    system_prompt = """你是一个邮件分类助手。请分析邮件并返回JSON格式结果：
{"category": "分类", "priority": "优先级", "confidence": 0.8, "reasoning": "判断理由"}

分类选项: work_urgent, work_normal, personal, subscription, notification, promotion, spam
优先级: critical, high, medium, low"""

    user_prompt = f"""分析以下邮件：
发件人: {test_email['sender_name']} <{test_email['sender_email']}>
主题: {test_email['subject']}
内容: {test_email['content']}
时间: {test_email['received_at']}"""

    print(f"测试邮件: {test_email['subject']}")

    try:
        response = await manager.chat_with_system(
            system_prompt=system_prompt,
            user_content=user_prompt
        )
        print(f"\nLLM响应:\n{response.content}")
        print(f"\n[PASS] 分类测试成功")
        return True
    except Exception as e:
        print(f"\n[FAIL] 分类测试失败: {e}")
        return False


async def test_agent_health():
    """测试Agent服务健康检查"""
    print("\n" + "="*50)
    print("测试3: Agent服务健康检查")
    print("="*50)

    config = get_config()
    url = f"http://localhost:{config.server.port}/api/v1/health"

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            response = await client.get(url)
            if response.status_code == 200:
                data = response.json()
                print(f"服务状态: {data.get('status')}")
                print(f"可用Provider: {data.get('providers')}")
                print(f"LLM状态: {data.get('llm_status')}")
                print("[PASS] Agent服务运行正常")
                return True
            else:
                print(f"[FAIL] Agent服务响应异常: {response.status_code}")
                return False
    except httpx.ConnectError:
        print(f"[FAIL] 无法连接Agent服务 ({url})")
        print("   请先启动Agent服务: cd email-agent && python app/main.py")
        return False
    except Exception as e:
        print(f"[FAIL] 健康检查失败: {e}")
        return False


async def test_agent_classify():
    """测试Agent分类API"""
    print("\n" + "="*50)
    print("测试4: Agent分类API")
    print("="*50)

    config = get_config()
    url = f"http://localhost:{config.server.port}/api/v1/classify"

    request_body = {
        "email_id": "test-001",
        "subject": "【重要】明天上午10点会议通知",
        "sender_name": "行政部",
        "sender_email": "admin@company.com",
        "content": "各位同事，明天上午10点在三楼会议室召开季度总结会议，请准时参加。",
        "received_at": "2026-04-21 09:00:00"
    }

    print(f"测试邮件: {request_body['subject']}")

    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(url, json=request_body)
            if response.status_code == 200:
                data = response.json()
                classification = data.get("classification", {})
                print(f"\n分类结果:")
                print(f"  类别: {classification.get('category')}")
                print(f"  优先级: {classification.get('priority')}")
                print(f"  置信度: {classification.get('confidence')}")
                print(f"  理由: {classification.get('reasoning')}")
                print("\n[PASS] Agent分类API测试成功")
                return True
            else:
                print(f"[FAIL] API响应异常: {response.status_code}")
                print(f"   响应: {response.text}")
                return False
    except httpx.ConnectError:
        print(f"[FAIL] 无法连接Agent服务")
        return False
    except Exception as e:
        print(f"[FAIL] 分类测试失败: {e}")
        return False


async def main():
    """运行所有测试"""
    print("\n" + "="*60)
    print(" Email Agent 集成测试")
    print("="*60)

    results = []

    # 测试1: LLM连接
    results.append(("LLM连接", await test_llm_connection()))

    # 测试2: LLM分类
    results.append(("LLM分类", await test_llm_classification()))

    # 测试3: Agent健康检查
    results.append(("Agent健康检查", await test_agent_health()))

    # 测试4: Agent分类API (需要Agent服务运行)
    results.append(("Agent分类API", await test_agent_classify()))

    # 汇总结果
    print("\n" + "="*60)
    print(" 测试结果汇总")
    print("="*60)

    passed = sum(1 for _, r in results if r)
    total = len(results)

    for name, result in results:
        status = "[PASS]" if result else "[FAIL]"
        print(f"  {name}: {status}")

    print(f"\n总计: {passed}/{total} 通过")

    if passed == total:
        print("\n[SUCCESS] 所有测试通过！系统准备就绪。")
    else:
        print("\n[WARNING] 部分测试未通过，请检查配置。")


if __name__ == "__main__":
    asyncio.run(main())
