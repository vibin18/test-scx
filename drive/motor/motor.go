package robots

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gobot.io/x/gobot" // Add missing import statement
	"log"
	"math"
	"scx24/config"
	"strconv"
	"time"
)

var app *config.Config

func NewRoboConfig(c *config.Config) {
	app = c
}

var cal = [...]float32{0, 0, 0}

var RoboWork = func() {
	log.Printf("Steering servo to: %v", app.SteerAngle)
	err := app.ServoSteer.Start()
	if err != nil {
		log.Printf("%v", err)
	}
	err = app.ServoSteer.Center()
	if err != nil {

		log.Printf("%v", err)
	}

	log.Printf("Driving motor to : %v", app.DrivePWM)
	err = app.ServoDrive.ServoWrite(byte(app.DrivePWM))
	if err != nil {
		log.Printf("%v", err)
	}

	var orientationX = 0
	var orientationY = 0
	var orientationZ = 0

	status := app.I2CGyro.Connection()
	log.Printf("%v", status)

	// drift := .40

	// work is a function that continuously reads the XYZ values from a sensor and updates the orientationX, orientationY, and orientationZ variables.
	// It uses the gobot.Every function to execute the code block every 1 millisecond.
	// The XYZ values are obtained from the d sensor, and any error that occurs during the reading is logged.
	// The orientationX, orientationY, and orientationZ variables are updated by adding the XYZ values and subtracting the drift value.
	// Finally, the updated orientation values are printed to the console.
	calibrateGyro()
	fmt.Printf("Calibration drift X : %v , y : %v, z : %v\n", cal[0], cal[1], cal[2])
	//dxx, dyy, dzz := calibGyro(d)
	//fmt.Printf("Calibration drift X : %v , y : %v, z : %v", dxx, dyy, dzz)
	go func() {
		gobot.Every(1*time.Millisecond, func() {
			xx, yy, zz, err := app.I2CGyro.XYZ()
			if err != nil {
				log.Printf("%v", err)
			}
			if cal[0] == 0 {
				calibrateGyro()
			}

			xx = xx - cal[0]
			yy = yy - cal[1]
			zz = zz - cal[2]

			xx *= 0.1
			yy *= 0.1
			zz *= 0.1

			orientationX += int(xx)
			orientationY += int(yy)
			orientationZ += int(zz)

			//fmt.Printf("X = %v , Y = %v , Z = %v\n", orientationX/100, orientationY/100, orientationZ/100)
			a := orientationZ / 100
			config.SetGyro(app, fmt.Sprintf("%v", a))
			//fmt.Println(app.Gyro)

			//fmt.Printf("X = %v , Y = %v , Z = %v \n", xx, yy, zz)
		})
	}()

}

func calibrateGyro() {
	fmt.Printf("Calibration Gyro..\n")
	for i := 0; i < 1000; i++ {
		gx, gy, gz, _ := app.I2CGyro.XYZ()
		cal[0] += float32(gx) / 1000000
		cal[1] += float32(gy) / 1000000
		cal[2] += float32(gz) / 1000000
		time.Sleep(time.Millisecond * 10)
	}
	cal[0] /= 100
	cal[1] /= 100
	cal[2] /= 100
}

func InitRobots(m *gobot.Master) {
	c := app
	m.AddRobot(gobot.NewRobot("car",
		[]gobot.Connection{app.RaspiAdapter},
		[]gobot.Device{app.ServoSteer, app.ServoDrive, app.I2CGyro},
		RoboWork))

	m.AddCommand("custom_gobot_command",
		func(params map[string]interface{}) interface{} {
			return "This command is attached to the mcp!"
		})

	m.Robot("car").AddCommand("steer", func(m map[string]interface{}) interface{} {
		if u, ok := m["steer_move"].(string); ok {
			log.Printf("Received new update for steering move: %v", u)
			ui, _ := strconv.Atoi(u)
			err := c.ServoSteer.Move(uint8(ui))
			if err != nil {
				log.Printf("Moving servo err: %v", err)
			}
			c.SteerAngle = ui
		}
		return fmt.Sprintf("Update new steering angle: %v", c.SteerAngle)
	})
	m.Robot("car").AddCommand("drive", func(m map[string]interface{}) interface{} {
		if u, ok := m["drive_pwm"].(string); ok {
			log.Printf("Received new update for drive pwm: %v", u)
			u, _ := strconv.Atoi(u)
			err := c.ServoDrive.ServoWrite(byte(u))
			if err != nil {
				log.Printf("Driving pwm error %v", err)
			}
			c.DrivePWM = u
		}
		return fmt.Sprintf("Update new pwm: %v", c.DrivePWM)
	})

}

// write a function that uses a PID controller which drives the car 90 degree
func Drive90Degrees(ctx *fiber.Ctx) error {
	fmt.Printf("Executing PID..\n")
	// Initialize PID controller parameters
	Kp := 1.0
	Ki := 0.1
	Kd := 0.01

	// Set target angle to 90 degrees
	minAngle := 15.0
	maxAngle := 35.0
	targetAngle := 0.0

	// Initialize PID controller variables
	prevError := 0.0
	integral := 0.0

	for {
		fmt.Println("Correcting angle")
		// Get current orientation angle
		gy := app.Gyro
		fmt.Println("Getting current angle " + *gy)
		gyi, _ := strconv.Atoi(*gy)
		currentAngle := float64(gyi) // 100.0

		// Calculate error
		er := targetAngle - currentAngle

		// Calculate PID terms
		proportional := Kp * er
		integral += Ki * er
		derivative := Kd * (er - prevError)

		// Calculate PID output
		output := proportional + integral + derivative

		allowedAngle := math.Max(minAngle, math.Min(maxAngle, output))

		// Apply PID output to control the car's steering
		err := app.ServoSteer.Move(uint8(allowedAngle))
		if err != nil {
			fmt.Printf("Moving servo err: %v", err)
		}

		// Update previous er
		prevError = er

		fmt.Println("Current error: " + fmt.Sprintf("%v", er))

		// Check if the car has reached the target angle
		//if math.Abs(er) < 0.1 {
		//	fmt.Println("Correct angle...exiting PID")
		//	break
		//}
		//fmt.Println("Correcting angle again..")
		//// Sleep for a short duration before the next iteration
		//time.Sleep(10 * time.Millisecond)
	}
	return nil
}
