package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/synctv-org/synctv/internal/conf"
	"github.com/synctv-org/synctv/internal/db"
	dbModel "github.com/synctv-org/synctv/internal/model"
	"github.com/synctv-org/synctv/internal/op"
	"github.com/synctv-org/synctv/internal/rtmp"
	"github.com/synctv-org/synctv/internal/vendor"
	pb "github.com/synctv-org/synctv/proto/message"
	"github.com/synctv-org/synctv/server/model"
	"github.com/synctv-org/synctv/utils"
	"github.com/synctv-org/vendors/api/bilibili"
	"github.com/zencoder/go-dash/v3/mpd"
	"github.com/zijiren233/gencontainer/refreshcache"
	"github.com/zijiren233/livelib/protocol/hls"
	"github.com/zijiren233/livelib/protocol/httpflv"
)

func GetPageAndPageSize(ctx *gin.Context) (int, int, error) {
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("max", "10"))
	if err != nil {
		return 0, 0, errors.New("max must be a number")
	}
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil {
		return 0, 0, errors.New("page must be a number")
	}
	return page, pageSize, nil
}

func GetPageItems[T any](ctx *gin.Context, items []T) ([]T, error) {
	page, max, err := GetPageAndPageSize(ctx)
	if err != nil {
		return nil, err
	}

	return utils.GetPageItems(items, page, max), nil
}

func MovieList(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	page, max, err := GetPageAndPageSize(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	m := room.GetMoviesWithPage(page, max)

	current, err := genCurrent(ctx, room.Current(), user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	mresp := make([]model.MoviesResp, len(m))
	for i, v := range m {
		mresp[i] = model.MoviesResp{
			Id:      v.ID,
			Base:    v.Base,
			Creator: op.GetUserName(v.CreatorID),
		}
		// hide headers when proxy
		if mresp[i].Base.Proxy {
			mresp[i].Base.Headers = nil
		}
	}

	current.UpdateSeek()

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"current": genCurrentResp(current),
		"total":   room.GetMoviesCount(),
		"movies":  mresp,
	}))
}

func genCurrent(ctx context.Context, current *op.Current, userID string) (*op.Current, error) {
	if current.Movie.Base.VendorInfo.Vendor != "" {
		return current, parse2VendorMovie(ctx, userID, &current.Movie)
	}
	return current, nil
}

func genCurrentResp(current *op.Current) *model.CurrentMovieResp {
	c := &model.CurrentMovieResp{
		Status: current.Status,
		Movie: model.MoviesResp{
			Id:      current.Movie.ID,
			Base:    current.Movie.Base,
			Creator: op.GetUserName(current.Movie.CreatorID),
		},
	}
	// hide headers when proxy
	if c.Movie.Base.Proxy {
		c.Movie.Base.Headers = nil
	}
	return c
}

func CurrentMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	current, err := genCurrent(ctx, room.Current(), user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	current.UpdateSeek()

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"current": genCurrentResp(current),
	}))
}

func Movies(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	// user := ctx.MustGet("user").(*op.User)

	page, max, err := GetPageAndPageSize(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	m := room.GetMoviesWithPage(int(page), int(max))

	mresp := make([]*model.MoviesResp, len(m))
	for i, v := range m {
		mresp[i] = &model.MoviesResp{
			Id:      v.ID,
			Base:    v.Base,
			Creator: op.GetUserName(v.CreatorID),
		}
		// hide headers when proxy
		if mresp[i].Base.Proxy {
			mresp[i].Base.Headers = nil
		}
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"total":  room.GetMoviesCount(),
		"movies": mresp,
	}))
}

func PushMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.PushMovieReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	err := user.AddMovieToRoom(room, (*dbModel.BaseMovie)(&req))
	if err != nil {
		if errors.Is(err, dbModel.ErrNoPermission) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorResp(err))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func PushMovies(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.PushMoviesReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	var ms []*dbModel.BaseMovie = make([]*dbModel.BaseMovie, len(req))

	for i, v := range req {
		m := (*dbModel.BaseMovie)(v)
		ms[i] = m
	}

	err := user.AddMoviesToRoom(room, ms)
	if err != nil {
		if errors.Is(err, dbModel.ErrNoPermission) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, model.NewApiErrorResp(err))
			return
		}
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func NewPublishKey(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.IdReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}
	movie, err := room.GetMovieByID(req.Id)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if movie.CreatorID != user.ID && !user.HasRoomPermission(room, dbModel.PermissionEditUser) {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	if !movie.Base.RtmpSource {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("only live movie can get publish key"))
		return
	}

	token, err := rtmp.NewRtmpAuthorization(movie.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	host := conf.Conf.Server.Rtmp.CustomPublishHost
	if host == "" {
		host = ctx.Request.Host
	}

	ctx.JSON(http.StatusOK, model.NewApiDataResp(gin.H{
		"host":  host,
		"app":   room.ID,
		"token": token,
	}))
}

func EditMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.EditMovieReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := user.UpdateMovie(room, req.Id, dbModel.BaseMovie(req.PushMovieReq)); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func DelMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.IdsReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	err := user.DeleteMoviesByID(room, req.Ids)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func ClearMovies(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	if err := user.ClearMovies(room); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func SwapMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.SwapMovieReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.SwapMoviePositions(req.Id1, req.Id2); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if err := room.Broadcast(&op.ElementMessage{
		ElementMessage: &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_MOVIES,
			Sender: user.Username,
		},
	}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func ChangeCurrentMovie(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	user := ctx.MustGet("user").(*op.User)

	req := model.IdCanEmptyReq{}
	if err := model.Decode(ctx, &req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if req.Id == "" {
		err := user.SetCurrentMovie(room, &dbModel.Movie{}, false)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
			return
		}
	} else if err := user.SetCurrentMovieByID(room, req.Id, true); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	current, err := genCurrent(ctx, room.Current(), user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
		return
	}
	current.UpdateSeek()

	if (current.Movie.Base.VendorInfo.Vendor == "") || (current.Movie.Base.VendorInfo.Vendor != "" && current.Movie.Base.VendorInfo.Shared) {
		if err := room.Broadcast(&op.ElementMessage{
			ElementMessage: &pb.ElementMessage{
				Type:    pb.ElementMessageType_CHANGE_CURRENT,
				Sender:  user.Username,
				Current: current.Proto(),
			},
		}); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
	} else {
		if err := room.SendToUser(user, &op.ElementMessage{
			ElementMessage: &pb.ElementMessage{
				Type:    pb.ElementMessageType_CHANGE_CURRENT,
				Sender:  user.Username,
				Current: current.Proto(),
			},
		}); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}

		m := &pb.ElementMessage{
			Type:   pb.ElementMessageType_CHANGE_CURRENT,
			Sender: user.Username,
		}
		if err := room.Broadcast(&op.ElementMessage{
			ElementMessage: m,
			BeforeSendFunc: func(sendTo *op.User) error {
				current, err := genCurrent(ctx, room.Current(), sendTo.ID)
				if err != nil {
					return err
				}
				current.UpdateSeek()
				m.Current = current.Proto()
				return nil
			},
		}, op.WithIgnoreId(user.ID)); err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
	}

	ctx.Status(http.StatusNoContent)
}

func ProxyMovie(ctx *gin.Context) {
	roomId := ctx.Param("roomId")
	if roomId == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("roomId is empty"))
		return
	}

	room, err := op.LoadOrInitRoomByID(roomId)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	m, err := room.GetMovieByID(ctx.Param("movieId"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}

	if !m.Base.Proxy || m.Base.Live || m.Base.RtmpSource {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("not support movie proxy"))
		return
	}

	if m.Base.VendorInfo.Vendor != "" {
		proxyVendorMovie(ctx, m.Movie)
		return
	}

	err = proxyURL(ctx, m.Base.Url, m.Base.Headers)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
		return
	}
}

