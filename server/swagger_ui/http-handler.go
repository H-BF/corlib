package swagger_ui

import (
	"encoding/json"
	"net/http"

	"github.com/go-openapi/spec"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
)

// SwaggerDef = spec.Swagger
type SwaggerDef = spec.Swagger

//NewHandler делаем такую штуку которая покажет ним сваггер документ
func NewHandler(sd *SwaggerDef) (http.Handler, error) {
	const api = "swagger_ui.NewHandler"

	doc, err := json.Marshal(sd)
	if err != nil {
		return nil, errors.Wrapf(err, "%s: marshall Swagger spec ", api)
	}
	var assets http.FileSystem
	if assets, err = fs.NewWithNamespace(Assets); err != nil {
		return nil, errors.Wrapf(err, "%s: make swagger-ui assets", api)
	}
	ret := &handlerImpl{
		swaggerDoc: doc,
		fileServer: http.FileServer(assets),
	}
	return ret, nil
}

var _ = NewHandler

type handlerImpl struct {
	swaggerDoc []byte
	fileServer http.Handler
}

func (h *handlerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/swagger.json" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(h.swaggerDoc)
		return
	}
	if r.URL.Path == "" {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusTemporaryRedirect)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}
