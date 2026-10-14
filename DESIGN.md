---
name: postiee
description: A terminal-grade, near-black system where the mono face is the voice and the scope table is the argument.
colors:
  canvas: "#000000"
  canvas-alt: "#0a0a0c"
  surface: "#121215"
  surface-2: "#1a1a1f"
  terminal: "#0d0d10"
  ink: "#f5f5f7"
  ink-2: "#a1a1a6"
  ink-3: "#86868c"
  accent: "#30d158"
  accent-dim: "#28b947"
  accent-ink: "#04150a"
  accent-hover: "#46e06d"
  signal-blue: "#2997ff"
  warn: "#ffd60a"
  danger: "#ff6961"
  hairline: "rgba(255, 255, 255, 0.11)"
  hairline-2: "rgba(255, 255, 255, 0.06)"
typography:
  display:
    fontFamily: "JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "clamp(1.75rem, .75rem + 4.6vw, 5rem)"
    fontWeight: 600
    lineHeight: 1.06
    letterSpacing: "-.045em"
  headline:
    fontFamily: "JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "clamp(1.6rem, 1.1rem + 2.2vw, 2.5rem)"
    fontWeight: 600
    lineHeight: 1.14
    letterSpacing: "-.035em"
  title:
    fontFamily: "JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "-.015em"
  body:
    fontFamily: "IBM Plex Sans, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.65
    letterSpacing: ".005em"
  label:
    fontFamily: "JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: ".78rem"
    fontWeight: 500
    letterSpacing: ".06em"
rounded:
  sm: "4px"
  md: "6px"
  lg: "8px"
spacing:
  nav-h: "3rem"
  gutter: "clamp(1.25rem, 5vw, 3rem)"
  measure: "38rem"
  shell: "80rem"
  stage: "68rem"
  cards: "78rem"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.accent-ink}"
    rounded: "{rounded.sm}"
    padding: "0 1.4rem"
    height: "2.75rem"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  button-ghost:
    textColor: "{colors.ink}"
    rounded: "{rounded.sm}"
    padding: "0 1.4rem"
    height: "2.75rem"
  card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.md}"
    padding: "1.75rem 1.5rem"
  codeblock:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-2}"
    rounded: "{rounded.sm}"
    padding: "1.1rem 1.25rem"
  note-warn:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-2}"
    rounded: "{rounded.sm}"
    padding: "1rem 1.25rem"
  badge:
    textColor: "{colors.accent}"
    rounded: "3px"
    padding: ".1rem .5rem"
---

# Design System: postiee

<!-- Every value below was read from `docs-site/src/styles/global.css`,
     `docs-site/src/layouts/*.astro` and `docs-site/src/components/*.astro`
     as of the pass that wrote this file. Prose restates a token only to say
     where it is used; the frontmatter is the normative value. -->

## Overview

**Creative North Star: "The Terminal, Argued"**

postiee is a terminal tool whose entire reason to exist is a constraint in
someone else's API. The site is built as the surface it describes: near-black,
hairline-ruled, monospaced, dense without being cramped. Nothing on the page
decorates. Every line of chrome is either a border that separates two regions
or a rule that a terminal would also have.

The personality is plain-spoken and short — a brand commitment from the author,
not an aesthetic accident. This is why the type carries the tone: **JetBrains
Mono is the voice** and IBM Plex Sans is the annotation layer beneath it. A
heading set in mono reads like something typed at a prompt, which is exactly
what this product is. A sans-serif hero would read like a SaaS landing page,
and this is not one.

Density is medium-high and deliberate. The landing page is centred and
full-bleed — no left-aligned content column, because a single centred column
puts the product's one argument (the scope table) directly under the headline
where it cannot be skimmed past. The docs shell flips to a two-column
sidebar-plus-prose model, because at reading length the priority becomes
scannability and navigation, not persuasion.

**Anti-reference, confirmed by the incumbent code:** no glassmorphism (the nav
is deliberately opaque), no glow, no gradient text, no soft/rounded-pill
buttons, no centred-hero-with-three-feature-cards marketing grammar, no
drop-shadow under body copy.

**Key Characteristics:**

- Mono carries every heading, every label, every table number, and every code
  sample. Sans carries only running prose.
- Geometry is sharp: a 4–6px radius scale, never a pill.
- Hairlines at 6% and 11% white do all the separating. There are almost no
  filled panels.
