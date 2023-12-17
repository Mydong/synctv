package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/synctv-org/synctv/server/handlers/vendors"
	"github.com/synctv-org/synctv/server/handlers/vendors/vendorAlist"
	"github.com/synctv-org/synctv/server/handlers/vendors/vendorBilibili"
	"github.com/synctv-org/synctv/server/middlewares"
	"github.com/synctv-org/synctv/utils"
)

func Init(e *gin.Engine) {
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
				admin.GET("/settings", AdminSettings)

				admin.GET("/settings/:group", AdminSettings)

				admin.POST("/settings", EditAdminSettings)

				admin.GET("/vendors", AdminGetVendorBackends)

				admin.POST("/vendors/add", AdminAddVendorBackend)

				admin.POST("/vendors/update", AdminUpdateVendorBackends)

				admin.POST("/vendors/delete", AdminDeleteVendorBackends)

				admin.POST("/vendors/reconnect", AdminReconnectVendorBackends)

				{
					user := admin.Group("/user")

					user.POST("/add", AddUser)

					user.POST("/delete", DeleteUser)

					user.POST("/password", AdminUserPassword)

					user.POST("/username", AdminUsername)

					// 查找用户
					user.GET("/list", Users)

					user.POST("/approve", ApprovePendingUser)

					user.POST("/ban", BanUser)

					user.POST("/unban", UnBanUser)

					// 查找某个用户的房间
					user.GET("/rooms", GetUserRooms)
				}

				{
					room := admin.Group("/room")

					room.POST("/password", AdminRoomPassword)

					// 查找房间
					room.GET("/list", Rooms)

					room.POST("/approve", ApprovePendingRoom)

					room.POST("/ban", BanRoom)

					room.POST("/unban", UnBanRoom)

					room.GET("/users", GetRoomUsers)
				}
			}

			{
				root.POST("/admin/add", AddAdmin)

				root.POST("/admin/delete", DeleteAdmin)
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

			needAuthRoom.GET("/settings", RoomSetting)

			needAuthRoom.POST("/settings", SetRoomSetting)

			needAuthRoom.GET("/users", RoomUsers)
		}

		{
			movie := api.Group("/movie")
			needAuthMovie := needAuthRoomApi.Group("/movie")

			needAuthMovie.GET("/list", MovieList)

			needAuthMovie.GET("/current", CurrentMovie)

			needAuthMovie.GET("/movies", Movies)

			needAuthMovie.POST("/current", ChangeCurrentMovie)

			needAuthMovie.POST("/push", PushMovie)

			needAuthMovie.POST("/pushs", PushMovies)

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
			user := api.Group("/user")
			needAuthUser := needAuthUserApi.Group("/user")

			user.POST("/login", LoginUser)

			needAuthUser.POST("/logout", LogoutUser)

			needAuthUser.GET("/me", Me)

			needAuthUser.GET("/rooms", UserRooms)

			needAuthUser.POST("/username", SetUsername)

			needAuthUser.POST("/password", SetUserPassword)

			needAuthUser.GET("/providers", UserBindProviders)
		}

		{
			vendor := needAuthUserApi.Group("/vendor")

			vendor.GET("/backends/:vendor", vendors.Backends)

			{
				bilibili := vendor.Group("/bilibili")

				login := bilibili.Group("/login")

				login.GET("/qr", vendorBilibili.NewQRCode)

				login.POST("/qr", vendorBilibili.LoginWithQR)

				login.GET("/captcha", vendorBilibili.NewCaptcha)

				login.POST("/sms/send", vendorBilibili.NewSMS)

				login.POST("/sms/login", vendorBilibili.LoginWithSMS)

				bilibili.POST("/parse", vendorBilibili.Parse)

				bilibili.GET("/me", vendorBilibili.Me)

				bilibili.POST("/logout", vendorBilibili.Logout)
			}

			{
				alist := vendor.Group("/alist")

				alist.POST("/login", vendorAlist.Login)

				alist.POST("/logout", vendorAlist.Logout)

				alist.POST("/list", vendorAlist.List)

				alist.GET("/me", vendorAlist.Me)
			}
		}
	}
}
