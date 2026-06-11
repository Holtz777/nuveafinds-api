package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Holtz777/nuveafinds-api/internal/ai"
	"github.com/Holtz777/nuveafinds-api/internal/config"
	"github.com/Holtz777/nuveafinds-api/internal/httpx"
	"github.com/Holtz777/nuveafinds-api/internal/pinterest"
)

type Deps struct {
	Config    *config.Config
	AI        *ai.Client
	Pinterest *pinterest.Client
	Tokens    *pinterest.TokenStore
}

func Register(mux *http.ServeMux, d *Deps) {
	mux.HandleFunc("GET /health", d.health)
	mux.HandleFunc("GET /boards", d.listBoards)
	mux.HandleFunc("POST /pin-upload", d.pinUpload)
	mux.HandleFunc("POST /pin-register-video", d.pinRegisterVideo)
	mux.HandleFunc("POST /proxy/upload-video", d.proxyUploadVideo)
	mux.HandleFunc("POST /pin-publish", d.pinPublish)
	mux.HandleFunc("GET /token-info", d.tokenInfo)
	mux.HandleFunc("POST /token-refresh", d.tokenRefresh)
	mux.HandleFunc("GET /mode", d.getMode)
	mux.HandleFunc("POST /mode", d.setMode)
}

func (d *Deps) health(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (d *Deps) listBoards(w http.ResponseWriter, r *http.Request) {
	if d.Config.PinterestAccessToken == "" && d.Tokens == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "PINTEREST_ACCESS_TOKEN not configured")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	boards, err := d.Pinterest.ListBoards(ctx)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "pinterest list boards: "+err.Error())
		return
	}

	type boardOut struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug,omitempty"`
	}
	out := make([]boardOut, len(boards))
	for i, b := range boards {
		out[i] = boardOut{ID: b.ID, Name: b.Name, Slug: d.Config.BoardMapInverse[b.ID]}
	}

	envFormat := make([]string, len(boards))
	for i, b := range boards {
		slug := d.Config.BoardMapInverse[b.ID]
		if slug != "" {
			envFormat[i] = slug + "=" + b.ID
		} else {
			envFormat[i] = b.Name + "=" + b.ID
		}
	}

	httpx.Success(w, map[string]any{
		"boards":        out,
		"envMapFormat":  envFormat,
		"boardMapValue": fmt.Sprint(strings.Join(envFormat, ",")),
	})
}

func (d *Deps) tokenInfo(w http.ResponseWriter, r *http.Request) {
	if d.Tokens == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "token store not initialized")
		return
	}
	httpx.Success(w, d.Tokens.Info())
}

func (d *Deps) tokenRefresh(w http.ResponseWriter, r *http.Request) {
	if d.Tokens == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "token store not initialized")
		return
	}
	oauthClient := &http.Client{Timeout: 30 * time.Second}
	if err := d.Tokens.Refresh(oauthClient); err != nil {
		httpx.Error(w, http.StatusBadGateway, "token refresh failed: "+err.Error())
		return
	}
	httpx.Success(w, d.Tokens.Info())
}

func (d *Deps) getMode(w http.ResponseWriter, r *http.Request) {
	httpx.Success(w, map[string]any{
		"sandbox":   d.Pinterest.IsSandbox(),
		"baseUrl":   d.Pinterest.BaseURL,
	})
}

type modeReq struct {
	Sandbox bool `json:"sandbox"`
}

func (d *Deps) setMode(w http.ResponseWriter, r *http.Request) {
	var in modeReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	d.Pinterest.SetSandbox(in.Sandbox)
	mode := "production"
	if in.Sandbox {
		mode = "sandbox"
	}
	log.Printf("pinterest: mode switched to %s", mode)
	httpx.Success(w, map[string]any{
		"sandbox": d.Pinterest.IsSandbox(),
		"baseUrl": d.Pinterest.BaseURL,
	})
}