- Colour is rare and load-bearing. The accent green appears on the thesis line,
  the active navigation item, the terminal prompt, and the primary button — and
  nowhere else at scale.
- Severity is always doubled by a word. Colour never carries a claim alone.
- Motion is transform/opacity only, entrance-only, and collapses entirely under
  `prefers-reduced-motion`.

## Colors

A near-black canvas carrying Apple's off-white (`#f5f5f7`) and Apple's grey
ramp, with one saturated green doing all the work.

### Primary

- **Signal Green** (`#30d158`): the single accent. It marks the thesis line of
  the hero (`Find it again later.`), every in-body link, the active sidebar
  item, the terminal prompt character, and the primary button background. Its
  region on the canvas is 10.39:1. It is also the only colour that appears
  *inside* running prose, via `a`.
- **Deep Ink** (`#04150a`): the near-black text colour used on top of Signal
  Green — primary button label and skip link. 9.30:1 against the accent, so the
  button is the highest-contrast pairing in the system.
- **Lift Green** (`#46e06d`): primary-button hover only. It is a hard-coded
  literal in the `.btn--primary:hover` rule, not a token.

### Secondary

- **Caveat Amber** (`#ffd60a`): "this exists but is partner-gated." Carried by
  the `.badge--mid` border/text, the `.dot--mid` marker in the scope table, and
  the `Careful` label on `.note.warn`. 14.88:1 on the canvas, 13.24:1 on
  `--surface`.
- **Closed Red** (`#ff6961`): "closed, or no API at any tier." The `.badge--no`
  and `.dot--no` marks, and the `Careful` tint on `.note.warn`. 7.45:1 on the
  canvas. It also tints the closed rows of the scope panel at 6% via
  `color-mix` — the one place colour carries a claim at region scale.

### Tertiary

- **Signal Blue** (`#2997ff`, CSS var `--blue`): **defined but currently
  unreferenced.** It is reserved, not in use. Do not reach for it to fill a gap
  without deciding that on purpose.

### Neutral

- **Canvas** (`#000000`, `--bg`): the page background.
- **Canvas Alt** (`#0a0a0c`, `--bg-alt`): the alternating band fill
  (`.section--alt`) and the nav background. One step off the canvas, enough to
  read as a different plane without a border.
- **Surface** (`#121215`): cards, `pre`, `.note`, and the terminal's border
  context.
- **Surface 2** (`#1a1a1f`): **defined but currently unreferenced.** Reserved.
- **Terminal Black** (`#0d0d10`, hard-coded in `Terminal.astro`): the terminal
  body's own fill — a hair cooler than `--surface` so the panel does not match
  the code blocks beside it.
- **Ink** (`#f5f5f7`, `--fg`): headings and primary text. 19.29:1 on the canvas.
- **Ink 2** (`#a1a1a6`, `--fg-2`): body copy, nav links, table cells, `.lead`.
  8.16:1 on the canvas, 7.27:1 on `--surface`.
- **Ink 3** (`#86868c`, `--fg-3`): the quietest tier — eyebrows, sidebar group
  titles, list markers, figcaptions, the footer, terminal output, the benchmark
  strip. **This is the floor of the type ramp.** 5.80:1 on the canvas, 5.47:1
  on `--bg-alt`, 5.17:1 on `--surface`, 4.79:1 on `--surface-2` — AA on all
  four.
- **Hairline** (`rgba(255, 255, 255, 0.11)`): interactive borders — cards at
  rest's hover target, ghost buttons, copy buttons, nav toggles.
- **Hairline 2** (`rgba(255, 255, 255, 0.06)`): structural rules — card borders
  at rest, `pre`, `.note`, table row rules, `hr`, the nav's bottom edge, the
  sidebar's right edge, prose `h2` top rules.

### Named Rules

**The One Green Rule.** Signal Green marks what is true and what works: the
thesis, the live link, the active item, the prompt, the primary action. It is
never used for decoration, never for a large filled region, and never for
severity. If a screen has more green on it than the eye can point at, something
is wrong.

**The Two Hairlines Rule.** There are exactly two border weights, 11% and 6%
white. 11% means "you can interact with this." 6% means "this is a region
boundary." Never introduce a third.

**The Word Before The Colour Rule.** No state is signalled by hue alone. The
scope table prints `self-serve` / `partner only` / `closed` next to every dot.
`.note.warn` prints the word `Careful`. Badges are bordered in `currentColor`
so their shape survives greyscale. This is a WCAG 1.4.1 commitment the author
asked for, not a stylistic one.

