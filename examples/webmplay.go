package main

import (
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/portaudio-go/portaudio"
	"flag"
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

var (
	in         = flag.String("i", "", "Input file")
	unsync     = flag.Bool("u", false, "Unsynchronized display")
	notc       = flag.Bool("t", false, "Ignore timecodes")
	blend      = flag.Bool("b", false, "Blend between images")
	fullscreen = flag.Bool("f", false, "Fullscreen mode")
	justaudio  = flag.Bool("a", false, "Just audio")
	justvideo  = flag.Bool("v", false, "Just video")
)

var ntex int

const vss = `
void main() {
  gl_TexCoord[0] = gl_MultiTexCoord0;
  gl_Position = ftransform();
}
`

const ycbcr2rgb = `
const mat3 ycbcr2rgb = mat3(
                          1.164, 0, 1.596,
                          1.164, -0.392, -0.813,
                          1.164, 2.017, 0.0
                          );
const float ysub = 0.0625;
vec3 ycbcr2rgb(vec3 c) {
   vec3 ycbcr = vec3(c.x - ysub, c.y - 0.5, c.z - 0.5);
   return ycbcr * ycbcr2rgb;
}
`

const fss = ycbcr2rgb + `
uniform sampler2D yt1;
uniform sampler2D cbt1;
uniform sampler2D crt1;

void main() {
   vec3 c = vec3(texture2D(yt1, gl_TexCoord[0].st).r,
                 texture2D(cbt1, gl_TexCoord[0].st).r,
                 texture2D(crt1, gl_TexCoord[0].st).r);
   gl_FragColor = vec4(ycbcr2rgb(c), 1.0);
}
`
const bfss = ycbcr2rgb + `
uniform sampler2D yt1;
uniform sampler2D cbt1;
uniform sampler2D crt1;
uniform sampler2D yt0;
uniform sampler2D cbt0;
uniform sampler2D crt0;
uniform float factor;

void main() {
   vec3 c0 = vec3(texture2D(yt0, gl_TexCoord[0].st).r,
                  texture2D(cbt0, gl_TexCoord[0].st).r,
                  texture2D(crt0, gl_TexCoord[0].st).r);
   vec3 c1 = vec3(texture2D(yt1, gl_TexCoord[0].st).r,
                  texture2D(cbt1, gl_TexCoord[0].st).r,
                  texture2D(crt1, gl_TexCoord[0].st).r);
   gl_FragColor = vec4(ycbcr2rgb(mix(c0, c1, factor)), 1);
}
`

func texinit(id int) {
	gl.BindTexture(gl.TEXTURE_2D, gl.Uint(id))
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}

func shinit() gl.Int {
	vs := loadShader(gl.VERTEX_SHADER, vss)
	var sfs string
	if *blend {
		sfs = bfss
	} else {
		sfs = fss
	}
	fs := loadShader(gl.FRAGMENT_SHADER, sfs)
	prg := gl.CreateProgram()
	gl.AttachShader(prg, vs)
	gl.AttachShader(prg, fs)
	gl.LinkProgram(prg)
	var l int
	if *blend {
		l = 6
	} else {
		l = 3
	}
	gl.UseProgram(prg)
	names := []string{"yt1", "cbt1", "crt1", "yt0", "cbt0", "crt0"}
	for i := 0; i < l; i++ {
		loc := gl.GetUniformLocation(prg, gl.GLString(names[i]))
		gl.Uniform1i(loc, gl.Int(i))
	}
	return gl.GetUniformLocation(prg, gl.GLString("factor"))
}

func upload(id gl.Uint, data []byte, stride int, w int, h int) {
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, gl.Int(stride))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE, gl.Sizei(w), gl.Sizei(h), 0,
		gl.LUMINANCE, gl.UNSIGNED_BYTE, gl.Pointer(&data[0]))
}

func initquad() {
	ver := []gl.Float{-1, 1, 1, 1, -1, -1, 1, -1}
	gl.BindBuffer(gl.ARRAY_BUFFER, 1)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(ver)),
		gl.Pointer(&ver[0]), gl.STATIC_DRAW)
	gl.VertexPointer(2, gl.FLOAT, 0, nil)
	tex := []gl.Float{0, 0, 1, 0, 0, 1, 1, 1}
	gl.BindBuffer(gl.ARRAY_BUFFER, 2)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(tex)),
		gl.Pointer(&tex[0]), gl.STATIC_DRAW)
	gl.TexCoordPointer(2, gl.FLOAT, 0, nil)
	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
}

func loadShader(shtype gl.Enum, src string) gl.Uint {
	sh := gl.CreateShader(shtype)
	gsrc := gl.GLString(src)
	gl.ShaderSource(sh, 1, &gsrc, nil)
	gl.CompileShader(sh)
	return sh
}

func factor(t time.Time, tc0 time.Time, tc1 time.Time) gl.Float {
	num := t.Sub(tc0)
	den := tc1.Sub(tc0)
	res := num.Seconds() / den.Seconds()
	res = math.Max(res, 0)
	res = math.Min(res, 1)
	return gl.Float(res)
}

var steps = uint(0xffffffff)
var seek = time.Duration(-1)
var duration = time.Duration(-1)

func mbhandler(button, state int) {
	if state == glfw.KeyRelease {
		x, _ := glfw.MousePos()
		w, _ := glfw.WindowSize()
		factor := float64(x) / float64(w)
		seek = time.Duration(float64(duration) * factor)
	}
}

func khandler(key, state int) {
	if state == glfw.KeyRelease {
		switch key {
		case 'P':
			steps = 0
		case 'R':
			steps = 0xffffffff
		case 'S':
			steps = 1
		}
	}
}

