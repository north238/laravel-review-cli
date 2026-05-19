package main

import (
	"os"

	"github.com/north238/lrv/internal/cli"
)

func main() {
	// config, err := config.NewConfig()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// client, err := llm.NewAnthropicClient(config.APIKey)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	// req := llm.ReviewRequest{
	// 	SystemPrompt: "Today's date is 2026-05-19.",
	// 	UserPrompt:   "Hello, world",
	// 	Model:        config.Model,
	// 	MaxTokens:    256,
	// }

	// resp, err := client.Review(ctx, req)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// fmt.Println(resp)

	err := cli.NewRootCommand().Execute()
	if err != nil {
		os.Exit(1)
	}
}
