package main

import (
	"code.google.com/p/ebml-go/common"
	"code.google.com/p/ffvorbis-go/ffvorbis"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	"math"
	"runtime"
	"time"
//	"fmt"
//	"unsafe"
)

/*
 #cgo LDFLAGS: -framework OpenAL
#include <OpenAL/al.h>
#include <OpenAL/alc.h>
*/
import "C"

var (
	unsync     = flag.Bool("u", false, "Unsynchronized display")
	notc       = flag.Bool("t", false, "Ignore timecodes")
	blend      = flag.Bool("b", false, "Blend between images")
	fullscreen = flag.Bool("f", false, "Fullscreen mode")
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
uniform sampler2D yt0;
uniform sampler2D cbt0;
uniform sampler2D crt0;

void main() {
   vec3 c = vec3(texture2D(yt0, gl_TexCoord[0].st).r,
                 texture2D(cbt0, gl_TexCoord[0].st).r,
                 texture2D(crt0, gl_TexCoord[0].st).r);
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
	names := []string{"yt0", "cbt0", "crt0", "yt1", "cbt1", "crt1"}
	for i := 0; i < l; i++ {
		loc := gl.GetUniformLocation(prg, gl.GLString(names[i]))
		gl.Uniform1i(loc, gl.Int(i))
	}
	return gl.GetUniformLocation(prg, gl.GLString("factor"))
}

func upload(id gl.Uint, data []byte, stride int, w int, h int) {
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, gl.Int(stride))
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

func vpresent(wchan <-chan *ffvp8.Frame) {
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
	glfw.SetWindowTitle(*common.In)
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
		if *notc || t.After(tbase.Add(img.Timecode)) {
			var ok bool
			pimg = img
			img, ok = <-wchan
			if !ok {
				return
			}
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
			upload(1, pimg.Y, pimg.YStride, w, h)
			gl.ActiveTexture(gl.TEXTURE4)
			upload(2, pimg.Cb, pimg.CStride, w/2, h/2)
			gl.ActiveTexture(gl.TEXTURE5)
			upload(3, pimg.Cr, pimg.CStride, w/2, h/2)
		}
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		runtime.GC()
		glfw.SwapBuffers()
	}
}

func apresent(wchan <-chan *ffvorbis.Samples) {
	dev := C.alcOpenDevice(nil)
	defer C.alcCloseDevice(dev)
	ctx := C.alcCreateContext(dev, nil)
	defer C.alcDestroyContext(ctx)
	C.alcMakeContextCurrent(ctx)
	defer C.alcMakeContextCurrent(nil)
	C.alListener3f(C.AL_POSITION, 0, 0, 0)
	C.alListener3f(C.AL_VELOCITY, 0, 0, 0)
	C.alListener3f(C.AL_ORIENTATION, 0, 0, -1)
	var src C.ALuint
	C.alGenSources(1, &src)
	C.alSourcef(src, C.AL_PITCH, 1)
	C.alSourcef(src, C.AL_GAIN, 1)
	C.alSource3f(src, C.AL_POSITION, 0, 0, 0)
	C.alSource3f(src, C.AL_VELOCITY, 0, 0, 0)
	C.alSourcei(src, C.AL_LOOPING, C.AL_FALSE)
	const nbuf = 16
	var buf [nbuf]C.ALuint
	C.alGenBuffers(nbuf, &buf[0])
//	curr := 0
	for p := range wchan {
/*
		var proc C.ALint
		proc = 0
		if curr >= nbuf {
			for proc == 0 {
				C.alGetSourcei(src, C.AL_BUFFERS_PROCESSED, &proc)
				time.Sleep(time.Millisecond)
			}
			var ub C.ALuint
			C.alSourceUnqueueBuffers(src, 1, &ub)
		}
		b := buf[curr%nbuf]
		curr++
		C.alBufferData(b, C.AL_FORMAT_MONO16,
			unsafe.Pointer(&p.Data[0]), C.ALsizei(2*len(p.Data)), 44100)
		C.alSourceQueueBuffers(src, 1, &b)
		switch {
		case curr == nbuf:
			C.alSourcePlay(src)
		case curr > nbuf:
			var state C.ALint
			C.alGetSourcei(src, C.AL_SOURCE_STATE, &state)
			if state != C.AL_PLAYING {
				C.alSourcePlay(src)
			}
		}
		C.alcProcessContext(ctx)
*/
	}
}

func main() {
	common.Main(vpresent, apresent)
}