func vpresent(wchan <-chan webm.Frame, reader *webm.Reader) {
	if *blend {
		ntex = 6
	} else {
		ntex = 3
	}
	img := <-wchan
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	gl.Init()
	glfw.Init()
	defer glfw.Terminate()
	mode := glfw.Windowed
	ww := w
	wh := h
	if *fullscreen {
		mode = glfw.Fullscreen
		ww = 1440
		wh = 900
	}
	glfw.OpenWindow(ww, wh, 0, 0, 0, 0, 0, 0, mode)
	glfw.SetKeyCallback(khandler)
	glfw.SetMouseButtonCallback(mbhandler)
	defer glfw.CloseWindow()
	glfw.SetWindowSizeCallback(func(ww, wh int) {
		oaspect := float64(w) / float64(h)
		haspect := float64(ww) / float64(wh)
		vaspect := float64(wh) / float64(ww)
		var scx, scy float64
		if oaspect > haspect {
			scx = 1
			scy = haspect / oaspect
		} else {
			scx = vaspect * oaspect
			scy = 1
		}
		gl.Viewport(0, 0, gl.Sizei(ww), gl.Sizei(wh))
		gl.LoadIdentity()
		gl.Scaled(gl.Double(scx), gl.Double(scy), 1)
	})
	if !*unsync {
		glfw.SetSwapInterval(1)
	}
	glfw.SetWindowTitle(*in)
	for i := 0; i < ntex; i++ {
		texinit(i + 1)
	}
	factorloc := shinit()
	initquad()
	gl.Enable(gl.TEXTURE_2D)
	tbase := time.Now()
	pimg := img
	for glfw.WindowParam(glfw.Opened) == 1 {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		t := time.Now()
		if steps > 0 && (*notc || t.After(tbase.Add(img.Timecode))) {
			pimg = img
			var ok bool
			img, ok = <-wchan
			if !ok {
				return
			}
			if img.Timecode == pimg.Timecode {
				log.Println("same timecode", img.Timecode)
			}
			steps--
		}
		gl.ActiveTexture(gl.TEXTURE0)
		upload(1, img.Y, img.YStride, w, h)
		gl.ActiveTexture(gl.TEXTURE1)
		upload(2, img.Cb, img.CStride, w/2, h/2)
		gl.ActiveTexture(gl.TEXTURE2)
		upload(3, img.Cr, img.CStride, w/2, h/2)
		if *blend {
			gl.Uniform1f(factorloc, factor(t,
				tbase.Add(pimg.Timecode),
				tbase.Add(img.Timecode)))
			gl.ActiveTexture(gl.TEXTURE3)
			upload(4, pimg.Y, pimg.YStride, w, h)
			gl.ActiveTexture(gl.TEXTURE4)
			upload(5, pimg.Cb, pimg.CStride, w/2, h/2)
			gl.ActiveTexture(gl.TEXTURE5)
			upload(6, pimg.Cr, pimg.CStride, w/2, h/2)
		}
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		runtime.GC()
		if seek >= 0 {
			reader.Seek(seek)
			seek = time.Duration(-1)
		}
		glfw.SwapBuffers()
	}
}

type AudioWriter struct {
	ch       <-chan webm.Samples
	channels int
	active   bool
	sofar    int
	curr     webm.Samples
}

func (aw *AudioWriter) ProcessAudio(in, out []float32) {
	for sent, lo := 0, len(out); sent < lo; {
		if aw.sofar == len(aw.curr.Data) {
			aw.curr, aw.active = <-aw.ch
			aw.sofar = 0
			//log.Println("timecode", aw.curr.Timecode)
		}
		s := copy(out[sent:], aw.curr.Data[aw.sofar:])
		sent += s
		aw.sofar += s
	}
}

func apresent(wchan <-chan webm.Samples, audio *webm.Audio) {
	chk := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	channels := int(audio.Channels)
	aw := AudioWriter{ch: wchan, channels: channels, active: true}
	stream, err := portaudio.OpenDefaultStream(0, channels,
		audio.SamplingFrequency, 0, &aw)
	defer stream.Close()
	chk(err)
	chk(stream.Start())
	defer stream.Stop()
	for aw.active {
		time.Sleep(time.Second)
	}
}

func main() {
	flag.Parse()

	var err error
	var wm webm.WebM
	r, err := os.Open(*in)
	defer r.Close()
	if err != nil {
		log.Panic("Unable to open file " + *in)
	}
	reader, err := webm.Parse(r, &wm)
	if err != nil {
		log.Panic("Unable to parse file:", err)
	}
	duration = time.Duration(wm.GetDuration()) * time.Second

	splitter := webm.NewSplitter(reader.Chan)

	var vtrack *webm.TrackEntry
	var vstream *webm.Stream
	if !*justaudio {
		vtrack = wm.FindFirstVideoTrack()
	}
	if vtrack != nil {
		vstream = webm.NewStream(vtrack)
	}
	var astream *webm.Stream
	var atrack *webm.TrackEntry
	if !*justvideo {
		atrack = wm.FindFirstAudioTrack()
	}
	if atrack != nil {
		astream = webm.NewStream(atrack)
	}
	splitter.Split(astream, vstream)
	switch {
	case astream != nil && vstream != nil:
		go apresent(astream.AudioChannel(), &atrack.Audio)
		fallthrough
	case vstream != nil:
		vpresent(vstream.VideoChannel(), reader)
	case astream != nil:
		apresent(astream.AudioChannel(), &atrack.Audio)
	}
}
