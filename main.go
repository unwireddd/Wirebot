package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	//"reflect"

	"time"

	"github.com/anaskhan96/soup"
	"github.com/wader/goutubedl"

	tele "gopkg.in/telebot.v3"
)

func pingWebsite(url string) bool {
	parts := strings.Split(url, ":")
	domain := parts[0]
	port := "80"
	if len(parts) > 1 {
		port = parts[1]
	}
	conn, err := net.DialTimeout("tcp", domain+":"+port, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}

func main() {

	pref := tele.Settings{
		Token:  "",
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle("/get", func(c tele.Context) error {

		arguments := c.Args()
		if len(arguments) > 1 {
			return c.Send("You can only download one video at once")
		}

		if len(arguments) == 0 {
			return c.Send("The correct syntax for this command is: /dlp [link to video you want to download]")
		}

		//isWebsite := arguments[0]
		//u, err := url.Parse(isWebsite)
		//if err != nil {
		//	fmt.Println(u)
		//	return c.Send("The url you have just provided doesnt seem to be a valid website.")
		//} else {
		//	fmt.Println("Valid URL")
		//}

		result, err := goutubedl.New(context.Background(), arguments[0], goutubedl.Options{})
		if err != nil {
			return c.Send("Something went wrong, try again later")
		}
		c.Send("Your video is getting downloaded...")
		downloadResult, err := result.Download(context.Background(), "best")
		if err != nil {
			return c.Send(err)
		}
		defer downloadResult.Close()
		f, _ := os.Create("output")
		defer f.Close()
		io.Copy(f, downloadResult)
		v := &tele.Video{File: tele.FromDisk("./output"), Caption: "Your video"}
		defer os.Remove("output")
		return c.Send(v)

	})

	b.Handle("/remindme", func(c tele.Context) error {

		syntaxErr := "The correct syntax for this command is: [time you want me to send you a reminder from now], [time unit], [the thing you want me to remind you about] For example: /remindme 1 h watch a movie"
		arguments := c.Args()
		if len(arguments) < 2 {
			return c.Send(syntaxErr)
		}
		duration := arguments[0]
		timeUnit := arguments[1]
		task := arguments[2:]

		//test := "no"
		fixedDuration := time.Duration(1) * time.Minute
		var successMessage string

		//if ok, _ := regexp.MatchString(`^\d+$`, duration); ok {
		//	return c.Send(syntaxErr)
		//}

		if timeUnit != "m" && timeUnit != "h" {
			return c.Send(syntaxErr)
		}

		intDuration, _ := strconv.Atoi(duration)
		if intDuration < 0 {
			return c.Send("The minimum time you can set a reminder for is 1 minute.")
		}
		if timeUnit == "m" {
			fixedDuration = time.Duration(intDuration) * time.Minute
			successMessage = fmt.Sprintf("Ok, Im gonna remind you about your task: %v in about %v m.", task, duration)
		} else {
			fixedDuration = time.Duration(intDuration) * time.Hour
			successMessage = fmt.Sprintf("Ok, Im gonna remind you about your task: %v in about %v h.", task, duration)
		}

		returnMessage := fmt.Sprintf("Hey, %v minutes passed, I came back to remind you about the task you wanted me to remind you about: %v", duration, task)
		c.Send(successMessage)
		//c.Send(b.GetC)
		time.Sleep(fixedDuration)
		//c.Send(successMessage)
		return c.Send(returnMessage)

	})

	b.Handle("/ping", func(c tele.Context) error {

		arguments := c.Args()
		if len(arguments) > 1 {
			return c.Send("You can only ping one website at once")
		}

		if len(arguments) == 0 {
			return c.Send("The correct syntax for this command is: /ping [link to the webiste you want to ping]")
		}
		website := arguments[0]
		siteUp := fmt.Sprintf("%v seems to be up for now.", website)
		siteDown := fmt.Sprintf("%v seems to be down for now.", website)
		pingWebsite(website)
		if pingWebsite(website) {
			return c.Send(siteUp)
		} else {
			return c.Send(siteDown)

		}

	})

	b.Handle("/weather", func(c tele.Context) error {

		accu, _ := soup.Get("https://pogoda.onet.pl/prognoza-pogody/wroclaw-362450")
		currentDate := time.Now()
		weather := soup.HTMLParse(accu)
		owtput := weather.Find("div", "class", "temp").FullText()
		desc := weather.Find("div", "class", "forecastDesc").FullText()
		outString := fmt.Sprintf("The current weather in Wrocław for now (%s) is %s, %s", currentDate.Format("15:04, 02.01, 2006"), owtput, desc)
		return c.Send(outString)

	})

	b.Start()
}
