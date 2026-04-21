// +build ignore

// 126邮箱Provider测试程序
// 使用方法: go run test_126_provider.go
//
// 需要设置环境变量:
//   EMAIL_126_ADDRESS=your_email@126.com
//   EMAIL_126_CREDENTIAL=your_authorization_code
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"email-backend/server/pkg/email/provider"
)

func main() {
	fmt.Println("============================================")
	fmt.Println(" 126邮箱Provider测试")
	fmt.Println("============================================")

	// 从环境变量获取配置
	email := os.Getenv("EMAIL_126_ADDRESS")
	credential := os.Getenv("EMAIL_126_CREDENTIAL")

	if email == "" || credential == "" {
		fmt.Println("\n请设置环境变量:")
		fmt.Println("  EMAIL_126_ADDRESS=your_email@126.com")
		fmt.Println("  EMAIL_126_CREDENTIAL=your_authorization_code")
		fmt.Println("\n示例:")
		fmt.Println("  EMAIL_126_ADDRESS=test@126.com EMAIL_126_CREDENTIAL=ABCD1234EFGH5678 go run test_126_provider.go")
		return
	}

	// 创建Provider
	p := provider.NewNet126Provider(nil).(*provider.Net126Provider)
	fmt.Printf("\nProvider: %s\n", p.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试连接
	fmt.Printf("\n[测试1] 连接邮箱服务器 %s@126.com...\n", email[:len(email)-10]+"***")
	err := p.Connect(ctx, email, credential)
	if err != nil {
		fmt.Printf("[失败] 连接失败: %v\n", err)
		return
	}
	fmt.Println("[成功] 连接成功")
	defer p.Disconnect()

	// 测试连接状态
	fmt.Println("\n[测试2] 测试连接状态...")
	result, err := p.TestConnection(ctx)
	if err != nil {
		fmt.Printf("[失败] 测试连接失败: %v\n", err)
		return
	}
	fmt.Printf("[成功] 连接状态: %s\n", result.Message)

	// 获取邮件列表
	fmt.Println("\n[测试3] 获取最近7天的邮件列表...")
	since := time.Now().AddDate(0, 0, -7)
	summaries, err := p.FetchEmailList(ctx, since, 10)
	if err != nil {
		fmt.Printf("[失败] 获取邮件列表失败: %v\n", err)
		return
	}

	fmt.Printf("[成功] 获取到 %d 封邮件\n", len(summaries))
	for i, s := range summaries {
		if i >= 5 {
			fmt.Printf("  ... 还有 %d 封邮件\n", len(summaries)-5)
			break
		}
		fmt.Printf("  %d. [%s] %s - %s\n", i+1, s.ReceivedAt.Format("01-02 15:04"), s.SenderEmail, truncate(s.Subject, 30))
	}

	if len(summaries) > 0 {
		// 获取第一封邮件详情
		fmt.Println("\n[测试4] 获取邮件详情...")
		firstEmail := summaries[0]
		detail, err := p.FetchEmailDetail(ctx, firstEmail.MessageID)
		if err != nil {
			fmt.Printf("[失败] 获取详情失败: %v\n", err)
			return
		}
		fmt.Printf("[成功] 邮件主题: %s\n", detail.Subject)
		fmt.Printf("       发件人: %s <%s>\n", detail.SenderName, detail.SenderEmail)
		fmt.Printf("       内容长度: %d 字符\n", len(detail.Content))
	}

	fmt.Println("\n============================================")
	fmt.Println(" 所有测试通过!")
	fmt.Println("============================================")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