func proxyURL(ctx *gin.Context, u string, headers map[string]string) error {
	if !conf.Conf.Proxy.AllowProxyToLocal {
		if l, err := utils.ParseURLIsLocalIP(u); err != nil {
			return err
		} else if l {
			return errors.New("not allow proxy to local")
		}
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	r.Header.Set("Range", ctx.GetHeader("Range"))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ctx.Header("Content-Type", resp.Header.Get("Content-Type"))
	ctx.Header("Content-Length", resp.Header.Get("Content-Length"))
	ctx.Header("Accept-Ranges", resp.Header.Get("Accept-Ranges"))
	ctx.Header("Cache-Control", resp.Header.Get("Cache-Control"))
	ctx.Header("Content-Range", resp.Header.Get("Content-Range"))
	ctx.Status(resp.StatusCode)
	_, err = io.Copy(ctx.Writer, resp.Body)
	return err
}

type FormatErrNotSupportFileType string

func (e FormatErrNotSupportFileType) Error() string {
	return fmt.Sprintf("not support file type %s", string(e))
}

func JoinLive(ctx *gin.Context) {
	room := ctx.MustGet("room").(*op.Room)
	// user := ctx.MustGet("user").(*op.User)

	movieId := strings.Trim(ctx.Param("movieId"), "/")
	fileExt := path.Ext(movieId)
	splitedMovieId := strings.Split(movieId, "/")
	channelName := strings.TrimSuffix(splitedMovieId[0], fileExt)
	channel, err := room.GetChannel(channelName)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
		return
	}
	switch fileExt {
	case ".flv":
		ctx.Header("Cache-Control", "no-store")
		w := httpflv.NewHttpFLVWriter(ctx.Writer)
		defer w.Close()
		channel.AddPlayer(w)
		w.SendPacket()
	case ".m3u8":
		ctx.Header("Cache-Control", "no-store")
		b, err := channel.GenM3U8File(func(tsName string) (tsPath string) {
			ext := "ts"
			if conf.Conf.Server.Rtmp.TsDisguisedAsPng {
				ext = "png"
			}
			return fmt.Sprintf("/api/movie/live/%s/%s.%s", channelName, tsName, ext)
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
			return
		}
		ctx.Data(http.StatusOK, hls.M3U8ContentType, b)
	case ".ts":
		if conf.Conf.Server.Rtmp.TsDisguisedAsPng {
			ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(FormatErrNotSupportFileType(fileExt)))
			return
		}
		b, err := channel.GetTsFile(splitedMovieId[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
			return
		}
		ctx.Header("Cache-Control", "public, max-age=90")
		ctx.Data(http.StatusOK, hls.TSContentType, b)
	case ".png":
		if !conf.Conf.Server.Rtmp.TsDisguisedAsPng {
			ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(FormatErrNotSupportFileType(fileExt)))
			return
		}
		b, err := channel.GetTsFile(splitedMovieId[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusNotFound, model.NewApiErrorResp(err))
			return
		}
		ctx.Header("Cache-Control", "public, max-age=90")
		img := image.NewGray(image.Rect(0, 0, 1, 1))
		img.Set(1, 1, color.Gray{uint8(rand.Intn(255))})
		cache := bytes.NewBuffer(make([]byte, 0, 71))
		err = png.Encode(cache, img)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		ctx.Data(http.StatusOK, "image/png", append(cache.Bytes(), b...))
	default:
		ctx.Header("Cache-Control", "no-store")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(FormatErrNotSupportFileType(fileExt)))
	}
}

func initBilibiliMPDCache(ctx context.Context, cli vendor.Bilibili, cookies []*http.Cookie, bvid string, cid, epid uint64, roomID, movieID string) *refreshcache.RefreshCache[*dbModel.MPDCache] {
	return refreshcache.NewRefreshCache[*dbModel.MPDCache](func() (*dbModel.MPDCache, error) {
		var m *mpd.MPD
		if bvid != "" && cid != 0 {
			resp, err := cli.GetDashVideoURL(ctx, &bilibili.GetDashVideoURLReq{
				Cookies: utils.HttpCookieToMap(cookies),
				Bvid:    bvid,
				Cid:     cid,
			})
			if err != nil {
				return nil, err
			}
			m, err = mpd.ReadFromString(resp.Mpd)
			if err != nil {
				return nil, err
			}
		} else if epid != 0 {
			resp, err := cli.GetDashPGCURL(ctx, &bilibili.GetDashPGCURLReq{
				Cookies: utils.HttpCookieToMap(cookies),
				Epid:    epid,
			})
			if err != nil {
				return nil, err
			}
			m, err = mpd.ReadFromString(resp.Mpd)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("bvid and epid are empty")
		}
		m.BaseURL = append(m.BaseURL, fmt.Sprintf("/api/movie/proxy/%s/", roomID))
		id := 0
		movies := []string{}
		for _, p := range m.Periods {
			for _, as := range p.AdaptationSets {
				for _, r := range as.Representations {
					for i := range r.BaseURL {
						movies = append(movies, r.BaseURL[i])
						r.BaseURL[i] = fmt.Sprintf("%s?id=%d", movieID, id)
						id++
					}
				}
			}
		}
		s, err := m.WriteToString()
		if err != nil {
			return nil, err
		}
		return &dbModel.MPDCache{
			URLs:    movies,
			MPDFile: s,
		}, nil
	}, time.Minute*119)
}

