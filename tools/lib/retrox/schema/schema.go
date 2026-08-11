// Package schema defines the Go representation of every Retro-X document.
// The structs mirror RETROX.md one to one and are the single source of truth
// for the format: exporters marshal these types, the validator checks them,
// and no tool hand-writes Retro-X JSON.
package schema

import (
	"encoding/json"
	"fmt"
)

// FormatName is the literal value of every document's "format" field.
const FormatName = "retro-x"

// Version is the Retro-X version this package implements. It is bumped only
// for compatibility-breaking changes; additive optional fields do not bump it.
const Version = 1

// Header opens every Retro-X JSON document.
type Header struct {
	Format  string `json:"format"`
	Version int    `json:"version"`
}

// NewHeader returns the header every emitted document carries.
func NewHeader() Header { return Header{Format: FormatName, Version: Version} }

// Check rejects documents that are not Retro-X or are newer than this package.
func (h Header) Check() error {
	if h.Format != FormatName {
		return fmt.Errorf("format %q is not %q", h.Format, FormatName)
	}
	if h.Version < 1 || h.Version > Version {
		return fmt.Errorf("version %d unsupported (this implementation speaks up to %d)", h.Version, Version)
	}
	return nil
}

// Index is the root index.json listing the games of a publication.
type Index struct {
	Header
	Games []string `json:"games"`
}

// ---------------------------------------------------------------------------
// manifest.json

type Manifest struct {
	Header
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Platform    string    `json:"platform,omitempty"`
	Year        int       `json:"year,omitempty"`
	Description string    `json:"description,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Display     Display   `json:"display"`
	Docs        []DocPage `json:"docs,omitempty"`
	Assets      []Asset   `json:"assets"`
}

type Display struct {
	Native Size   `json:"native"`
	TickHz int    `json:"tickHz"`
	Filter string `json:"filter,omitempty"`
	// TexFilter is how the platform's rasterizer sampled textures:
	// "linear" (N64-style bilinear with mipmaps), "bilinear" (bilinear with
	// NO mipmaps — for games whose textures ship without mip chains, where
	// hardware always reads mip 0; generated mips band fine gradients), or
	// "nearest" (PSX-style point sampling). Empty leaves the viewer's
	// default behaviour.
	TexFilter string `json:"texFilter,omitempty"`
}

type Size struct {
	W int `json:"w"`
	H int `json:"h"`
}

type DocPage struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	File  string `json:"file"`
}

// Asset categories (closed set).
const (
	CategoryLevel   = "level"
	CategoryObject  = "object"
	CategoryMusic   = "music"
	CategorySFX     = "sfx"
	CategoryPicture = "picture"
	CategoryVideo   = "video"
)

// Categories lists the closed set of asset categories.
var Categories = []string{
	CategoryLevel, CategoryObject, CategoryMusic, CategorySFX, CategoryPicture, CategoryVideo,
}

// Asset is one entry of manifest.assets. Fields beyond the common set apply
// only to particular categories (see RETROX.md §4).
type Asset struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Name        string   `json:"name"`
	Group       string   `json:"group,omitempty"`
	Description string   `json:"description,omitempty"`
	Related     []string `json:"related,omitempty"`
	File        string   `json:"file"`

	// music
	Loop bool `json:"loop,omitempty"`
	// music / video: seconds
	Duration float64 `json:"duration,omitempty"`
	// picture / video: source pixel dimensions
	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`
	// video
	FPS float64 `json:"fps,omitempty"`

	// free-form key→value rows for the viewer's info panel; for pictures
	// this is where the platform truth lives (native resolution, colour
	// depth) when the exported PNG is a baked upscale
	Stats map[string]any `json:"stats,omitempty"`
}

// ---------------------------------------------------------------------------
// Level documents

const (
	LevelTilemap = "tilemap"
	LevelScene3D = "scene3d"
)

type Level struct {
	Header
	Type       string      `json:"type"`
	Music      string      `json:"music,omitempty"`
	PixelGrid  *PixelGrid  `json:"pixelGrid,omitempty"`
	Camera     *Camera     `json:"camera,omitempty"`
	Variants   []Variant   `json:"variants,omitempty"`
	Tilemap    *Tilemap    `json:"tilemap,omitempty"`
	Scene      *Scene      `json:"scene,omitempty"`
	Collision  *Collision  `json:"collision,omitempty"`
	Placements []Placement `json:"placements,omitempty"`
	Pools      []Pool      `json:"pools,omitempty"`
	Routes     []Route     `json:"routes,omitempty"`
	Scripts    []ScriptRef `json:"scripts,omitempty"`
}

