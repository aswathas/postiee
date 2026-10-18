// Hand-written commands layered on the generated CLI. Lives outside generated
// files so a regeneration never conflicts.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/aswathas/postiee/internal/cliutil"
	"github.com/aswathas/postiee/internal/history"
	"github.com/aswathas/postiee/internal/postiee"
)

// postieeToken resolves the OAuth token, preferring the env var then falling
// back to the credential store the generated auth layer writes. `auth login`
// persists to that store, so reading only the env var would ignore a completed
// login.
func postieeToken() string {
	for _, k := range []string{"POSTIEE_TOKEN", "LINKEDIN_ACCESS_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	if creds, ok, err := cliutil.LoadCredentials(); err == nil && ok && creds != nil {
		if v := strings.TrimSpace(creds.AccessToken); v != "" {
			return v
		}
	}
	return ""
}

// historyStore opens the local history database.
func historyStore() (*history.Store, error) {
	path := strings.TrimSpace(os.Getenv("POSTIEE_HISTORY_DB"))
	if path == "" {
		path = history.DefaultPath()
	}
	return history.Open(path)
}

// newPostieePostCmd posts as the authenticated member and records it locally.
//
// This is the reason the tool exists: LinkedIn will not let anyone read your
// posts back (`r_member_social` is closed), so whatever this writes to the
// local database is the only durable record of the post.
func newPostieePostCmd(flags *rootFlags) *cobra.Command {
	var images []string
	var altText, visibility, file string

	cmd := &cobra.Command{
		Use:   "post <text>",
		Short: "Post to LinkedIn and record it in local history",
		Long: "Posts to your own LinkedIn feed, then writes the post to the local\n" +
			"history database.\n\n" +
			"The history write is not a convenience. LinkedIn's read-your-own-posts\n" +
			"permission is closed, so this local record is the only place the post id\n" +
			"will ever exist.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if file != "" {
				b, err := os.ReadFile(file)
				if err != nil {
					return fmt.Errorf("reading %s: %w", file, err)
				}
				text = strings.TrimSpace(string(b))
			}
			if strings.TrimSpace(text) == "" {
				return fmt.Errorf("nothing to post: pass text as an argument or use --file")
			}

			// Short-circuit before any network call. Without this, --dry-run
			// publishes a real post: CreatePost is not dry-run aware.
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "post to LinkedIn and record it in local history")
			}

			token := postieeToken()
			if token == "" {
				return fmt.Errorf("no access token found; run 'postiee auth login' or set POSTIEE_TOKEN")
			}

			ctx := context.Background()
			c := &postiee.Client{Token: token}

			// Resolve the member id once and reuse it: person ids are app-scoped
			// and must never be hardcoded or typed by hand.
			author, err := c.AuthorURN(ctx)
			if err != nil {
				return err
			}
			ctx = postiee.WithAuthor(ctx, author)

			req := postiee.CreatePostRequest{
				Author:     author,
				Commentary: text,
				Visibility: visibility,
			}

			// Images attach through the /rest/posts content field; the legacy
			// UGC media array is not modelled. Upload first, then post exactly
			// once — two CreatePost calls would publish two real posts.
			//
			// One image uses content.media. Two or more use content.multiImage,
			// which is a different shape LinkedIn will not accept as media.
			var mediaURN string
			switch len(images) {
			case 0:
			case 1:
				mediaURN, err = c.UploadImage(ctx, images[0], altText)
				if err != nil {
					return err
				}
				req = withImage(req, mediaURN, altText)
			default:
				alts := splitAlts(altText)
				urns := make([]string, 0, len(images))
				for i, path := range images {
					alt := ""
					if i < len(alts) {
						alt = alts[i]
					}
					urn, err := c.UploadImage(ctx, path, alt)
					if err != nil {
						return fmt.Errorf("uploading %s (image %d of %d): %w", filepath.Base(path), i+1, len(images), err)
					}
					urns = append(urns, urn)
				}
				req = postiee.WithMultiImage(req, urns, alts)
			}

			created, err := c.CreatePost(ctx, req)
			if err != nil {
				return err
			}

			st, serr := historyStore()
			if serr != nil {
				// The post exists; failing to record it is bad but losing the
				// just-created post id would be worse. Print it either way.
				fmt.Fprintf(cmd.OutOrStdout(), "Posted %s\nWARNING: history unavailable: %v\n", created.Permalink, serr)
				return nil
			}
			defer st.Close()

			if err := st.Record(ctx, history.Entry{
				URN: created.URN, Body: text, Permalink: created.Permalink,
				MediaURN: mediaURN, Visibility: orDefault(visibility, "PUBLIC"),
			}); err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Posted %s\nWARNING: could not record to history: %v\n", created.Permalink, err)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), created.Permalink)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&images, "image", nil,
		"path to an image to attach (jpg, png, gif). Repeat for several: two or more images post as one multi-image post.")
	cmd.Flags().StringVar(&altText, "alt", "",
		"alt text for the image. With several images, one per image separated by a pipe |")
	cmd.Flags().StringVarP(&visibility, "visibility", "v", "PUBLIC", "PUBLIC, CONNECTIONS, or LOGGED_IN")
	cmd.Flags().StringVar(&file, "file", "", "read the post body from a file")
	return cmd
}

