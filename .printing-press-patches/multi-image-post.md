# Patch: multi-image organic posts

## Why

LinkedIn supports two distinct organic post shapes and they are not
interchangeable:

- `content.media` — one image, video, or document
- `content.multiImage` — two or more images as a single post

Sending an array under `content.media` is rejected by the API. This CLI
only implemented the first, so a multi-photo post — the thing LinkedIn's own
app calls a multi-photo or carousel post — was impossible to create.

Do not confuse this with the **Carousel** API, which is ads-only and requires
different permissions. That is a different product.

## What changed

### `internal/postiee/postiee.go`

- `PostContent.Media` became `*PostMedia` instead of an inline anonymous
  struct, so `multiImage` can sit beside it without either being emitted when
  absent. `omitempty` on a pointer omits a nil field; it does nothing for a
  zero struct.
- Added `PostMedia`, `MultiImageBlock`, `MultiImageItem`.
- Added `WithMultiImage(in, urns, alts)`. Per-image alt text, positionally
  matched. Zero images is a no-op that leaves the request untouched, because
  an empty `content` object is not the same as omitting `content`.

### `internal/cli/postiee_history.go`

- `--image` became a repeatable `StringSliceVar`. One value still produces a
  single-image post; two or more produce a multiImage post.
- `--alt` accepts pipe-separated alt text for multi-image posts. A pipe
  because alt text is prose and rarely contains one; a comma would split a
  sentence.
- An upload failure names which image failed out of how many, since a
  five-image post that dies on the third is otherwise opaque.

## Tests

`internal/postiee/postiee_test.go`

- `TestCreatePostIncludesAltTextWhenProvided` — updated for the pointer
  change; this is the pre-existing test that caught it.
- `TestWithMultiImageEmitsMultiImageShape` — asserts `multiImage` present and
  `media` absent. The two shapes being confused is the whole failure mode.
- `TestWithMultiImageKeepsAltTextPerImage`
- `TestWithMultiImageToleratesMissingAlts` — fewer alts than images must not
  shift a value onto the wrong image.
- `TestWithMultiImageEmptyDoesNotAttachContent` — caught a real bug where zero
  images emitted an empty `multiImage.images` array.

## Usage

```bash
postiee post --image a.png --image b.png --image c.png \
  --alt "first|second|third" "caption text"
```

## Risk

The generated tree can be overwritten by a reprint. If multi-image support
disappears after regenerating, re-apply this patch — or better, fix it
upstream in CLI Printing Press so it survives.