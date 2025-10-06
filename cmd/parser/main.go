package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	// создаём контекст для браузера
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// переменная для HTML страницы
	var html string

	// открываем сайт и ждём загрузки
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://ticketon.kz/`),
		chromedp.Sleep(5*time.Second), // ждём 5 сек, чтобы всё подгрузилось
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Длина HTML:", len(html))

	// пример — достанем названия мероприятий
	var titles []string
	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://ticketon.kz/`),
		chromedp.Sleep(5*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll(".event-card__title"))
			     .map(el => el.innerText)
		`, &titles),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n🎫 События Ticketon:")
	for _, title := range titles {
		fmt.Println("-", title)
	}
}