type PixelGrid struct {
	Lines int `json:"lines"`
}

type Variant struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default,omitempty"`
}

type ScriptRef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description,omitempty"`
}

// --- tilemap body ---

type Tilemap struct {
	TileSize  int        `json:"tileSize"`
	Width     int        `json:"width"`
	Height    int        `json:"height"`
	Atlas     TileAtlas  `json:"atlas"`
	Cells     []int      `json:"cells"`
	HFlipMask int        `json:"hflipMask,omitempty"`
	Blocks    *Blocks    `json:"blocks,omitempty"`
	Wrap      string     `json:"wrap,omitempty"` // "" | "none" | "x"
	View      *Rect      `json:"view,omitempty"`
	Spawn     *Spawn     `json:"spawn,omitempty"`
	TileAnims []TileAnim `json:"tileAnims,omitempty"`
	CellAnims []CellAnim `json:"cellAnims,omitempty"`
	PaletteFx *PaletteFx `json:"paletteFx,omitempty"`
	Layers    []MapLayer `json:"layers,omitempty"`
}

// MapLayer is a named group of a tilemap level's placements, so a viewer can
// show and hide them independently (Placement.Layer names one). The first
// declared layer is the default for placements that name none.
type MapLayer struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Visible *bool  `json:"visible,omitempty"`
}

// IsVisible resolves the initial-visibility default (true).
func (l *MapLayer) IsVisible() bool { return l.Visible == nil || *l.Visible }

type TileAtlas struct {
	File   string `json:"file"`
	Cols   int    `json:"cols"`
	Gutter int    `json:"gutter,omitempty"`
}

type Blocks struct {
	Size   int     `json:"size"`
	Tiles  [][]int `json:"tiles"`
	Shapes []int   `json:"shapes,omitempty"`
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Spawn struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Object string `json:"object,omitempty"`
	Anim   string `json:"anim,omitempty"`
	Tint   string `json:"tint,omitempty"`
}

type TileAnim struct {
	Tiles        []int   `json:"tiles"`
	Frames       [][]int `json:"frames"`
	PeriodFrames int     `json:"periodFrames"`
}

type CellAnim struct {
	TX     int         `json:"tx"`
	TY     int         `json:"ty"`
	TW     int         `json:"tw"`
	TH     int         `json:"th"`
	Phases []CellPhase `json:"phases"`
}

type CellPhase struct {
	Tiles  []int `json:"tiles"`
	Frames int   `json:"frames"`
}

type PaletteFx struct {
	Palette []string        `json:"palette,omitempty"`
	Cycle   *PaletteCycle   `json:"cycle,omitempty"`
	Regions []PaletteRegion `json:"regions,omitempty"`
}

type PaletteCycle struct {
	Slots        []int      `json:"slots"` // palette indices that cycle
	Steps        [][]string `json:"steps"` // per step: one colour per slot
	PeriodFrames int        `json:"periodFrames"`
	Tiles        []int      `json:"tiles,omitempty"` // tile ids that use a cycling slot
}

// PaletteRegion is a rectangular area drawn with an alternate palette (a
// raster split on the original hardware — an underwater zone, a cave veil).
type PaletteRegion struct {
	Name    string   `json:"name,omitempty"`
	Rect    Rect     `json:"rect"` // world pixels
	Palette []string `json:"palette"`
}

// --- scene3d body ---

type Scene struct {
	Background string  `json:"background,omitempty"`
	Fog        *Fog    `json:"fog,omitempty"`
	Layers     []Layer `json:"layers"`
	Rooms      *Rooms  `json:"rooms,omitempty"`
	// Lights the scene is lit by, when the game's own light rig is known. A
	// viewer that gets these must use them *instead of* its default rig, not on
	// top of it. Absent, the viewer lights the scene however it likes — which
	// is right for content exported unlit.
	Lights []Light `json:"lights,omitempty"`
}

type Fog struct {
	Color string  `json:"color"`
	Near  float64 `json:"near"`
	Far   float64 `json:"far"`
}

