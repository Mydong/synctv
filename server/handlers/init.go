package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/synctv-org/synctv/public"
	Vbilibili "github.com/synctv-org/synctv/server/handlers/vendors/bilibili"
	"github.com/synctv-org/synctv/server/middlewares"
	"github.com/synctv-org/synctv/utils"
)

func Init(e *gin.Engine) {
	{
		e.GET("/", func(ctx *gin.Context) {
			ctx.Redirect(http.StatusMovedPermanently, "/web/")
		})

		web := e.Group("/web")

		web.Use(middlewares.NewDistCacheControl("/web/"))

		web.StaticFS("", http.FS(public.Public))
	}

	{
		api := e.Group("/api")

		needAuthUserApi := api.Group("")
		needAuthUserApi.Use(middlewares.AuthUserMiddleware)

		needAuthRoomApi := api.Group("")
		needAuthRoomApi.Use(middlewares.AuthRoomMiddleware)

		{
			public := api.Group("/public")

			public.GET("/settings", Settings)
		}

		{
			admin := api.Group("/admin")
			root := api.Group("/admin")
			admin.Use(middlewares.AuthAdminMiddleware)
			root.Use(middlewares.AuthRootMiddleware)

			{
				admin.GET("/settings/:group", AdminSettings)

				admin.POST("/settings", EditAdminSettings)

				admin.GET("/users", Users)
			}

			{
				root.GET("/admins", Admins)

				root.POST("/addAdmin", AddAdmin)

				root.POST("deleteAdmin", DeleteAdmin)
			}
		}

		{
			room := api.Group("/room")
			needAuthRoom := needAuthRoomApi.Group("/room")
			needAuthUser := needAuthUserApi.Group("/room")

			room.GET("/ws", NewWebSocketHandler(utils.NewWebSocketServer()))

			room.GET("/check", CheckRoom)

			room.GET("/hot", RoomHotList)

			room.GET("/list", RoomList)

			needAuthUser.POST("/create", CreateRoom)

			needAuthUser.POST("/login", LoginRoom)

			needAuthRoom.POST("/delete", DeleteRoom)

			needAuthRoom.POST("/pwd", SetRoomPassword)

			needAuthRoom.GET("/setting", RoomSetting)
		}

		{
			movie := api.Group("/movie")
			needAuthMovie := needAuthRoomApi.Group("/movie")

			needAuthMovie.GET("/list", MovieList)

			needAuthMovie.GET("/current", CurrentMovie)

			needAuthMovie.GET("/movies", Movies)

			needAuthMovie.POST("/current", ChangeCurrentMovie)

			needAuthMovie.POST("/push", PushMovie)

			needAuthMovie.POST("/edit", EditMovie)

			needAuthMovie.POST("/swap", SwapMovie)

			needAuthMovie.POST("/delete", DelMovie)

			needAuthMovie.POST("/clear", ClearMovies)

			movie.HEAD("/proxy/:roomId/:movieId", ProxyMovie)

			movie.GET("/proxy/:roomId/:movieId", ProxyMovie)

			{
				live := needAuthMovie.Group("/live")

				live.POST("/publishKey", NewPublishKey)

				live.GET("/*movieId", JoinLive)
			}
		}

		{
			// user := api.Group("/user")
			needAuthUser := needAuthUserApi.Group("/user")

			needAuthUser.POST("/logout", LogoutUser)

			needAuthUser.GET("/me", Me)

			needAuthUser.GET("/rooms", UserRooms)

			needAuthUser.POST("/username", SetUsername)
		}

		{
			vendor := needAuthUserApi.Group("/vendor")

			{
				bilibili := vendor.Group("/bilibili")

				bilibili.GET("/qr", Vbilibili.QRCode)

				bilibili.POST("/login", Vbilibili.Login)

				bilibili.POST("/parse", Vbilibili.Parse)
			}
		}
	}
}
