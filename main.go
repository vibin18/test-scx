package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/api"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
	"log"
	"scx24/config"
	"scx24/drive/motor"
	"scx24/handler"
)

const (
	Backend            = "http://localhost:3000"
	InitialDrivePWM    = 25
	InitialSteerAngle  = 90
	ForwardSlowSpeed   = 26
	ForwardFastSpeed   = 35
	ForwardMediumSpeed = 30
	ReverseSlowSpeed   = 24
	ReverseFastSpeed   = 20
	ReverseMediumSpeed = 15
	LeftShortAngle     = 30
	LeftLongAngle      = 35
	RightShortAngle    = 22
	StraightAngle      = 26
	RightLongAngle     = 15
)

func main() {

	//firmataAdaptor1 := firmata.NewTCPAdaptor("192.168.1.128:3030")
	raspi := raspi.NewAdaptor()

	PWM := gpio.NewDirectPinDriver(raspi, "33")
	Servo := gpio.NewServoDriver(raspi, "32")
	I2CGyro := i2c.NewL3GD20HDriver(raspi,
		i2c.WithBus(17),
		i2c.WithAddress(0x6b))
	mbot := gobot.NewMaster()

	myConfig := &config.Config{
		DrivePWM:          InitialDrivePWM,
		SteerAngle:        InitialSteerAngle,
		ServoDrive:        PWM,
		ServoSteer:        Servo,
		RaspiAdapter:      raspi,
		BotMaster:         mbot,
		I2CGyro:           I2CGyro,
		ComChannel:        make(chan string),
		Gyro:              new(string),
		InitialDrivePWM:   InitialDrivePWM,
		InitialSteerAngle: InitialSteerAngle,
		Frontend: config.Frontend{
			BackendURL: Backend,
		},
		Drive: config.Drive{
			ForwardSlowSpeed:   ForwardSlowSpeed,
			ForwardFastSpeed:   ForwardFastSpeed,
			ForwardMediumSpeed: ForwardMediumSpeed,
			ReverseSlowSpeed:   ReverseSlowSpeed,
			ReverseFastSpeed:   ReverseFastSpeed,
			ReverseMediumSpeed: ReverseMediumSpeed,
		},
		Steer: config.Steer{
			LeftShortAngle:  LeftShortAngle,
			LeftLongAngle:   LeftLongAngle,
			RightShortAngle: RightShortAngle,
			RightLongAngle:  RightLongAngle,
			StraightAngle:   StraightAngle,
		},
	}

	cd := config.NewConfig(myConfig)
	handler.NewHandlerConfig(cd)
	robots.NewRoboConfig(cd)
	robots.InitRobots(mbot)

	a := api.NewAPI(mbot)
	// a.Debug()
	a.AddC3PIORoutes()
	a.StartWithoutDefaults()

	//// enable template rendering on fiber
	engine := html.New("html", ".html")
	engine.Reload(true)
	engine.Debug(true)

	web := fiber.New(fiber.Config{Views: engine})
	web.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	web.Get("/ws", websocket.New(handler.WsGyro))
	//web.Get("/ws", websocket.New(func(c *websocket.Conn) {
	//
	//	for {
	//		select {
	//		case Chan := <-cd.ComChannel:
	//			//gy := config.GetGyro(cd)
	//			htmlMsg := fmt.Sprintf("<span id=\"message\">Gyro val: %v </span>", Chan)
	//			msg := []byte(htmlMsg)
	//			if err := c.WriteMessage(1, msg); err != nil {
	//				log.Println("write:", err)
	//				break
	//			}
	//			//time.Sleep(1 * time.Second)
	//		}
	//	}
	//
	//}))
	web.Get("/", handler.Home)
	web.Post("/stop_button", handler.StopButton)
	web.Post("/updatePWM", handler.UpdateDriverSlider)
	web.Post("/forward", handler.DriveForwardSlow)
	web.Post("/reverse", handler.DriveReverseSlow)
	web.Post("/right", handler.TurnRightLong)
	web.Post("/left", handler.TurnLeftLong)
	//web.Post("/straight", handler.TurnStraight)
	web.Post("/straight", robots.Drive90Degrees)

	go func() { log.Fatal(web.Listen(":8080")) }()
	err := mbot.Start()
	if err != nil {
		log.Printf("%v", err)
	}
}