// Layer modes.
const (
	LayerBase   = "base"
	LayerToggle = "toggle"
	// Exclusive layers use "exclusive:<group>".
)

type Layer struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	// File is the layer's GLB payload. It may be omitted, and then the layer is
	// a pure PLACEMENT GROUP: it carries no geometry of its own and exists to
	// give the placements naming it a shared toggle (the same idea the tilemap
	// body's `layers` list has). A file-less layer must be named by at least
	// one placement.
	File          string  `json:"file,omitempty"`
	Mode          string  `json:"mode,omitempty"` // default "base"
	Visible       *bool   `json:"visible,omitempty"`
	Attach        string  `json:"attach,omitempty"` // "" == "world" | "camera" | "cameraYaw"
	RenderOrder   float64 `json:"renderOrder,omitempty"`
	Transparent   bool    `json:"transparent,omitempty"`
	DepthTest     *bool   `json:"depthTest,omitempty"`
	PolygonOffset float64 `json:"polygonOffset,omitempty"`
	// Role "shadow" is geometry that is never drawn: it exists only to cast
	// the scene's shadows, which is what the guest's own depth-shadow proxy
	// models are for.
	Role   string   `json:"role,omitempty"`   // "collision" | "sky" | "water" | "shadow"
	EnvMap []string `json:"envMap,omitempty"` // 6 cube faces (+x,-x,+y,-y,+z,-z) for sheen-marked materials
}

// IsVisible resolves the initial-visibility default (true).
func (l *Layer) IsVisible() bool { return l.Visible == nil || *l.Visible }

// DepthTests resolves the depth-test default (true).
func (l *Layer) DepthTests() bool { return l.DepthTest == nil || *l.DepthTest }

type Rooms struct {
	Areas  []Area `json:"areas,omitempty"`
	Stream bool   `json:"stream,omitempty"`
	List   []Room `json:"list"`
}

// Area is a named spatial group of rooms (a storey, a wing) driving the
// viewer's peel toggles. Rooms reference it by id, like every cross-reference.
type Area struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Room struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
	File string `json:"file"`
	Area string `json:"area,omitempty"`
	AABB *AABB  `json:"aabb,omitempty"`
}

type AABB struct {
	Min [3]float64 `json:"min"`
	Max [3]float64 `json:"max"`
}

// --- 2-D collision ---

type Collision struct {
	Kind   string            `json:"kind"` // "grid" | "profiles"
	Sub    int               `json:"sub,omitempty"`
	Solid  []int             `json:"solid,omitempty"`
	Legend map[string]string `json:"legend,omitempty"`
	File   string            `json:"file,omitempty"`
}

// Shapes is the "profiles" collision side-car document.
type Shapes struct {
	Header
	Count    int       `json:"count"`
	Profiles [][]int   `json:"profiles"`
	Angles   []float64 `json:"angles"`
}

// --- placements ---

type Placement struct {
	ID        int            `json:"id"`
	Object    string         `json:"object"`
	Pos       []float64      `json:"pos,omitempty"`
	Rot       []float64      `json:"rot,omitempty"`
	Scale     Scale          `json:"scale,omitempty"`
	Matrix    []float64      `json:"matrix,omitempty"`
	Anim      string         `json:"anim,omitempty"`
	HFlip     bool           `json:"hflip,omitempty"`
	Tint      string         `json:"tint,omitempty"`
	Hard      bool           `json:"hard,omitempty"`
	Layer     string         `json:"layer,omitempty"`
	Room      *int           `json:"room,omitempty"`
	Variants  []string       `json:"variants,omitempty"`
	Collision *ObjCollision  `json:"collision,omitempty"`
	Route     *RouteRef      `json:"route,omitempty"`
	Behavior  *Behavior      `json:"behavior,omitempty"`
	OnClick   *OnClick       `json:"onClick,omitempty"`
	Name      string         `json:"name,omitempty"`
	Info      *Info          `json:"info,omitempty"`
	Props     map[string]any `json:"props,omitempty"`
}

// Scale marshals as a scalar when uniform ([s]) and as [sx,sy,sz] otherwise.
type Scale []float64

func (s Scale) MarshalJSON() ([]byte, error) {
	switch len(s) {
	case 0:
		return []byte("null"), nil
	case 1:
		return json.Marshal(s[0])
	default:
		return json.Marshal([]float64(s))
	}
}

