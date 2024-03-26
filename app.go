package kuu

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strings"
)

type Map = map[string]any

type App struct {
	instanceID string
	raw        *fiber.App
	cache      Cache
}

type Config struct {
	fiber.Config
	Cache Cache
}

func New(conf ...Config) *App {
	var config Config
	if len(conf) > 0 {
		config = conf[0]
	}
	if config.Config.BodyLimit == 0 {
		// set default limit to	50MB
		config.Config.BodyLimit = 50 * 1024 * 1024
	}
	fa := fiber.New(config.Config)
	app := &App{
		instanceID: NewUUID(true, false),
		raw:        fa,
	}
	app.staticDefault()
	fa.Use(app.createModContext())
	fa.Use(logger.New())
	fa.Use(cors.New())
	fa.Use(recover.New())
	fa.Use(compress.New())
	//fa.Use(cache.New(cache.Config{
	//	CacheControl: true,
	//	Expiration:   10 * time.Second,
	//}))
	if config.Cache != nil {
		app.cache = config.Cache
	}
	return app
}

func (a *App) createModContext() fiber.Handler {
	return func(fc *fiber.Ctx) error {
		c := &Context{
			raw:       fc,
			app:       a,
			requestId: NewUUIDToken(),
		}
		fc.Context().SetUserValue(ContextValueKey, c)
		return fc.Next()
	}
}

func (a *App) Cache() Cache {
	return a.cache
}

func (a *App) InstanceID() string {
	return a.instanceID
}

func (a *App) staticDefault() {
	s := strings.TrimSpace(os.Getenv(ConfigStatic))
	if s != "" {
		for _, item := range strings.Split(s, ",") {
			v := strings.Split(item, ":")
			if len(v) == 2 {
				a.raw.Static(v[0], v[1])
			}
		}
	}
}

func (a *App) Static(prefix, root string, config ...fiber.Static) *App {
	a.raw.Static(prefix, root, config...)
	return a
}

func Use(a *App, handler Handler) {
	a.raw.Use(func(fc *fiber.Ctx) error {
		c := fc.Context().UserValue(ContextValueKey).(*Context)
		reply := handler(c)
		if reply != nil {
			if v, ok := reply.Data.(error); ok {
				Errorln(v)
				reply.Data = nil
			}
			return fc.JSON(reply)
		}
		return nil
	})
}

func (a *App) Run(addr ...string) {
	address := ":8080"
	if len(addr) > 0 {
		address = addr[0]
	}
	log.Fatal(a.raw.Listen(address))
}