func withImage(in postiee.CreatePostRequest, urn, alt string) postiee.CreatePostRequest {
	in.Content = &postiee.PostContent{
		Media: &postiee.PostMedia{ID: urn, AltText: alt},
	}
	return in
}

// splitAlts separates per-image alt text. A pipe is the delimiter because alt
// text is prose and rarely contains one; commas would split a sentence.
func splitAlts(alt string) []string {
	if strings.TrimSpace(alt) == "" {
		return nil
	}
	parts := strings.Split(alt, "|")
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.TrimSpace(p)
	}
	return out
}

func orDefault(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

// newPostieeHistoryCmd lists everything postiee has posted, newest first.
func newPostieeHistoryCmd(flags *rootFlags) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:     "history",
		Short:   "List posts recorded locally, newest first",
		// No "posts" alias: it shadows the generated `posts` parent group and
		// makes `posts delete` unreachable. Use `history` or `posts list`.
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := historyStore()
			if err != nil {
				return err
			}
			defer st.Close()

			entries, err := st.Search(context.Background(), "", limit)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No posts recorded yet.")
				return nil
			}
			for _, e := range entries {
				first := e.Body
				if i := strings.IndexByte(first, '\n'); i > 0 {
					first = first[:i]
				}
				if len(first) > 72 {
					first = first[:72] + "..."
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n",
					e.CreatedAt.Format("2006-01-02"), e.URN, first)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum entries to show")
	return cmd
}

// newPostieeSearchCmd full-text searches recorded posts.
func newPostieeSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search your recorded posts",
		Long: "Searches post bodies recorded by 'postiee post'.\n\n" +
			"This searches only what postiee posted through this tool. Posts made\n" +
			"from the LinkedIn app or website are not visible to any API, so they\n" +
			"cannot be included.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := historyStore()
			if err != nil {
				return err
			}
			defer st.Close()

			entries, err := st.Search(context.Background(), strings.Join(args, " "), limit)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No matching posts.")
				return nil
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(entries)
			}
			for _, e := range entries {
				fmt.Fprintf(cmd.OutOrStdout(), "%s  %s\n%s\n\n",
					e.CreatedAt.Format("2006-01-02"), e.URN, e.Body)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 25, "maximum results")
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit results as JSON")
	return cmd
}

// newPostieeForgetCmd drops a post from local history.
func newPostieeForgetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "forget <post-urn>",
		Short: "Remove a post from local history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "remove "+args[0]+" from local history")
			}
			st, err := historyStore()
			if err != nil {
				return err
			}
			defer st.Close()
			if err := st.Forget(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from local history.\n", args[0])
			return nil
		},
	}
}

// newPostieeUnpublishCmd deletes a post on LinkedIn and drops it from local
// history. `forget` only removes the local record, which leaves the post live —
// keeping both commands means the two halves can never be confused.
//
// Unlike the generated `posts delete`, this takes no --authorization flag: a
// token on the command line lands in shell history and process listings.
func newPostieeUnpublishCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "unpublish <post-urn>",
		Short: "Delete a post on LinkedIn and remove it from local history",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			urn := args[0]
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "delete "+urn+" on LinkedIn and remove it from local history")
			}

			token := postieeToken()
			if token == "" {
				return fmt.Errorf("no access token found; run 'postiee auth login' or set POSTIEE_TOKEN")
			}
			c := &postiee.Client{Token: token}

			// Delete remotely first. If LinkedIn refuses, the local record stays
			// so the post is still findable and the user can retry.
			if err := c.DeletePost(context.Background(), urn); err != nil {
				return err
			}

			// Remote delete succeeded, so the local row is now a lie. Drop it.
			st, err := historyStore()
			if err != nil {
				return err
			}
			defer st.Close()
			if err := st.Forget(context.Background(), urn); err != nil {
				return fmt.Errorf("deleted on LinkedIn but could not remove %s from local history: %w", urn, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s on LinkedIn and removed it from local history.\n", urn)
			return nil
		},
	}
}

// init registers the hand-written commands through the generator's own
// novelCommandHooks extension point, so root.go is never edited and a
// regeneration cannot lose them.
func init() {
	novelCommandHooks = append(novelCommandHooks, func(root *cobra.Command, flags *rootFlags) {
		root.AddCommand(newPostieePostCmd(flags))
		root.AddCommand(newPostieeHistoryCmd(flags))
		root.AddCommand(newPostieeSearchCmd(flags))
		root.AddCommand(newPostieeUnpublishCmd(flags))
		root.AddCommand(newPostieeForgetCmd(flags))
	})
}