## Typography

**Display / Heading Font:** JetBrains Mono (fallback `ui-monospace,
SFMono-Regular, Menlo, monospace`)
**Body Font:** IBM Plex Sans (fallback `-apple-system, BlinkMacSystemFont,
'Segoe UI', sans-serif`)
**Label/Mono Font:** JetBrains Mono — the same face as display, not a third face.

Both are loaded from Google Fonts at weights 400, 500, 600 only. Nothing else
loads, which is why `strong`/`b` is explicitly set to 600 rather than left at
the UA default of 700 (700 would be synthesised and look wrong).

**Character:** A single monospaced voice with a quiet humanist narrator
underneath. The mono face is doing the persuading; the sans face is doing the
explaining. Any time both appear together, the mono is the louder of the two —
that inversion from the usual convention is the whole identity.

### Hierarchy

- **Display** (600, `clamp(1.75rem, .75rem + 4.6vw, 5rem)`, 1.06, `-.045em`):
  the hero `h1` only, and nowhere else. The clamp is *solved*, not eyeballed —
  JetBrains Mono advances 0.6em per glyph, so the longest line ("Find it again
  later.", 20 characters) needs 12em of room, and the clamp was fitted so that
  fits at every width down to ~410px. `max-width: 20ch` with `text-wrap:
  balance`; a `text-indent: -.0225em` optically re-centres the block against
  the trailing letter-space skew. **If you change the font size, re-solve this
  clamp — it is not decorative.**
- **Headline** (600, `clamp(1.6rem, 1.1rem + 2.2vw, 2.5rem)`, 1.14, `-.035em`):
  section band headings and the closing CTA band. One rule serves both
  (`.section__head h2, .cta-band h2`).
- **Title** — the prose scale, widest gap at the top: **2.5 / 1.5 / 1.2 / 1.0
  (body)**.
  - Prose `h1` (600, `clamp(1.7rem, 1.2rem + 2vw, 2.5rem)`, 1.2, `-.025em`)
  - Prose `h2` (600, `1.5rem`, 1.25, `-.015em`) — separated from the previous
    section by a hairline and `2rem` of padding-top
  - Prose `h3` (500, `1.2rem`, 1.35, `+.005em`)
  - Card `h3` (500, `1rem`, `+.01em`) — note the *positive* tracking. Mono set
    small needs to open up, not tighten.
- **Body** (400, `1rem`, 1.65, `+.005em`): running prose. The leading and
  tracking are both nudged up from a print-page default because light-on-dark
  reads smaller than light-on-light. Measure is `--measure` = **38rem (~74ch)**;
  40rem was measured at ~78ch, which overruns the 45–75ch band.
- **Label** (500, `.78rem`, `+.06em`, uppercase): sidebar group titles, `.eyebrow`
  and `.hero__eyebrow` (which prefix a literal `// ` in Signal Green), the
  `.strip` benchmark numbers, `.install__label`, and the `.note` severity labels.

Additional prose roles: `.lead` (1.2rem, `--fg-2`) opens a docs page.
`.hero__sub` is `clamp(1.125rem, .95rem + .9vw, 1.5rem)` at `--fg-2`, capped at
34ch.

Inline `code` is mono at `.85em`, `.01em` tracking, `rgba(255,255,255,.08)`
background, `5px` radius. Inside a `pre` it drops the background and runs at
`.85rem` / 1.7.

### Named Rules

**The Mono Voice Rule.** Every heading, every eyebrow, every label, every table
number (`.num` uses `font-variant-numeric: tabular-nums`), and every code
sample is JetBrains Mono. IBM Plex Sans is reserved for running prose, body
copy, and prose lists. Setting a heading in the sans face breaks the identity
more than any colour change would.

**The 16px Floor Rule.** `body` is `1rem` and the incumbent comments call it
explicitly: nothing below 16px on any viewport. Small text is made *quieter*
(smaller weight, more tracking, lower colour tier), never smaller in absolute
terms than the body role.

## Layout

Two shells, deliberately unlike each other.

### Landing shell

Full-bleed and centred throughout. There is no left-aligned content column
anywhere on the landing page — a centred stack puts the hero, the stage, the
benchmark strip, and the install commands in one vertical reading order, which
is what the argument needs.

- Nav: sticky, `3rem` tall, full-width (not page-constrained), background
  `--bg-alt`, bottom hairline 2.
- Hero: centred, padding `clamp(4.5rem, 12vh, 9rem)` block / `--gutter` inline,
  bottom `clamp(3.5rem, 8vh, 6rem)`.
- Stage (terminal + scope panel): single-column grid, `max-width: 68rem`,
  auto margins, top margin `clamp(2.5rem, 6vh, 4rem)`. **Stacked, not side by
  side**, and that is a measured decision, not a default: the scope table is
  ~545px of content and the terminal ~260px, so a two-column stage forced the
  terminal to stretch and stranded ~285px of empty frame below its last line.
  The grid gap is `1px` against a `--hairline-2` background, so the panels read
  as a single table with a hairline seam rather than two floating cards.
- Benchmark strip: centred flex-wrap, `.78rem` mono, `--fg-3`, max-width 44rem,
  gap `1.5rem 2.5rem`.
- Install block: centred, `max-width: 44rem`, `text-align: left` — the one
  left-aligned island on the page, because a command line must not be centred.
- Sections: `clamp(4.5rem, 10vh, 8rem)` block padding, alternating
  `--bg-alt` fill via `.section--alt`. Section head is centred, capped at
  40rem.
- Cards: `repeat(auto-fit, minmax(17rem, 1fr))`, gap `1.25rem`, `max-width:
  78rem`.
- Prose islands inside sections (used three times on the landing page) are
  `.prose` with an inline override of `max-width: 42rem`, `margin-inline: auto`,
  `padding-inline: 0`.

### Docs shell

- `.docs` is `display: flex; align-items: flex-start`.
- Sidebar: sticky at `top: var(--nav-h)`, `16.5rem` fixed, height
  `calc(100dvh - var(--nav-h))`, own scroll with `overscroll-behavior: contain`,
  padding `2rem 1rem 4rem var(--gutter)`, right hairline 2. Groups separated by
  `2rem`.
- Prose: `flex: 1; min-width: 0`, `max-width: calc(var(--measure) + 5rem)`,
  auto margins, padding `3.5rem var(--gutter) 7rem`. The `+ 5rem` gives the
  measure room to centre inside the remaining flex space.
- Nav inner is `max-width: 80rem` — wider than every content block, because the
  nav is chrome rather than content.

### Responsive

| Breakpoint | Change |
|---|---|
| `max-width: 63.99rem` | Nav links collapse behind the `.nav__toggle` hamburger, which opens a full-width stacked panel at `top: var(--nav-h)`. Docs flips to `flex-direction: column`; the sidebar goes `static`, `height: auto`, loses its right border and gains a bottom hairline 2, and is **hidden by default** on this width, toggled by `.nav__toggle--docs`. Prose drops `max-width` to none and its padding to `2.5rem var(--gutter) 4rem`. `.stage` remains single column. |
| `min-width: 64rem` | Docs sidebar toggle is hidden. `.scope` gains generous horizontal padding — `clamp(4rem, 14%, 9rem)` — so the panel reads full-bleed while its capability/status pairs stay adjacent. |
| `max-width: 30rem` | Hero top padding tightens to `3.5rem`; card padding to `1.5rem 1.25rem`. |

Tables are wrapped in `.table-scroll` (`overflow-x: auto`) rather than forcing
the page wide. `pre` is `overflow-x: auto` with `white-space: pre`, and the
`.codeblock` reserves `padding-right: 4.75rem` so a long line scrolls *under*
the copy button instead of colliding with it.

### Spacing rhythm

There is no spacing scale token. Spacing is written per-rule in `rem` and
clusters around a 4px base: `.25 .3 .35 .45 .5 .6 .7 .75 1 1.15 1.25 1.35 1.5
1.75 2 2.25 2.5 3.5 4 7`. The two exceptions that *are* tokens are `--gutter`
(responsive page inset) and `--measure`. Do not introduce a third.

## Elevation & Depth

**This system has no elevation.** There is no shadow vocabulary in
`global.css`. Depth is conveyed entirely by tonal layering — `#000` canvas,
`#0a0a0c` alternating band, `#121215` surface — plus hairlines at 6% and 11%
white. The single exception is `Terminal.astro`, which carries an inset
top highlight (`0 1px 0 rgba(255,255,255,.04) inset`) and one large soft drop
(`0 24px 60px -30px rgba(0,0,0,.9)`) because a simulated OS window needs to
read as a window. That shadow is not part of the site's elevation model.

### Named Rules

**The Flat-By-Default Rule.** Surfaces are flat at rest, permanently. Nothing
in this design lifts off the canvas. If a new component wants a shadow, the
question to ask is whether it is a window (which earns one) or a panel (which
does not).

## Shapes

Sharp, near-rectangular geometry. The radius scale is **4px / 6px / 8px**
(`--r-sm` / `--r-md` / `--r-lg`), with three hard-coded exceptions in active
use: `5px` for inline `code`, `3px` for `.badge`, and `50%` for the 7px status
dots and the terminal's traffic-light dots.

- `--r-sm` (4px): buttons, nav links, `pre`, `.copy`, `.note`, `.skip-link`,
  nav toggles, sidebar links, focus rings.
- `--r-md` (6px): cards and the stage container.
- `--r-lg` (8px): **defined but currently unreferenced.**
- Pills and fully-rounded corners do not exist in this system except on the
  7px status dots, where `50%` is the point.

`:focus-visible` sets `border-radius: 4px` on the outline box so the focus
ring is not sharper than the control it wraps.

Borders are 1px throughout. Nothing is 2px except `.hero__eyebrow::before` and
`.scope__wave::before`, which are 1px rules, and the graph-paper grid, whose
lines are 1px at 3.2% opacity.

## Components

### Buttons

Flat, sharp, and confident. There is no shadow, no gradient, and no lift on
hover — state is carried by background and border colour, plus a 3% scale-down
on press.

- **Shape:** `--r-sm` (4px), `1px solid transparent`, `min-height: 2.75rem`,
  `padding: 0 1.4rem`, `1rem` / 500 / `-.01em`.
- **Primary:** Signal Green background, Deep Ink text. Hover lifts the green to
  `#46e06d`. `text-decoration: none` on hover.
- **Ghost:** transparent background, `--fg` text, `1px solid var(--hairline)`.
  Hover raises the border to `rgba(255,255,255,.2)` and fills at 6% white.
- **States:** `:active` applies `transform: scale(.97)`. Transition is
  `background .2s ease, color .2s ease, border-color .2s ease, transform .12s
  ease` — transform is faster than colour, so the press feels immediate.

### Nav

Sticky, opaque, `3rem`. Full-bleed with an `80rem` inner. Brand is `1rem` / 600
with an inline accent-coloured SVG mark. Links are `.82rem` / 400 in `--fg-2`;
hover raises colour to `--fg` and fills 7% white; `[aria-current='page']` uses
`--fg` on 10% white. Two toggles exist below 64rem — the docs menu toggle
(sits left of the primary) and the primary nav toggle — both `2.25rem` square
with a hairline border and an `sr-only` label.

**Deliberately no `backdrop-filter`.** A backdrop-filter on a sticky
full-width bar is re-rasterised on every scroll frame, and there is nothing
behind this bar worth blurring: the graph-paper grid is masked out and the
canvas is `#000`.

### Cards

`1px solid var(--hairline-2)`, `--r-md`, `--surface` background, `1.75rem
1.5rem` padding, centred. Hover raises the border to `--hairline` and lifts
`-2px` (`transform` + `border-color`, `.25s ease`). No shadow at rest or on
hover.

### Code blocks

`pre` on `--surface` with a `--hairline-2` border, `--r-sm`, `1.1rem 1.25rem`
padding. Copy button is absolutely positioned `top/right: .7rem`,
`min-width: 2.5rem; min-height: 2.25rem`, 6% white fill, hairline border. On
`[data-state='done']` both its colour and border go Signal Green. It has an
`sr-only` "Copy code to clipboard" in addition to its visible label, and falls
back to printing `Ctrl+C` if the clipboard API rejects.

### Notes

`.note` is `--surface` with a hairline-2 border, `--r-sm`, `1rem 1.25rem`
padding. `.note.warn` tints the border 35% toward `--warn`; `.note.info` tints
it 30% toward `--accent`. **Both variants inject a `::before` severity label on
their first paragraph** — `Careful` in amber, `Good to know` in green, both
`.74rem` mono uppercase with `.06em` tracking — so severity is never carried
by border tint alone. Unlabelled `.note` stays neutral.

### Badges

Inline, `3px` radius, `.78rem` mono / 500, `+.03em` tracking, `1px solid
currentColor` so the border survives greyscale, `white-space: nowrap`.
`.badge.ok` accent, `.badge.no` danger, `.badge.mid` warn.

### The Stage (signature component)

The hero's terminal-plus-scope-panel grid, `68rem` max, `1px` gap over a
hairline-2 background, `--r-md`, `overflow: hidden`. Its children are filled
`--bg` so the gap reads as a seam. Two things make it the signature:

1. **The scope panel *is* the argument.** It is not a feature list. It states,
   in nine rows grouped into three labelled tiers, what postiee cannot do —
   comments, analytics, reading posts back, DMs, invitations, ads — and why.
   The copy above it says "most of it is red," so the red is shown rather than
   merely described.
2. **Colour at region scale, exactly once.** `.scope__row:has(.dot--no)` gets a
   6% `--danger` `color-mix` tint. Every other row is untinted. Tier labels are
   `--fg-3` uppercase with a short leading rule; closed rows drop their
   hairline so the three tiers read as three blocks.

Inside the panel the scope rows animate in **three beats, not nine**: tier
labels at `tier × 0.5s`, then a `.07s` base and `.05s` stagger within each tier.
This is why the panel reads as three arguments rather than a list.

### The Terminal (signature component)

`#0d0d10` fill, `--r-md`, hairline-2 border, and the one authored shadow in the
system. Traffic-light dots (`#ff5f57`, `#febc2e`, `#28c840`) are hard-coded
literals in the markup. Body is `.82rem` mono / 1.75 with an accent `$` prompt
and a 8px blinking caret. Lines rise in sequence (delays `.35s → 3.1s`) while
command text types in at 22–52ms per character. **The caret stops blinking the
instant the last command lands** (`data-quiet`) — the post ID is on disk, there
is nothing left to type, and the page stops moving. It carries `role="img"`
with a full text `aria-label`, and its markup holds the final text so the
content is readable with JS disabled.

### Scroll reveal

Content is visible by default. JS adds the `js` class to `<html>` and only then
does `.reveal` get `opacity: 0` — so a failed script can never blank the page.
An `IntersectionObserver` (`rootMargin: '0px 0px -12% 0px'`) adds `.in` and
unobserves. Cards cascade by `:nth-child` delay (`.05 .12 .19 .26s`) instead of
snapping in as one block.

### Diagram components

`PostLifecycle.astro` and `ScopeGate.astro` are inline SVG diagrams. **They do
not use the token palette.** They carry GitHub's primitives — `#e6edf3`,
`#a8b0bb`, `#98a0ab`, `#d29922`, `#3fb950`, `#f85149` — at 11–14px with
`ui-monospace, monospace`. These are in-figure colours and are deliberately
outside the site ramp; do not "fix" them to match the tokens without deciding
that on purpose.

## Motion

Everything animates on **transform and opacity only**. Nothing animates colour,
width, height, or position — except the terminal caret, which animates
`visibility` via `steps(2, start)`.

### Easing

| Token in use | Value | Where |
|---|---|---|
| Standard | `ease` | nav link, button, card, copy-button transitions |
| **Entrance** | `cubic-bezier(.16, 1, .3, 1)` | hero sequence, scroll reveals, scope panel tiers |
| Terminal rise | `cubic-bezier(.2, .7, .3, 1)` | terminal line entrance |
| Step | `steps(2, start)` | terminal caret blink |

The entrance curve is a long decelerating tail — fast start, slow settle. It is
used for everything that appears on the page.

### Durations

| Value | Applied to |
|---|---|
| `.12s` | button press transform |
| `.15s` | sidebar link colour/background |
| `.2s` | button, nav link, copy button transitions |
| `.25s` | card hover border + transform |
| `.45s` | scope panel tier rows |
| `.5s` | terminal line rise |
| `.7s` | scroll reveal opacity + transform |
| `.8s` | hero entrance |
| `1.1s` | terminal caret blink (infinite) |

### What animates, in order

1. **Hero entrance** — a sequence down the stack, `backwards` fill so nothing
   flashes before its delay: eyebrow `.05s` → h1 `.15s` → sub `.28s` → CTA
   `.40s` → stage `.55s`. Each rises 14px.
2. **Terminal lines** — rise in sequence from `.35s` to `3.1s`, with typed
   command text.
3. **Scope panel** — three tier beats, then a within-tier stagger.
4. **Scroll reveals** — driven by the observer, not by load time, so a section
   animates when you actually reach it. Cards cascade.
5. **Caret blink** — infinite until the last command lands, then stopped.

### The reduced-motion contract

Non-negotiable, per the author. Implemented in three places, all in
`global.css` and the two animated components:

```css
@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  *, *::before, *::after {
    animation-duration: .01ms !important;
    transition-duration: .01ms !important;
  }
  .reveal { opacity: 1 !important; transform: none !important; }
}
```

`Terminal.astro` and `ScopePanel.astro` each also carry their own
`prefers-reduced-motion` block that sets elements to `opacity: 1` and
`animation: none` explicitly — belt and braces, because those are scoped styles
with `opacity: 0` defaults. The terminal's script checks the media query
directly and fills the command spans immediately instead of typing them, then
suppresses the caret. **The contract is: with reduced motion, the page renders
fully settled and complete, nothing moves, and no content is blank or missing.**

The old `body { opacity: 0; animation: fade }` font-load trick has been removed:
it blanked the page for anyone whose script never ran.

## Accessibility

The author's stated requirement, verified against the CSS:

### Contrast — computed from the token values, not estimated

| Foreground | on `--bg` | on `--bg-alt` | on `--surface` | on `--surface-2` |
|---|---|---|---|---|
| `--fg` `#f5f5f7` | **19.29:1** | 18.17:1 | 17.17:1 | 15.92:1 |
| `--fg-2` `#a1a1a6` | **8.16:1** | 7.69:1 | 7.27:1 | 6.74:1 |
| `--fg-3` `#86868c` | **5.80:1** | 5.47:1 | 5.17:1 | 4.79:1 |

Accents on the canvas: `--accent` **10.39:1**, `--accent-dim` 8.12:1,
`--signal-blue` 6.96:1, `--warn` 14.88:1, `--danger` 7.45:1,
`--accent-hover` 12.14:1. On `--surface`: accent 9.25:1, warn 13.24:1,
danger 6.63:1. `--accent-ink` on `--accent` (the primary button):
**9.30:1**.

Everything clears WCAG AA (4.5:1 for body, 3:1 for large text) on all four
backgrounds, including the quietest tier. `--fg-3` at 4.79:1 on `--surface-2`
is the tightest margin in the system.

### The body-text floor

`body` is `1rem`. The incumbent comment marks it explicitly as the 16px floor,
honouring the author's requirement that body text never drops below 16px on
mobile. Only genuinely non-prose text (eyebrows, labels, badges, figcaptions,
terminal chrome) goes below, and it always steps *up* in colour tier or
letter-spacing to compensate.

### Focus

Never removed, only made consistent. A single global rule:

```css
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 3px;
  border-radius: 4px;
}
```

The accent outline clears 3:1 against every background in the system by a wide
margin (10.39:1 on the canvas). The `border-radius` on the outline box keeps the
ring from being sharper than the control it wraps.

### Skip link

`<a class="skip-link" href="#main">Skip to content</a>` is the first element in
`<body>`. It sits at `left: -9999px` off-screen and moves to `left: 0` on
`:focus`, styled Signal Green on Deep Ink (9.30:1), `padding: .75rem 1.25rem`,
`font-weight: 600`, with a `--r-sm` bottom-right radius only — it reads as a
tab hanging off the corner. It targets `<main id="main">`.

### Touch targets

`.btn` has `min-height: 2.75rem` (44px). `.copy` has `min-height: 2.25rem`
(36px) with `min-width: 2.5rem`. `.nav__toggle` is `2.25rem` square. These meet
the 24px WCAG 2.2 minimum; the primary actions meet the 44px advisory target.

### Colour-independence

Covered by the Word Before The Colour Rule above. Additionally: status dots are
`aria-hidden` with the meaning carried by the adjacent text; `.badge` borders
use `currentColor` so shape survives greyscale; the terminal is `role="img"`
with a complete `aria-label`; both nav toggles carry `sr-only` labels plus
`aria-expanded` and `aria-controls`.

### Semantic structure

One `<h1>` per page, `<nav>` labelled `Main` and `Documentation`, `<main
id="main">`, `<footer>`, `<table>` with `<th>` and `font-weight: 500`,
`aria-current="page"` on both nav and sidebar links. `.sr-only` is available as
a utility class.

## Do's and Don'ts

### Do:

- **Do** set every heading, eyebrow, label, table number and code sample in
  JetBrains Mono, and running prose in IBM Plex Sans.
- **Do** check any new colour against the full four-background contrast table
  before shipping it, and treat 4.5:1 as the floor.
- **Do** pair every new coloured state with a word or a shape, so it survives
  greyscale and every form of colour vision.
- **Do** keep geometry inside the 4 / 6 / 8px radius scale, and reach for
  `50%` only on the 7px status dots and terminal traffic lights.
- **Do** animate transform and opacity only, and mirror the entrance curve
  `cubic-bezier(.16, 1, .3, 1)` for anything that appears on the page.
- **Do** verify new motion collapses cleanly under
  `prefers-reduced-motion: reduce`, including any scoped component that starts
  elements at `opacity: 0`.
- **Do** re-solve the hero `h1` clamp if the mono font size or the headline
  copy changes. The current value fits "Find it again later." down to ~410px.
- **Do** let content render fully settled and complete with JS disabled — the
  scroll-reveal hidden state is applied by the `js` class, never by default.
- **Do** use Signal Green for what is true and what works, and keep it rare.

### Don't:

- **Don't** reintroduce a `backdrop-filter` on the nav, or any glassmorphism.
  It is re-rasterised on every scroll frame and there is nothing behind the bar
  worth blurring.
- **Don't** add a shadow. Surfaces are flat at rest, permanently. The terminal
  is the only exception and it is a window, not a panel.
- **Don't** set body text below 16px, or introduce a third hairline weight
  alongside 6% and 11% white.
- **Don't** make the landing page a left-aligned column. It is centred
  full-bleed so the scope table sits directly under the headline.
- **Don't** side-by-side the stage on desktop. The stacked layout is a measured
  fix for ~285px of stranded empty frame in the two-column version.
- **Don't** signal severity with hue alone. Every `.note` variant prints its
  severity word; every scope row prints its status.
- **Don't** build a hero on the centre-heading-three-sub-cards marketing
  grammar. The scope table is the argument and it belongs in the hero.
- **Don't** promote the benchmark strip to a headline stat. It is deliberately
  a quiet mono line at `--fg-3`, and the numbers must always be numbers a user
  can reproduce with `go test -bench=. -benchmem ./bench/`.
- **Don't** animate width, height, colour, or position. Transform and opacity
  only.
- **Don't** use the reserved tokens (`--signal-blue`, `--surface-2`,
  `--r-lg`, `--accent-dim`) without deciding that on purpose — they are
  currently unreferenced by any rule.

