package main

import (
	"code.google.com/p/ebml-go/common"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	"math"
	"runtime"
	"time"
)

var (
	unsync = flag.Bool("u", false, "Unsynchronized display")
	notc   = flag.Bool("t", false, "Ignore timecodes")
	blend  = flag.Bool("b", false, "Blend between images")
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
	return
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

func setupvp(w, h int) {
	gl.Viewport(0, 0, gl.Sizei(w), gl.Sizei(h))
}

func factor(t time.Time, tc0 time.Time, tc1 time.Time) gl.Float {
	num := t.Sub(tc0)
	den := tc1.Sub(tc0)
	res := num.Seconds() / den.Seconds()
	res = math.Max(res, 0)
	res = math.Min(res, 1)
	return gl.Float(res)
}

func write(wchan <-chan *ffvp8.Frame) {
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
	glfw.OpenWindow(w, h, 0, 0, 0, 0, 0, 0, glfw.Windowed)
	defer glfw.CloseWindow()
	glfw.SetWindowSizeCallback(setupvp)
	if !*unsync {
		glfw.SetSwapInterval(1)
	}
	glfw.SetWindowTitle("webmplay")
	for i := 0; i < ntex; i++ {
		texinit(i + 1)
	}
	factorloc := shinit()
	initquad()
	gl.Enable(gl.TEXTURE_2D)
	tbase := time.Now()
	pimg := img
	for glfw.WindowParam(glfw.Opened) == 1 {
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

func main() {
	common.Main(write)
}