func (s *Scale) UnmarshalJSON(b []byte) error {
	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		*s = Scale{f}
		return nil
	}
	var v []float64
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("scale must be a number or an array: %w", err)
	}
	*s = Scale(v)
	return nil
}

type ObjCollision struct {
	File   string    `json:"file"`
	Matrix []float64 `json:"matrix,omitempty"` // 12 floats, rows of a 3×4
}

type RouteRef struct {
	ID    string  `json:"id"`
	Speed float64 `json:"speed"`
	Mode  string  `json:"mode,omitempty"` // "loop" (default) | "pingpong"
	Face  bool    `json:"face,omitempty"`
}

// Behaviour kinds (closed vocabulary for version 1; unknown kinds are ignored
// by viewers).
const (
	BehaviorSpin  = "spin"
	BehaviorFlyer = "flyer"
)

type Behavior struct {
	Kind string `json:"kind"`
	// spin
	Axis []float64 `json:"axis,omitempty"`
	Rate float64   `json:"rate,omitempty"` // rad/s
	// flyer
	Keys     []FlyKey  `json:"keys,omitempty"`
	Loop     bool      `json:"loop,omitempty"`
	SpinPart *SpinPart `json:"spinPart,omitempty"`
}

type FlyKey struct {
	Pos  []float64 `json:"pos"`
	Dur  int       `json:"dur"`            // engine frames
	Hold int       `json:"hold,omitempty"` // engine frames
	Yaw  float64   `json:"yaw,omitempty"`
}

type SpinPart struct {
	Node string    `json:"node"`
	Axis []float64 `json:"axis"`
	Rate float64   `json:"rate"`
}

// OnClick actions (extensible; unknown actions are no-ops in viewers).
const (
	ActionAnimate = "animate"
	ActionText    = "text"
)

type OnClick struct {
	Action string `json:"action"`
	// animate
	Target string `json:"target,omitempty"` // "" == "self", or a placement id (decimal string)
	// With names placements that move TOGETHER with this one: clicking either
	// runs both. Each partner plays its OWN onClick — its own clip, hold and
	// toggle — because a pair need not share one animation: a double door's two
	// leaves swing on mirrored clips. That is the difference from Target, which
	// plays THIS placement's clip somewhere else.
	//
	// One thing in the world is often several placements (a double door, a star
	// gate's two halves, the castle trapdoor's two leaves), and a click that
	// moves half of it reads as a bug. Pairs list each other, so whichever half
	// is clicked, both move. A viewer that does not implement it simply keeps
	// the old per-placement behaviour.
	With   []string `json:"with,omitempty"`
	Clip   string   `json:"clip,omitempty"`
	HoldAt float64  `json:"holdAt,omitempty"` // normalized clip time
	Toggle bool     `json:"toggle,omitempty"`
	SFX    []SFXCue `json:"sfx,omitempty"`
	// text
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

type SFXCue struct {
	ID string  `json:"id"`
	At float64 `json:"at,omitempty"` // seconds into the clip
}

type Info struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Quote string `json:"quote,omitempty"`
}

type Pool struct {
	ID         string      `json:"id"`
	Count      int         `json:"count"`
	Object     string      `json:"object"`
	Candidates [][]float64 `json:"candidates"`
	Seedable   bool        `json:"seedable,omitempty"`
	Anim       string      `json:"anim,omitempty"`
	Tint       string      `json:"tint,omitempty"`
	Name       string      `json:"name,omitempty"` // instance display name
	Info       *Info       `json:"info,omitempty"` // instance info-card text
	Variants   []string    `json:"variants,omitempty"`
}

type Route struct {
	ID     string      `json:"id"`
	Loop   bool        `json:"loop,omitempty"`
	Points [][]float64 `json:"points"`
}

// ---------------------------------------------------------------------------
// Object documents

const (
	ObjectSprite2D    = "sprite2d"
	ObjectModel3D     = "model3d"
	ObjectBillboard3D = "billboard3d"
	ObjectWireframe3D = "wireframe3d"
)

