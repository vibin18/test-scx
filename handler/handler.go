package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
	"net/http"
	"scx24/config"
	"strconv"
	"time"
)

var app *config.Config

func NewHandlerConfig(c *config.Config) {
	app = c
}
func Home(ctx *fiber.Ctx) error {
	app.PwmPercentage = calculatePercentage(app.DrivePWM)
	g := config.GetGyro(app)

	type Data struct {
		DrivePWM      int
		SteerAngle    int
		PwmPercentage float64
		Gyro          string
	}
	nd := Data{
		DrivePWM:      app.DrivePWM,
		SteerAngle:    app.SteerAngle,
		PwmPercentage: app.PwmPercentage,
		Gyro:          g,
	}
	go func() {
		for {
			app.ComChannel <- "test"
			time.Sleep(1 * time.Second)
		}

	}()
	fmt.Println(nd.Gyro)
	return ctx.Render("index", nd)
}

func WsGyro(conn *websocket.Conn) {
	for {
		msg := <-app.ComChannel
		htmlMsg := fmt.Sprintf("<span id=\"message\">Gyro val: %v </span>", msg)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(htmlMsg)); err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func StopButton(ctx *fiber.Ctx) error {
	log.Printf("Stopping motor with %v", app.InitialDrivePWM)
	b := make(map[string]string)
	spwm := fmt.Sprintf("%v", app.InitialDrivePWM)
	b["drive_pwm"] = spwm
	bm, _ := json.Marshal(b)
	r, err := http.NewRequest("POST", app.BackendURL+"/api/robots/car/commands/drive", bytes.NewBuffer(bm))
	if err != nil {
		log.Printf("%v", err)
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Host", "localhost")

	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		log.Printf("%v", err)
	}
	app.DrivePWM = app.InitialDrivePWM
	app.PwmPercentage = calculatePercentage(app.DrivePWM)
	return ctx.Render("index", app)
}

func UpdateDriverSlider(ctx *fiber.Ctx) error {
	l := ctx.FormValue("slider")

	if l != "" {
		log.Printf("Slider value recived : %v", l)
		app.DrivePWM, _ = strconv.Atoi(l)
		app.PwmPercentage = calculatePercentage(app.DrivePWM)
		b := make(map[string]string)
		b["drive_pwm"] = l

		log.Printf("preparing the paylod: %v", b["drive_pwm"])
		bm, _ := json.Marshal(b)
		r, err := http.NewRequest("POST", app.Frontend.BackendURL+"/api/robots/car/commands/drive", bytes.NewBuffer(bm))
		if err != nil {
			log.Printf("%v", err)
		}
		r.Header.Add("Content-Type", "application/json")
		r.Header.Add("Host", "localhost")

		client := &http.Client{}
		_, err = client.Do(r)
		if err != nil {
			log.Printf("%v", err)
		}

	} else {
		log.Printf("empty slider value")
	}
	return ctx.Render("index", app)
}

func UpdateServo(ctx *fiber.Ctx) error {
	pwm := ctx.FormValue("servo_move")

	if pwm != "" {
		log.Printf("Servo value recived : %v", pwm)
		err := updateSteeringMove(pwm)
		if err != nil {
			log.Printf("%v", err)
		}

	} else {
		log.Printf("empty slider value")
	}
	return ctx.Render("index", app)
}

func DriveForwardSlow(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Drive.ForwardSlowSpeed)
	err := updateDrivePWM(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func DriveForwardMed(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Drive.ForwardMediumSpeed)
	err := updateDrivePWM(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func DriveReverseMed(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Drive.ReverseMediumSpeed)
	spwm := fmt.Sprintf("%v", app.InitialDrivePWM)
	err := updateDrivePWM(spwm)
	if err != nil {
		log.Printf("%v", err)
	}
	time.Sleep(3 * time.Second)
	err = updateDrivePWM(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func TurnStraight(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Steer.StraightAngle)
	err := updateSteeringMove(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func DriveReverseSlow(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Drive.ReverseSlowSpeed)
	spwm := fmt.Sprintf("%v", app.InitialDrivePWM)
	err := updateDrivePWM(spwm)
	if err != nil {
		log.Printf("%v", err)
	}
	time.Sleep(200 * time.Millisecond)

	err = updateDrivePWM(pwm)
	if err != nil {
		log.Printf("%v", err)
	}

	err = updateDrivePWM(spwm)
	if err != nil {
		log.Printf("%v", err)
	}
	time.Sleep(100 * time.Millisecond)

	err = updateDrivePWM(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func TurnRightLong(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Steer.RightLongAngle)
	err := updateSteeringMove(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}
func TurnLeftLong(ctx *fiber.Ctx) error {
	pwm := fmt.Sprintf("%v", app.Steer.LeftLongAngle)
	err := updateSteeringMove(pwm)
	if err != nil {
		log.Printf("%v", err)
	}
	return ctx.Render("index", app)
}

func updateDrivePWM(angle string) error {
	app.DrivePWM, _ = strconv.Atoi(angle)
	app.PwmPercentage = calculatePercentage(app.DrivePWM)
	b := make(map[string]string)
	b["drive_pwm"] = angle

	log.Printf("preparing the paylod: %v", b["drive_pwm"])
	bm, _ := json.Marshal(b)
	r, err := http.NewRequest("POST", app.Frontend.BackendURL+"/api/robots/car/commands/drive", bytes.NewBuffer(bm))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Host", "localhost")

	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		return err
	}
	return nil
}

func updateSteeringMove(pwm string) error {
	app.SteerAngle, _ = strconv.Atoi(pwm)

	b := make(map[string]string)
	b["steer_move"] = pwm

	log.Printf("preparing the paylod: %v", b["steer_move"])
	bm, _ := json.Marshal(b)
	r, err := http.NewRequest("POST", app.Frontend.BackendURL+"/api/robots/car/commands/steer", bytes.NewBuffer(bm))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Host", "localhost")

	client := &http.Client{}
	_, err = client.Do(r)
	if err != nil {
		return err
	}
	return nil
}

func calculatePercentage(value int) float64 {
	return float64(value) / float64(20) * 35
}
