package config

import (
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

type Config struct {
	DrivePWM          int
	ServoDrive        *gpio.DirectPinDriver
	I2CGyro           *i2c.L3GD20HDriver
	PwmPercentage     float64
	ServoSteer        *gpio.ServoDriver
	SteerAngle        int
	RaspiAdapter      *raspi.Adaptor
	BotMaster         *gobot.Master
	InitialDrivePWM   int
	InitialSteerAngle int
	Gyro              *string
	ComChannel        chan string
	Frontend
	Drive
	Steer
}

type Frontend struct {
	BackendURL string
}

type Drive struct {
	ForwardSlowSpeed   int
	ForwardFastSpeed   int
	ForwardMediumSpeed int
	ReverseSlowSpeed   int
	ReverseFastSpeed   int
	ReverseMediumSpeed int
}

type Steer struct {
	LeftShortAngle  int
	LeftLongAngle   int
	RightShortAngle int
	RightLongAngle  int
	StraightAngle   int
}

func NewConfig(c *Config) *Config {
	return c
}

func SetGyro(c *Config, gyro string) {
	c.Gyro = &gyro
	go func() { c.ComChannel <- gyro }()
}

func GetGyro(c *Config) string {
	return *c.Gyro
}
