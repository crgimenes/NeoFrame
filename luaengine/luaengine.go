package luaengine

import (
	"bufio"
	"log"
	"os"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

// LuaExtender holds an instance of the moon interpreter and the state variables of the extensions we made.
type LuaExtender struct {
	mutex       sync.RWMutex
	luaState    *lua.LState
	Proto       *lua.FunctionProto
	triggerList map[string]*lua.LFunction
	Frame       Frame
	AppCtrl     AppCtrl
}

type KeyValue struct {
	Key   string
	Value string
}

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

type Frame interface {
	DebugPrint(str string)
	DrawLine(x1, y1, x2, y2 int, colorstr string) error
	DrawText(x, y int, size float64, textstr string, fgColor string) error
	GetScreenSize() (width, height int)
	SetWindowTitle(title string)
	SetWindowPosition(x, y int)
}

type AppCtrl interface {
	Shutdown(ret int)
}

// New creates a new instance of LuaExtender.
func New(f Frame, ac AppCtrl) *LuaExtender {

	le := &LuaExtender{
		Frame:   f,
		AppCtrl: ac,
	}
	le.triggerList = make(map[string]*lua.LFunction)
	le.luaState = lua.NewState()
	le.luaState.SetGlobal("clearTriggers", le.luaState.NewFunction(le.ClearTriggers))
	le.luaState.SetGlobal("debugPrint", le.luaState.NewFunction(le.DebugPrint))
	le.luaState.SetGlobal("drawLine", le.luaState.NewFunction(le.DrawLine))
	le.luaState.SetGlobal("drawText", le.luaState.NewFunction(le.DrawText))
	le.luaState.SetGlobal("fileExists", le.luaState.NewFunction(le.fileExists))
	le.luaState.SetGlobal("getScreenSize", le.luaState.NewFunction(le.getScreenSize))
	le.luaState.SetGlobal("logf", le.luaState.NewFunction(le.logf))
	le.luaState.SetGlobal("pwd", le.luaState.NewFunction(le.pwd))
	le.luaState.SetGlobal("readFile", le.luaState.NewFunction(le.readFile))
	le.luaState.SetGlobal("rmTrigger", le.luaState.NewFunction(le.removeTrigger))
	le.luaState.SetGlobal("setWindowTitle", le.luaState.NewFunction(le.SetWindowTitle))
	le.luaState.SetGlobal("shutdown", le.luaState.NewFunction(le.Shutdown))
	le.luaState.SetGlobal("timer", le.luaState.NewFunction(le.timer))
	le.luaState.SetGlobal("trigger", le.luaState.NewFunction(le.trigger))
	le.luaState.SetGlobal("setWindowPosition", le.luaState.NewFunction(le.SetWindowPosition))

	return le
}

func (le *LuaExtender) SetWindowPosition(l *lua.LState) int {
	x := l.ToInt(1)
	y := l.ToInt(2)
	le.Frame.SetWindowPosition(x, y)
	return 0
}

func (le *LuaExtender) DebugPrint(l *lua.LState) int {
	str := l.ToString(1)
	le.Frame.DebugPrint(str)
	return 0
}

func (le *LuaExtender) DrawText(l *lua.LState) int {
	// drawText(x, y, size, text, fgColor)
	// drawText(10, 10, 12, "Hello World", "FFFFFFFF")
	x := l.ToInt(1)
	y := l.ToInt(2)
	size := float64(l.ToNumber(3))
	textstr := l.ToString(4)
	fgColor := l.ToString(5)

	err := le.Frame.DrawText(x, y, size, textstr, fgColor)
	if err != nil {
		log.Println(err)
	}

	return 0
}

func (le *LuaExtender) DrawLine(l *lua.LState) int {
	x1 := l.ToInt(1)
	y1 := l.ToInt(2)
	x2 := l.ToInt(3)
	y2 := l.ToInt(4)
	colorstr := l.ToString(5)
	err := le.Frame.DrawLine(x1, y1, x2, y2, colorstr)
	if err != nil {
		l.Push(lua.LString(err.Error()))
		return 1
	}
	return 0
}

func (le *LuaExtender) Shutdown(l *lua.LState) int {
	ret := l.ToInt(1)
	le.AppCtrl.Shutdown(ret)
	return 0
}

func (le *LuaExtender) SetWindowTitle(l *lua.LState) int {
	title := l.ToString(1)
	le.Frame.SetWindowTitle(title)
	return 0
}

func (le *LuaExtender) getScreenSize(l *lua.LState) int {
	width, height := le.Frame.GetScreenSize()
	l.Push(lua.LNumber(width))
	l.Push(lua.LNumber(height))
	return 2
}

// Run executes the passed lua code.
func (le *LuaExtender) Run(code string) error {
	return le.luaState.DoString(code)
}

func (le *LuaExtender) logf(l *lua.LState) int {
	format := l.ToString(1)
	args := make([]interface{}, l.GetTop()-1)
	for i := 2; i <= l.GetTop(); i++ {
		args[i-2] = l.ToString(i)
	}
	log.Printf(format, args...)
	return 0
}

func (le *LuaExtender) Close() error {
	le.luaState.Close()
	return nil
}

// GetState returns the state of the moon interpreter.
func (le *LuaExtender) GetState() *lua.LState {
	return le.luaState
}

// CompileLua reads the passed lua file from disk and compiles it.
func (le *LuaExtender) Compile(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := parse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}

// doCompiledFile takes a FunctionProto, as returned by CompileLua, and runs it in the LState. It is equivalent
// to calling DoFile on the LState with the original source file.
func (le *LuaExtender) doCompiledFile(L *lua.LState, proto *lua.FunctionProto) error {
	lfunc := L.NewFunctionFromProto(proto)
	L.Push(lfunc)
	return L.PCall(0, lua.MultRet, nil)
}

// InitState starts the lua interpreter with a script.
func (le *LuaExtender) InitStateWithProto() error {
	return le.doCompiledFile(le.luaState, le.Proto)
}

// RunTrigger executes a pre-configured trigger.
func (le *LuaExtender) RunTrigger(name string) (bool, error) {
	f, ok := le.triggerList[name]
	if !ok {
		return false, nil
	}

	err := le.luaState.CallByParam(lua.P{
		Fn:      f,    // Lua function
		NRet:    0,    // number of returned values
		Protect: true, // return err or panic
	})
	return true, err
}

func (le *LuaExtender) removeTrigger(l *lua.LState) int {
	n := l.ToString(1) // name
	le.mutex.Lock()
	delete(le.triggerList, n)
	le.mutex.Unlock()
	return 0
}

func (le *LuaExtender) ClearTriggers(l *lua.LState) int {
	le.mutex.Lock()
	le.triggerList = make(map[string]*lua.LFunction)
	le.mutex.Unlock()
	return 0
}

func (le *LuaExtender) timer(l *lua.LState) int {
	n := l.ToString(1)   // name
	t := l.ToInt(2)      // timer
	f := l.ToFunction(3) // function

	if n == "" {
		n = "timer"
	}

	le.mutex.Lock()
	le.triggerList[n] = f
	le.mutex.Unlock()

	go func() {
		for {
			<-time.After(time.Duration(t) * time.Millisecond)
			le.mutex.Lock()
			_, ok := le.triggerList[n]
			le.mutex.Unlock()
			le.mutex.Lock()
			ok, err := le.RunTrigger(n)
			le.mutex.Unlock()
			if err != nil {
				log.Println(n, "timer trigger error", err)
				return
			}
			if !ok {
				return
			}
		}
	}()

	return 0
}

func (le *LuaExtender) trigger(l *lua.LState) int {
	a := l.ToString(1)
	f := l.ToFunction(2)

	le.mutex.Lock()
	le.triggerList[a] = f
	le.mutex.Unlock()

	res := lua.LString(a)
	l.Push(res)
	return 1
}

func pwd() string {
	d, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	return d
}

func (le *LuaExtender) pwd(l *lua.LState) int {
	res := lua.LString(pwd())
	l.Push(res)
	return 1
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (le *LuaExtender) fileExists(l *lua.LState) int {
	filename := l.ToString(1)
	res := lua.LBool(fileExists(filename))
	l.Push(res)
	return 1
}

func (le *LuaExtender) readFile(l *lua.LState) int {
	file := l.ToString(1)
	content, err := os.ReadFile(file)
	if err != nil {
		log.Printf("error reading file %v, %v", file, err)
		return 0
	}
	l.Push(lua.LString(string(content)))
	return 1
}
