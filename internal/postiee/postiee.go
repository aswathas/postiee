// Package postiee adds the LinkedIn-specific behaviour the OpenAPI spec cannot
// express. Everything here is hand-written on top of the generated client and
// lives outside generated files so a regeneration never conflicts.
package postiee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Version is the LinkedIn-Version header value. Monthly YYYYMM; each version is
// supported at least 12 months. 202510 sunsets 2026-10-15. Change it here and
// nowhere else.
const Version = "202609"

// PostIDHeader carries the created post's urn. LinkedIn returns 201 with an
// EMPTY body on create, so this header is the only place the ID exists — and
// `r_member_social` is a closed permission, meaning the API will never let us
// read the post back. If this ID is not captured, it is lost permanently.
const PostIDHeader = "x-restli-id"

// Client is the minimum surface postiee needs. Accepting an interface keeps
// this testable without a live LinkedIn account.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Token   string
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) endpoint(path string) string {
	base := c.BaseURL
	if base == "" {
		base = "https://api.linkedin.com"
	}
	return strings.TrimRight(base, "/") + path
}

// headers builds the common header set. X-Restli-Protocol-Version is required
// on every versioned route.
func (c *Client) headers() http.Header {
	h := http.Header{}
	if c.Token != "" {
		h.Set("Authorization", "Bearer "+c.Token)
	}
	h.Set("Linkedin-Version", Version)
	h.Set("X-Restli-Protocol-Version", "2.0.0")
	h.Set("Content-Type", "application/json")
	return h
}

// AuthorURN resolves the authenticated member's person urn.
//
// Person IDs are scoped to a single LinkedIn application: the same human has a
// different id under a different app, and a 404 results if you reuse one
// across apps. So this must be resolved at runtime via /v2/userinfo, never
// hardcoded and never copied from another app's config.
func (c *Client) AuthorURN(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint("/v2/userinfo"), nil)
	if err != nil {
		return "", err
	}
	req.Header = c.headers()

	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("resolving member id: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolving member id: %s: %s", resp.Status, snippet(body))
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(body, &claims); err != nil {
		return "", fmt.Errorf("decoding member id: %w", err)
	}
	if claims.Sub == "" {
		return "", fmt.Errorf("member id response had no `sub` field; token may lack the openid scope")
	}
	return "urn:li:person:" + claims.Sub, nil
}

// CreatePostRequest is the POST /rest/posts body. Content is omitted entirely
// for a text-only post rather than sent as an empty object.
type CreatePostRequest struct {
	Author                   string `json:"author"`
	Commentary               string `json:"commentary"`
	Visibility               string `json:"visibility"`
	LifecycleState           string `json:"lifecycleState"`
	IsReshareDisabledByAuthor bool   `json:"isReshareDisabledByAuthor,omitempty"`
	Distribution             struct {
		FeedDistribution               string `json:"feedDistribution"`
		TargetEntities                 []any  `json:"targetEntities"`
		ThirdPartyDistributionChannels []any  `json:"thirdPartyDistributionChannels"`
	} `json:"distribution"`
	Content *PostContent `json:"content,omitempty"`
}

// PostContent carries the attached media. AltText is optional but strongly
// recommended: LinkedIn caps it at 4,086 characters and suggests under 120.
type PostContent struct {
	Media struct {
		ID      string `json:"id"`
		AltText string `json:"altText,omitempty"`
	} `json:"media"`
}

// CreatedPost is the result of a successful create.
type CreatedPost struct {
	URN       string
	Permalink string
}