func initBilibiliShareCache(ctx context.Context, cli vendor.Bilibili, cookies []*http.Cookie, bvid string, cid, epid uint64) *refreshcache.RefreshCache[string] {
	return refreshcache.NewRefreshCache[string](func() (string, error) {
		var u string
		if bvid != "" {
			resp, err := cli.GetVideoURL(ctx, &bilibili.GetVideoURLReq{
				Cookies: utils.HttpCookieToMap(cookies),
				Bvid:    bvid,
				Cid:     cid,
			})
			if err != nil {
				return "", err
			}
			u = resp.Url
		} else if epid != 0 {
			resp, err := cli.GetPGCURL(ctx, &bilibili.GetPGCURLReq{
				Cookies: utils.HttpCookieToMap(cookies),
				Epid:    epid,
			})
			if err != nil {
				return "", err
			}
			u = resp.Url
		} else {
			return "", errors.New("bvid and epid are empty")
		}
		return u, nil
	}, time.Minute*119)
}

func proxyVendorMovie(ctx *gin.Context, movie *dbModel.Movie) {
	switch movie.Base.VendorInfo.Vendor {
	case dbModel.StreamingVendorBilibili:
		bvc, err := movie.Base.VendorInfo.Bilibili.InitOrLoadMPDCache(func(info *dbModel.BilibiliVendorInfo) (*refreshcache.RefreshCache[*dbModel.MPDCache], error) {
			v, err := db.FirstOrInitVendorByUserIDAndVendor(movie.CreatorID, dbModel.StreamingVendorBilibili)
			if err != nil {
				return nil, err
			}
			return initBilibiliMPDCache(ctx, vendor.BilibiliClient(info.VendorName), v.Cookies, info.Bvid, info.Cid, info.Epid, movie.RoomID, movie.ID), nil
		})
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}
		mpd, err := bvc.Get()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, model.NewApiErrorResp(err))
			return
		}

		if id := ctx.Query("id"); id == "" {
			ctx.Data(http.StatusOK, "application/dash+xml", []byte(mpd.MPDFile))
			return
		} else {
			streamId, err := strconv.Atoi(id)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorResp(err))
				return
			}
			if streamId >= len(mpd.URLs) {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("stream id out of range"))
				return
			}
			proxyURL(ctx, mpd.URLs[streamId], movie.Base.Headers)
			return
		}

	default:
		ctx.AbortWithStatusJSON(http.StatusBadRequest, model.NewApiErrorStringResp("vendor not support"))
		return
	}
}

func parse2VendorMovie(ctx context.Context, userID string, movie *dbModel.Movie) (err error) {
	if movie.Base.VendorInfo.Shared {
		userID = movie.CreatorID
	}

	switch movie.Base.VendorInfo.Vendor {
	case dbModel.StreamingVendorBilibili:
		if !movie.Base.Proxy {
			c, err := movie.Base.VendorInfo.Bilibili.InitOrLoadURLCache(userID, func(info *dbModel.BilibiliVendorInfo) (*refreshcache.RefreshCache[string], error) {
				v, err := db.FirstOrInitVendorByUserIDAndVendor(userID, dbModel.StreamingVendorBilibili)
				if err != nil {
					return nil, err
				}
				return initBilibiliShareCache(ctx, vendor.BilibiliClient(info.VendorName), v.Cookies, info.Bvid, info.Cid, info.Epid), nil
			})
			if err != nil {
				return err
			}

			data, err := c.Get()
			if err != nil {
				return err
			}

			movie.Base.Url = data
		} else {
			movie.Base.Type = "mpd"
		}

		return nil

	default:
		return fmt.Errorf("vendor not support")
	}
}