---

## Known inconsistencies and in-progress work

Recorded so a later pass does not mistake them for settled decisions.

1. **The token comment at the top of `global.css` is stale.** Its summary block
   claims `--fg-3` is 4.92:1 on `--bg` and *fails* AA on `--surface` (4.38:1)
   and `--surface-2` (4.06:1), telling future authors never to set `--fg-3` on a
   surface. The inline comment on the `--fg-3` declaration itself gives the
   correct figures for the current value `#86868c`: 5.8 / 5.2 / 4.8. Recomputing
   from the actual hex gives **5.80 / 5.17 / 4.79** — the inline comment is
   right, the summary block is describing the *old* `--fg-3` value `#7a7a80` and
   was not updated when the token was lightened. The rule "never put `--fg-3`
   on a surface" is now unnecessary; the section above reflects the corrected
   state.
2. **Four tokens are defined but unreferenced:** `--surface-2` (`#1a1a1f`),
   `--accent-dim` (`#28b947`), `--blue` (`#2997ff`), `--r-lg` (`8px`).
   Documented as reserved rather than removed, since removing them was out of
   scope for this pass.
3. **`--accent-dim` has no job.** Its likely intent — a pressed/active state for
   the primary button — was implemented instead as `transform: scale(.97)` on
   `:active` with a `#46e06d` hover literal.
4. **The diagram SVGs are off-palette.** `PostLifecycle.astro` and
   `ScopeGate.astro` use GitHub's grey/amber/green/red primitives rather than
   the site ramp. Recorded as a deliberate-looking exception; unverified either
   way.
5. **Values in this file may drift.** This document was generated from a
   working tree where `global.css`, `index.astro`, `ScopePanel.astro`,
   `Terminal.astro` and the docs pages were all being edited in parallel. If a
   value here disagrees with the CSS, the CSS wins and this file is the thing
   that is out of date.