// CreatePost posts as the authenticated member.
//
// lifecycleState accepts PUBLISHED only at creation; DRAFT and
// PUBLISH_REQUESTED are response-only. There is no publishAt field anywhere in
// LinkedIn's post schema, so scheduling is not an API feature.
func (c *Client) CreatePost(ctx context.Context, in CreatePostRequest) (CreatedPost, error) {
	if in.Author == "" {
		return CreatedPost{}, fmt.Errorf("author urn is required; call AuthorURN first")
	}
	if strings.TrimSpace(in.Commentary) == "" {
		return CreatedPost{}, fmt.Errorf("commentary is empty; nothing to post")
	}
	if in.LifecycleState == "" {
		in.LifecycleState = "PUBLISHED"
	}
	if in.Visibility == "" {
		in.Visibility = "PUBLIC"
	}
	if in.Distribution.FeedDistribution == "" {
		in.Distribution.FeedDistribution = "MAIN_FEED"
	}
	if in.Distribution.TargetEntities == nil {
		in.Distribution.TargetEntities = []any{}
	}
	if in.Distribution.ThirdPartyDistributionChannels == nil {
		in.Distribution.ThirdPartyDistributionChannels = []any{}
	}

	body, err := json.Marshal(in)
	if err != nil {
		return CreatedPost{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint("/rest/posts"), bytes.NewReader(body))
	if err != nil {
		return CreatedPost{}, err
	}
	req.Header = c.headers()

	resp, err := c.http().Do(req)
	if err != nil {
		return CreatedPost{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return CreatedPost{}, fmt.Errorf("create post failed: %s: %s", resp.Status, snippet(respBody))
	}

	urn := resp.Header.Get(PostIDHeader)
	if urn == "" {
		// The post may well exist, but without this ID we cannot delete it or
		// record it locally, and LinkedIn will not let us read it back.
		return CreatedPost{}, fmt.Errorf(
			"post returned %s but no %s header. The post may have been created; "+
				"check your profile before retrying, or you may duplicate it",
			resp.Status, PostIDHeader)
	}

	return CreatedPost{URN: urn, Permalink: Permalink(urn)}, nil
}

// Permalink renders the human-facing URL for a post urn.
func Permalink(urn string) string {
	if urn == "" {
		return ""
	}
	return "https://www.linkedin.com/feed/update/" + urn + "/"
}

// UploadImage performs the two-step image upload and returns the final urn.
//
// There is NO finalizeUpload for images — that step belongs to the Videos API.
// initializeUpload already returns the final urn, so the only work left is
// putting the bytes at the signed URL.
//
// The PUT's exact required headers are not documented on any page that still
// exists (the vector-asset-api page 404s), so this sends the Authorization
// header with a binary body and accepts 200 or 201. If LinkedIn changes this,
// the failure surfaces here rather than silently producing a broken post.
func (c *Client) UploadImage(ctx context.Context, path, altText string) (string, error) {
	owner := c.authorURNFromContext(ctx)

	initReq := map[string]any{"initializeUploadRequest": map[string]any{"owner": owner}}
	body, err := json.Marshal(initReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.endpoint("/rest/images?action=initializeUpload"), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header = c.headers()

	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("initializing image upload: %w", err)
	}
	defer resp.Body.Close()

	initBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("initializing image upload: %s: %s", resp.Status, snippet(initBody))
	}

	var initResp struct {
		Value struct {
			UploadURL         string `json:"uploadUrl"`
			UploadURLExpiresAt int64 `json:"uploadUrlExpiresAt"`
			Image             string `json:"image"`
		} `json:"value"`
	}
	if err := json.Unmarshal(initBody, &initResp); err != nil {
		return "", fmt.Errorf("decoding initializeUpload response: %w", err)
	}
	if initResp.Value.Image == "" {
		return "", fmt.Errorf("initializeUpload returned no image urn; cannot attach media")
	}
	if initResp.Value.UploadURL == "" {
		return "", fmt.Errorf("initializeUpload returned no uploadUrl; cannot upload bytes")
	}

	image, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}

	// The signed URL is single-use and time-limited; never cache it.
	put, err := http.NewRequestWithContext(ctx, http.MethodPut, initResp.Value.UploadURL, bytes.NewReader(image))
	if err != nil {
		return "", err
	}
	put.Header = http.Header{}
	if c.Token != "" {
		put.Header.Set("Authorization", "Bearer "+c.Token)
	}
	put.ContentLength = int64(len(image))

	putResp, err := c.http().Do(put)
	if err != nil {
		return "", fmt.Errorf("uploading image bytes: %w", err)
	}
	defer putResp.Body.Close()
	putBody, _ := io.ReadAll(io.LimitReader(putResp.Body, 1<<20))

	if putResp.StatusCode != http.StatusCreated && putResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image upload failed: %s: %s", putResp.Status, snippet(putBody))
	}
	return initResp.Value.Image, nil
}

// EncodeURN percent-encodes a LinkedIn urn for use as a URL path segment.
//
// url.PathEscape is not enough here: it leaves ':' literal because Go allows
// colons in path segments, but LinkedIn requires the fully encoded form
// (urn%3Ali%3AugcPost%3A123).
func EncodeURN(urn string) string {
	return url.QueryEscape(urn)
}

// DeletePost removes a post by urn. Idempotent: LinkedIn returns 204 whether or
// not the post existed.
func (c *Client) DeletePost(ctx context.Context, urn string) error {
	if urn == "" {
		return fmt.Errorf("post urn is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.endpoint("/rest/posts/"+EncodeURN(urn)), nil)
	if err != nil {
		return err
	}
	req.Header = c.headers()
	req.Header.Set("X-RestLi-Method", "DELETE")

	resp, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s: %s", resp.Status, snippet(body))
	}
	return nil
}

// authorURNFromContext reads the author the caller already resolved, so the
// upload path does not make a second /v2/userinfo round trip.
func (c *Client) authorURNFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(authorKey{}).(string); ok && v != "" {
		return v
	}
	// Fall back to resolving rather than failing: a wrong author is worse than
	// an extra request.
	if urn, err := c.AuthorURN(ctx); err == nil {
		return urn
	}
	return ""
}

type authorKey struct{}

// WithAuthor attaches a resolved author urn to the context so a single
// command resolves the member id once and reuses it across calls.
func WithAuthor(ctx context.Context, urn string) context.Context {
	return context.WithValue(ctx, authorKey{}, urn)
}

// snippet trims an error body so a failure message stays readable.
func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "(empty body)"
	}
	if len(s) > 400 {
		return s[:400] + "..."
	}
	return s
}