type Object struct {
	Header
	Type string `json:"type"`
	Name string `json:"name"`

	// sprite2d / billboard3d
	Atlas      *SpriteAtlas `json:"atlas,omitempty"`
	Animations []Animation  `json:"animations,omitempty"`

	// model3d
	Model        string         `json:"model,omitempty"`
	Variants     []ModelVariant `json:"variants,omitempty"` // independent alternates, one glTF scene each
	EnvMap       []string       `json:"envMap,omitempty"`   // 6 cube faces (+x,-x,+y,-y,+z,-z) for sheen-marked materials
	Instanced    bool           `json:"instanced,omitempty"`
	SkinnedClone bool           `json:"skinnedClone,omitempty"`
	Billboard    string         `json:"billboard,omitempty"` // "yaw"
	UVAnims      []UVAnim       `json:"uvAnims,omitempty"`
	Flipbooks    []Flipbook     `json:"flipbooks,omitempty"`
	AtlasPicture string         `json:"atlasPicture,omitempty"`

	// billboard3d
	Views      int       `json:"views,omitempty"`
	Heading    float64   `json:"heading,omitempty"`
	Mode       string    `json:"mode,omitempty"`       // "camera" | "yaw"
	Size       []float64 `json:"size,omitempty"`       // world-unit quad [w,h]
	AnchorMode string    `json:"anchorMode,omitempty"` // "center" (default) | "bottom"
	Blend      string    `json:"blend,omitempty"`      // "opaque" (default) | "alpha" | "additive"

	// wireframe3d
	Wireframe *Wireframe `json:"wireframe,omitempty"`

	Stats map[string]any `json:"stats,omitempty"`
	Props map[string]any `json:"props,omitempty"`
}

// ModelVariant names one independent alternate of a model3d — an extra glTF
// scene in the same GLB (LOD levels, a shadow-caster proxy, livery recolours).
// The first entry is the default and must name scene 0; the viewer offers the
// rest without re-fetching the model. (Distinct from Variant, the level-level
// variant list.)
type ModelVariant struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Scene       string `json:"scene"` // glTF scene name in the model GLB
	Description string `json:"description,omitempty"`
}

type SpriteAtlas struct {
	File   string `json:"file"`
	CellW  int    `json:"cellW"`
	CellH  int    `json:"cellH"`
	Anchor []int  `json:"anchor,omitempty"` // [ax, ay] within a cell
}

// Animation covers all object types; the applicable fields depend on the type
// (sprite2d: row/frames/durations|steps/path; model3d: clip/fps;
// billboard3d: col/framesPerView/fps).
type Animation struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Loop        string `json:"loop"` // "once" | "loop" | "pingpong" | "hold"
	Description string `json:"description,omitempty"`

	// sprite2d
	Row       int         `json:"row,omitempty"`
	Frames    int         `json:"frames,omitempty"`
	Durations []int       `json:"durations,omitempty"`
	Steps     [][]int     `json:"steps,omitempty"` // [frameIndex, hold]
	Path      [][]int     `json:"path,omitempty"`  // per-frame [dx,dy]
	Anchor    []int       `json:"anchor,omitempty"`
	Mirror    string      `json:"mirror,omitempty"`
	Events    []AnimEvent `json:"events,omitempty"`

	// model3d
	Clip string  `json:"clip,omitempty"`
	FPS  float64 `json:"fps,omitempty"`

	// billboard3d
	Col           int `json:"col,omitempty"`
	FramesPerView int `json:"framesPerView,omitempty"`
}

type AnimEvent struct {
	Frame int    `json:"frame"`
	SFX   string `json:"sfx"`
}

type UVAnim struct {
	Material string   `json:"material"`
	Frames   int      `json:"frames"`
	ScaleS   *Channel `json:"scaleS,omitempty"`
	ScaleT   *Channel `json:"scaleT,omitempty"`
	Rot      *Channel `json:"rot,omitempty"`
	TransS   *Channel `json:"transS,omitempty"`
	TransT   *Channel `json:"transT,omitempty"`
}

// Channel is either a constant or sampled every Step frames.
type Channel struct {
	Const   *float64  `json:"const,omitempty"`
	Samples []float64 `json:"samples,omitempty"`
	Step    int       `json:"step,omitempty"`
}

type Flipbook struct {
	Material string   `json:"material"`
	Textures []string `json:"textures"`
	Sequence []int    `json:"sequence"`
	Step     int      `json:"step"`
}

type Wireframe struct {
	Positions   []float64   `json:"positions"`
	Edges       [][]int     `json:"edges"` // [v0, v1, faceA, faceB]
	Faces       [][]float64 `json:"faces"` // outward normals
	FaceCenters [][]float64 `json:"faceCenters"`
}

