package internal

import (
	"github.com/whyweexist/evm/common"
	"github.com/whyweexist/evm/config"
	"github.com/whyweexist/evm/internal/delivery/rest"
	"github.com/whyweexist/evm/internal/job"
	restRepository "github.com/whyweexist/evm/internal/repository/rest"
	sqlRepository "github.com/whyweexist/evm/internal/repository/sql"
	"github.com/whyweexist/evm/internal/service"
	"github.com/whyweexist/evm/pkg/http/wrapper"
	"github.com/whyweexist/evm/web"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io"
	"io/fs"
	"net/http"
)

type embeddedFile struct {
	fs.File
}

func (f *embeddedFile) Close() error {
	return nil
}

func (f *embeddedFile) Seek(offset int64, whence int) (int64, error) {
	return f.File.(io.Seeker).Seek(offset, whence)
}

func (boot *boostrap) newPublicAPIProvider() {
	router := boot.engine
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusTemporaryRedirect, "/admin")
	})
	router.StaticFS("/admin", http.FS(web.SPAAssets()))
	router.NoRoute(func(ctx *gin.Context) {
		file, err := web.SPAAssets().Open("index.html")
		if err != nil {
			ctx.String(http.StatusInternalServerError,
				"failed to open spa file: ", err.Error())
			return
		}
		defer func() { _ = file.Close() }()
		fileInfo, err := file.Stat()
		if err != nil {
			ctx.String(http.StatusInternalServerError,
				"failed to get spa file info: ", err.Error())
			return
		}
		http.ServeContent(ctx.Writer, ctx.Request, fileInfo.Name(), fileInfo.ModTime(), &embeddedFile{file})
	})
	router.GET("/ping", func(ctx *gin.Context) {
		wrapper.NewHTTPRespondWrapper(
			ctx, http.StatusOK, "PONG")
	})
	router.GET("/docs/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.DefaultModelsExpandDepth(
				common.SwaggerDefaultModelsExpandDepth)))
}

func (boot *boostrap) newTixAPIProvider() {
	routerGroupV1 := boot.engine.Group("api/v1")
	authRepository := restRepository.NewAuthRESTRepository(
		config.Instance.SupabaseProjectURL,
		config.Instance.SupabaseAPIKey,
		config.Instance.SupabaseAPIKeyRoot)
	tixRepository := sqlRepository.NewTixPostgreSQLRepository(boot.db)
	gsRepository := restRepository.NewGoogleServiceRepository(&config.FormsServiceWrapper{
		Service: boot.googleForm.Forms,
	})
	tixService := service.NewTixService(
		service.WithGoogleServiceRepository(gsRepository),
		service.WithRedisCache(boot.cache),
		service.WithAuthRESTRepository(authRepository),
		service.WithPostgreSQLRepository(tixRepository),
		service.WithMailer(boot.mailer))
	rest.NewAccountRESTHandler(routerGroupV1, tixService)
	rest.NewEventRESTHandler(routerGroupV1, tixService)
	rest.NewUserRESTHandler(routerGroupV1, tixService)
	job.NewEventJob(tixService, boot.cache)
}