func (d *Deps) pinUpload(w http.ResponseWriter, r *http.Request) {
	var in ai.PinGenRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if in.ProductName == "" || in.AffiliateLink == "" || in.ProductImageURL == "" {
		httpx.Error(w, http.StatusBadRequest, "productName, affiliateLink and productImageUrl are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	out, err := d.AI.GeneratePin(ctx, in)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "AI generation failed: "+err.Error())
		return
	}

	httpx.Success(w, out)
}

type registerVideoReq struct {
	FileName string `json:"fileName"`
	MimeType string `json:"mimeType"`
	FileSize int64  `json:"fileSize"`
}

type registerVideoResp struct {
	MediaID          string            `json:"mediaId"`
	UploadURL        string            `json:"uploadUrl"`
	UploadParameters map[string]string `json:"uploadParameters"`
}

func (d *Deps) pinRegisterVideo(w http.ResponseWriter, r *http.Request) {
	var in registerVideoReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if in.FileName == "" {
		httpx.Error(w, http.StatusBadRequest, "fileName is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	reg, err := d.Pinterest.RegisterVideo(ctx)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "pinterest register failed: "+err.Error())
		return
	}

	httpx.Success(w, registerVideoResp{
		MediaID:          reg.MediaID,
		UploadURL:        reg.UploadURL,
		UploadParameters: reg.UploadParameters,
	})
}

const maxUploadSize = 2 * 1024 * 1024 * 1024

func (d *Deps) proxyUploadVideo(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+10*1024*1024)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpx.Error(w, http.StatusBadRequest, "parse multipart: "+err.Error())
		return
	}

	uploadURL := r.FormValue("upload_url")
	paramsRaw := r.FormValue("upload_parameters")
	if uploadURL == "" || paramsRaw == "" {
		httpx.Error(w, http.StatusBadRequest, "upload_url and upload_parameters are required")
		return
	}

	var params map[string]string
	if err := json.Unmarshal([]byte(paramsRaw), &params); err != nil {
		var loose map[string]any
		if err2 := json.Unmarshal([]byte(paramsRaw), &loose); err2 != nil {
			httpx.Error(w, http.StatusBadRequest, "upload_parameters must be valid JSON: "+err.Error())
			return
		}
		params = make(map[string]string, len(loose))
		for k, v := range loose {
			params[k] = fmt.Sprint(v)
		}
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "file field missing: "+err.Error())
		return
	}
	defer file.Close()

	if header.Size == 0 {
		httpx.Error(w, http.StatusBadRequest, "file is empty")
		return
	}
	if header.Size > maxUploadSize {
		httpx.Error(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("file too large: %d bytes (max %d)", header.Size, maxUploadSize))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	if err := d.Pinterest.UploadToS3(ctx, uploadURL, params, header.Filename, header.Header.Get("Content-Type"), file); err != nil {
		httpx.Error(w, http.StatusBadGateway, err.Error())
		return
	}

	httpx.Success(w, map[string]string{"message": "Video uploaded to Pinterest successfully."})
}

type publishReq struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	AffiliateLink   string `json:"affiliateLink"`
	Board           string `json:"board"`
	MediaID         string `json:"mediaId"`
	ProductImageURL string `json:"productImageUrl"`
	ProductTags     string `json:"productTags,omitempty"`
}

type publishResp struct {
	PinID   string `json:"pinId"`
	PinURL  string `json:"pinUrl"`
	Title   string `json:"title"`
	BoardID string `json:"boardId"`
}

func (d *Deps) pinPublish(w http.ResponseWriter, r *http.Request) {
	var in publishReq
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	missing := []string{}
	if in.Title == "" {
		missing = append(missing, "title")
	}
	if in.Description == "" {
		missing = append(missing, "description")
	}
	if in.AffiliateLink == "" {
		missing = append(missing, "affiliateLink")
	}
	if in.Board == "" {
		missing = append(missing, "board")
	}
	if in.MediaID == "" {
		missing = append(missing, "mediaId")
	}
	if in.ProductImageURL == "" {
		missing = append(missing, "productImageUrl")
	}
	if len(missing) > 0 {
		httpx.Error(w, http.StatusBadRequest, "missing required fields: "+fmt.Sprint(missing))
		return
	}

	boardID, ok := d.Config.BoardMap[in.Board]
	if !ok {
		httpx.Error(w, http.StatusBadRequest, "unknown board slug: "+in.Board+" (add it to PINTEREST_BOARD_MAP)")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	if err := d.Pinterest.WaitForMediaReady(ctx, in.MediaID, 4*time.Minute); err != nil {
		httpx.Error(w, http.StatusBadGateway, "media not ready: "+err.Error())
		return
	}

	pin, err := d.Pinterest.CreateVideoPin(ctx, pinterest.CreatePinRequest{
		BoardID:     boardID,
		Title:       in.Title,
		Description: in.Description,
		Link:        in.AffiliateLink,
		MediaID:     in.MediaID,
		CoverImage:  in.ProductImageURL,
	})
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, "create pin: "+err.Error())
		return
	}

	httpx.Success(w, publishResp{
		PinID:   pin.ID,
		PinURL:  "https://www.pinterest.com/pin/" + pin.ID + "/",
		Title:   pin.Title,
		BoardID: pin.BoardID,
	})
}