// ---------------------------------------------------------------------------
// Cutscene scripts and camera tracks

type Script struct {
	Header
	Name  string  `json:"name,omitempty"`
	FPS   float64 `json:"fps"`
	Shots []Shot  `json:"shots"`
}

type Shot struct {
	ID     string     `json:"id"`
	Name   string     `json:"name,omitempty"`
	Frames int        `json:"frames"`
	Layers []string   `json:"layers"`
	Actors []Actor    `json:"actors,omitempty"`
	Camera ShotCamera `json:"camera"`
	Lights []Light    `json:"lights,omitempty"`
	Sounds []SoundCue `json:"sounds,omitempty"`
}

type Actor struct {
	Placement int       `json:"placement"`
	Clip      string    `json:"clip,omitempty"`
	Start     int       `json:"start,omitempty"`
	Matrix    []float64 `json:"matrix,omitempty"`
	Mirror    bool      `json:"mirror,omitempty"`
}

type ShotCamera struct {
	Near      float64     `json:"near,omitempty"`
	Far       float64     `json:"far,omitempty"`
	TrackFile string      `json:"trackFile,omitempty"`
	Track     []CamSample `json:"track,omitempty"`
}

type Light struct {
	ID   string     `json:"id"`
	Type string     `json:"type"` // "ambient" | "directional" | "point"
	Keys []LightKey `json:"keys"`
}

type LightKey struct {
	Frame int       `json:"frame"`
	Color string    `json:"color,omitempty"`
	Pos   []float64 `json:"pos,omitempty"`
	Dir   []float64 `json:"dir,omitempty"`
}

type SoundCue struct {
	SFX     string  `json:"sfx"`
	Start   int     `json:"start"`
	End     int     `json:"end,omitempty"`
	Volume  float64 `json:"volume,omitempty"`
	Pan     float64 `json:"pan,omitempty"`
	Reverse bool    `json:"reverse,omitempty"`
}

// CameraTrack is a baked per-frame camera path document.
type CameraTrack struct {
	Header
	Frames int         `json:"frames"`
	FPS    float64     `json:"fps"`
	Near   float64     `json:"near,omitempty"`
	Far    float64     `json:"far,omitempty"`
	Track  []CamSample `json:"track"`
}

type CamSample struct {
	Pos    []float64 `json:"pos"`
	Target []float64 `json:"target"`
	Roll   float64   `json:"roll,omitempty"`
	FOV    float64   `json:"fov,omitempty"`
}

// ---------------------------------------------------------------------------
// The camera block

type Camera struct {
	Mode   string    `json:"mode"` // "map2d" | "orbit" | "fly" | "ortho" | "pan2d"
	Pos    []float64 `json:"pos,omitempty"`
	Target []float64 `json:"target,omitempty"`
	FOV    float64   `json:"fov,omitempty"`
	Near   float64   `json:"near,omitempty"`
	Far    float64   `json:"far,omitempty"`
	Map2D  *Map2D    `json:"map2d,omitempty"`
	Orbit  *Orbit    `json:"orbit,omitempty"`
	Fly    *Fly      `json:"fly,omitempty"`
	Ortho  *Ortho    `json:"ortho,omitempty"`
	Drive  *Drive    `json:"drive,omitempty"`
}

type Map2D struct {
	MinFitFactor    float64 `json:"minFitFactor,omitempty"`
	MaxNativeFactor float64 `json:"maxNativeFactor,omitempty"`
}

type Orbit struct {
	MinDist         float64 `json:"minDist,omitempty"`
	MaxDist         float64 `json:"maxDist,omitempty"`
	AutoRotate      bool    `json:"autoRotate,omitempty"`
	AutoRotateSpeed float64 `json:"autoRotateSpeed,omitempty"`
}

type Fly struct {
	Speed float64 `json:"speed"`
}

type Ortho struct {
	Dir     []float64 `json:"dir"`
	ZoomMin float64   `json:"zoomMin,omitempty"`
	ZoomMax float64   `json:"zoomMax,omitempty"`
}

type Drive struct {
	Route     string  `json:"route"`
	EyeHeight float64 `json:"eyeHeight,omitempty"`
	Speed     float64 `json:"speed"`
	Mode      string  `json:"mode,omitempty"` // "loop" (default) | "pingpong"
}
