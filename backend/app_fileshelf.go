// Package backend provides file shelf staging, native OLE drag-and-drop, and media utilities.
package backend

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

var (
	procOleInitialize      = ole32.NewProc("OleInitialize")
	procOleUninitialize    = ole32.NewProc("OleUninitialize")
	procCoCreateInstance   = ole32.NewProc("CoCreateInstance")
	procCoTaskMemFree      = ole32.NewProc("CoTaskMemFree")
	procSHParseDisplayName = shell32.NewProc("SHParseDisplayName")
	procSHCreateDataObject = shell32.NewProc("SHCreateDataObject")
	procSHGetFileInfoW     = shell32.NewProc("SHGetFileInfoW")
	procSHDoDragDrop       = shell32.NewProc("SHDoDragDrop")
	procCreateDIBSection   = gdi32.NewProc("CreateDIBSection")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC           = gdi32.NewProc("DeleteDC")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procDrawIconEx         = user32.NewProc("DrawIconEx")
	procDestroyIcon        = user32.NewProc("DestroyIcon")
	procDrawTextW          = user32.NewProc("DrawTextW")
	procSetBkMode          = gdi32.NewProc("SetBkMode")
	procSetTextColor       = gdi32.NewProc("SetTextColor")
	procGlobalAlloc        = kernel32.NewProc("GlobalAlloc")
	procGlobalLock         = kernel32.NewProc("GlobalLock")
	procGlobalUnlock       = kernel32.NewProc("GlobalUnlock")
	procGlobalFree         = kernel32.NewProc("GlobalFree")
	procOpenClipboard      = user32.NewProc("OpenClipboard")
	procEmptyClipboard     = user32.NewProc("EmptyClipboard")
	procSetClipboardData   = user32.NewProc("SetClipboardData")
	procCloseClipboard     = user32.NewProc("CloseClipboard")
	procLoadCursorW        = user32.NewProc("LoadCursorW")
	procSetCursor          = user32.NewProc("SetCursor")
)

const (
	dropEffectCopy = 1
	dropEffectMove = 2
	dropEffectLink = 4

	cfHDrop           = 15
	cfUnicodeText     = 13
	dvaspectContent   = 1
	tymedHGlobal      = 1
	datadirGet        = 1
	gmemMoveable      = 0x0002
	gmemZeroInit      = 0x0040
	wmCancelMode      = 0x001F
	wmMouseMove       = 0x0200
	sOK               = 0x00000000
	sFalse            = 0x00000001
	eNotImpl          = 0x80004001
	eNoInterface      = 0x80004002
	ePointer          = 0x80004003
	eInvalidArg       = 0x80070057
	dvEFormatEtc      = 0x80040064
	dvETymed          = 0x80040069
	dvEDvaspect       = 0x8004006B
	oleEAdviseNotSupp = 0x80040003
	dragdropSDrop     = 0x00040100
	dragdropSCancel   = 0x00040101
	dragdropSDefault  = 0x00040102
)

var (
	iidIUnknown = windows.GUID{
		Data1: 0x00000000,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	iidIDataObject = windows.GUID{
		Data1: 0x0000010e,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	iidIDropSource = windows.GUID{
		Data1: 0x00000121,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	iidIEnumFORMATETC = windows.GUID{
		Data1: 0x00000103,
		Data2: 0x0000,
		Data3: 0x0000,
		Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46},
	}
	clsidDragDropHelper = windows.GUID{
		Data1: 0x4657278a,
		Data2: 0x411b,
		Data3: 0x11d2,
		Data4: [8]byte{0x83, 0x9a, 0x00, 0xc0, 0x4f, 0xd9, 0x18, 0xd0},
	}
	iidIDragSourceHelper = windows.GUID{
		Data1: 0xde5bf786,
		Data2: 0x477a,
		Data3: 0x11d2,
		Data4: [8]byte{0x83, 0x9d, 0x00, 0xc0, 0x4f, 0xd9, 0x18, 0xd0},
	}
)

type shFileInfoW struct {
	hIcon         uintptr
	iIcon         int32
	dwAttributes  uint32
	szDisplayName [260]uint16
	szTypeName    [80]uint16
}

type shDragImage struct {
	cx            int32
	cy            int32
	ptOffsetX     int32
	ptOffsetY     int32
	hbmpDragImage uintptr
	crColorKey    uint32
	_             [4]byte
}

type bitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type formatEtc struct {
	cfFormat uint16
	_        [6]byte
	ptd      uintptr
	dwAspect uint32
	lindex   int32
	tymed    uint32
	_        [4]byte
}

type stgMedium struct {
	tymed          uint32
	_              [4]byte
	hGlobal        uintptr
	pUnkForRelease uintptr
}

type dropFilesHeader struct {
	pFiles uint32
	ptX    int32
	ptY    int32
	fNC    int32
	fWide  int32
}

type comDropSourceVtbl struct {
	QueryInterface    uintptr
	AddRef            uintptr
	Release           uintptr
	QueryContinueDrag uintptr
	GiveFeedback      uintptr
}

type comDropSource struct {
	lpVtbl *comDropSourceVtbl
}

type comEnumFormatEtcVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Next           uintptr
	Skip           uintptr
	Reset          uintptr
	Clone          uintptr
}

type comEnumFormatEtc struct {
	lpVtbl *comEnumFormatEtcVtbl
	index  uint32
}

type comDataObjectVtbl struct {
	QueryInterface        uintptr
	AddRef                uintptr
	Release               uintptr
	GetData               uintptr
	GetDataHere           uintptr
	QueryGetData          uintptr
	GetCanonicalFormatEtc uintptr
	SetData               uintptr
	EnumFormatEtc         uintptr
	DAdvise               uintptr
	DUnadvise             uintptr
	EnumDAdvise           uintptr
}

type comDataObject struct {
	lpVtbl *comDataObjectVtbl
}

var (
	comDragOnce         sync.Once
	dropSourceVtbl      comDropSourceVtbl
	dropSourceInst      comDropSource
	enumFormatEtcVtbl   comEnumFormatEtcVtbl
	dataObjectVtbl      comDataObjectVtbl
	activeDragMu        sync.RWMutex
	activeDragPaths     []string
	enumPoolMu          sync.Mutex
	activeEnums         []*comEnumFormatEtc
	dragSessionMu       sync.Mutex
	dragSessionActive   bool
	dragCancelRequested atomic.Bool
	activeIslandMu      sync.RWMutex
	activeIslandHwnd    uintptr
	activeIslandRects   []pillHitBounds
	isDragOverIsland    atomic.Bool
	hasLeftIsland       atomic.Bool
)

// isCursorOverIsland checks whether the cursor is currently over Vinta's island bounds.
func isCursorOverIsland() bool {
	var curPt point32
	if res, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&curPt))); res == 0 {
		return false
	}

	activeIslandMu.RLock()
	hwnd := activeIslandHwnd
	rects := activeIslandRects
	activeIslandMu.RUnlock()

	if hwnd == 0 || len(rects) == 0 {
		return false
	}

	clientPt := curPt
	if res, _, _ := procScreenToClient.Call(hwnd, uintptr(unsafe.Pointer(&clientPt))); res == 0 {
		return false
	}

	for _, r := range rects {
		if clientPt.X >= r.left && clientPt.X <= r.right &&
			clientPt.Y >= r.top && clientPt.Y <= r.bottom {
			return true
		}
	}
	return false
}

func initCOMDragObjects() {
	comDragOnce.Do(func() {
		dropSourceVtbl = comDropSourceVtbl{
			QueryInterface: syscall.NewCallback(func(this uintptr, riid *windows.GUID, ppv *uintptr) uintptr {
				if ppv == nil {
					return ePointer
				}
				if riid != nil && (*riid == iidIUnknown || *riid == iidIDropSource) {
					*ppv = this
					return sOK
				}
				*ppv = 0
				return eNoInterface
			}),
			AddRef: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			Release: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			QueryContinueDrag: syscall.NewCallback(func(_ uintptr, fEscapePressed int32, grfKeyState uint32) uintptr {
				if fEscapePressed != 0 || dragCancelRequested.Load() {
					return dragdropSCancel
				}
				if escState, _, _ := procGetAsyncKeyState.Call(0x1B); (escState & 0x8000) != 0 {
					return dragdropSCancel
				}

				// Right-clicking while left-dragging cancels the drag
				if (grfKeyState&0x0001) != 0 && (grfKeyState&0x0002) != 0 {
					return dragdropSCancel
				}

				// When the left mouse button is released, complete the drop (or cancel if over island)
				if (grfKeyState & 0x0001) == 0 {
					if isCursorOverIsland() {
						return dragdropSCancel
					}
					return dragdropSDrop
				}
				return sOK
			}),
			GiveFeedback: syscall.NewCallback(func(_ uintptr, _ uint32) uintptr {
				over := isCursorOverIsland()
				if !over {
					hasLeftIsland.Store(true)
				}

				draggedBack := over && hasLeftIsland.Load()
				prev := isDragOverIsland.Swap(draggedBack)
				if draggedBack != prev {
					if app := application.Get(); app != nil && app.Event != nil {
						app.Event.Emit("fileshelf:drag-over-shelf", draggedBack)
					}
				}

				if draggedBack {
					hCursor, _, _ := procLoadCursorW.Call(0, 32512) // IDC_NO
					if hCursor != 0 {
						_, _, _ = procSetCursor.Call(hCursor)
						return sOK
					}
				}
				return dragdropSDefault
			}),
		}
		dropSourceInst.lpVtbl = &dropSourceVtbl

		enumFormatEtcVtbl = comEnumFormatEtcVtbl{
			QueryInterface: syscall.NewCallback(func(this uintptr, riid *windows.GUID, ppv *uintptr) uintptr {
				if ppv == nil {
					return ePointer
				}
				if riid != nil && (*riid == iidIUnknown || *riid == iidIEnumFORMATETC) {
					*ppv = this
					return sOK
				}
				*ppv = 0
				return eNoInterface
			}),
			AddRef: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			Release: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			Next: syscall.NewCallback(func(this uintptr, celt uint32, rgelt *formatEtc, pceltFetched *uint32) uintptr {
				if rgelt == nil {
					return ePointer
				}
				//nolint:govet // Win32 COM this pointer
				enumObj := (*comEnumFormatEtc)(unsafe.Pointer(this))
				if enumObj.index == 0 && celt >= 1 {
					*rgelt = formatEtc{
						cfFormat: cfHDrop,
						ptd:      0,
						dwAspect: dvaspectContent,
						lindex:   -1,
						tymed:    tymedHGlobal,
					}
					enumObj.index = 1
					if pceltFetched != nil {
						*pceltFetched = 1
					}
					if celt == 1 {
						return sOK
					}
					return sFalse
				}
				if pceltFetched != nil {
					*pceltFetched = 0
				}
				return sFalse
			}),
			Skip: syscall.NewCallback(func(this uintptr, celt uint32) uintptr {
				//nolint:govet // Win32 COM this pointer
				enumObj := (*comEnumFormatEtc)(unsafe.Pointer(this))
				enumObj.index += celt
				if enumObj.index > 1 {
					return sFalse
				}
				return sOK
			}),
			Reset: syscall.NewCallback(func(this uintptr) uintptr {
				//nolint:govet // Win32 COM this pointer
				enumObj := (*comEnumFormatEtc)(unsafe.Pointer(this))
				enumObj.index = 0
				return sOK
			}),
			Clone: syscall.NewCallback(func(this uintptr, ppEnum *uintptr) uintptr {
				if ppEnum == nil {
					return ePointer
				}
				//nolint:govet // Win32 COM this pointer
				orig := (*comEnumFormatEtc)(unsafe.Pointer(this))
				cloned := &comEnumFormatEtc{
					lpVtbl: &enumFormatEtcVtbl,
					index:  orig.index,
				}
				enumPoolMu.Lock()
				activeEnums = append(activeEnums, cloned)
				enumPoolMu.Unlock()
				*ppEnum = uintptr(unsafe.Pointer(cloned))
				return sOK
			}),
		}

		dataObjectVtbl = comDataObjectVtbl{
			QueryInterface: syscall.NewCallback(func(this uintptr, riid *windows.GUID, ppv *uintptr) uintptr {
				if ppv == nil {
					return ePointer
				}
				if riid != nil && (*riid == iidIUnknown || *riid == iidIDataObject) {
					*ppv = this
					return sOK
				}
				*ppv = 0
				return eNoInterface
			}),
			AddRef: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			Release: syscall.NewCallback(func(_ uintptr) uintptr {
				return 1
			}),
			GetData: syscall.NewCallback(func(_ uintptr, pFormatEtc *formatEtc, pMedium *stgMedium) uintptr {
				if pFormatEtc == nil || pMedium == nil {
					return ePointer
				}
				if pFormatEtc.cfFormat != cfHDrop {
					return dvEFormatEtc
				}
				if pFormatEtc.dwAspect != 0 && (pFormatEtc.dwAspect&dvaspectContent) == 0 {
					return dvEDvaspect
				}
				if (pFormatEtc.tymed & tymedHGlobal) == 0 {
					return dvETymed
				}
				activeDragMu.RLock()
				paths := make([]string, len(activeDragPaths))
				copy(paths, activeDragPaths)
				activeDragMu.RUnlock()

				hGlobal := buildCFHDropMedium(paths)
				if hGlobal == 0 {
					return dvEFormatEtc
				}
				pMedium.tymed = tymedHGlobal
				pMedium.hGlobal = hGlobal
				pMedium.pUnkForRelease = 0
				return sOK
			}),
			GetDataHere: syscall.NewCallback(func(_ uintptr, _ *formatEtc, _ *stgMedium) uintptr {
				return eNotImpl
			}),
			QueryGetData: syscall.NewCallback(func(_ uintptr, pFormatEtc *formatEtc) uintptr {
				if pFormatEtc == nil {
					return ePointer
				}
				if pFormatEtc.cfFormat != cfHDrop {
					return dvEFormatEtc
				}
				if pFormatEtc.dwAspect != 0 && (pFormatEtc.dwAspect&dvaspectContent) == 0 {
					return dvEDvaspect
				}
				if (pFormatEtc.tymed & tymedHGlobal) == 0 {
					return dvETymed
				}
				return sOK
			}),
			GetCanonicalFormatEtc: syscall.NewCallback(func(_ uintptr, _ *formatEtc, pFormatEtcOut *formatEtc) uintptr {
				if pFormatEtcOut == nil {
					return ePointer
				}
				pFormatEtcOut.ptd = 0
				return 0x00040130 // DATA_S_SAMEFORMATETC
			}),
			SetData: syscall.NewCallback(func(_ uintptr, _ *formatEtc, _ *stgMedium, _ int32) uintptr {
				return sOK
			}),
			EnumFormatEtc: syscall.NewCallback(func(_ uintptr, dwDirection uint32, ppEnumFormatEtc *uintptr) uintptr {
				if ppEnumFormatEtc == nil {
					return ePointer
				}
				if dwDirection != datadirGet {
					*ppEnumFormatEtc = 0
					return eNotImpl
				}
				enumInst := &comEnumFormatEtc{
					lpVtbl: &enumFormatEtcVtbl,
					index:  0,
				}
				enumPoolMu.Lock()
				activeEnums = append(activeEnums, enumInst)
				enumPoolMu.Unlock()
				*ppEnumFormatEtc = uintptr(unsafe.Pointer(enumInst))
				return sOK
			}),
			DAdvise: syscall.NewCallback(func(_ uintptr, _ *formatEtc, _ uint32, _ uintptr, _ *uint32) uintptr {
				return oleEAdviseNotSupp
			}),
			DUnadvise: syscall.NewCallback(func(_ uintptr, _ uint32) uintptr {
				return oleEAdviseNotSupp
			}),
			EnumDAdvise: syscall.NewCallback(func(_ uintptr, _ *uintptr) uintptr {
				return oleEAdviseNotSupp
			}),
		}
	})
}

// buildCFHDropMedium allocates a global memory block containing a DROPFILES header and double-null-terminated UTF-16 file paths.
func buildCFHDropMedium(paths []string) uintptr {
	var utf16Units []uint16
	for _, p := range paths {
		encoded, err := syscall.UTF16FromString(p)
		if err != nil || len(encoded) == 0 {
			continue
		}
		utf16Units = append(utf16Units, encoded...)
	}
	if len(utf16Units) == 0 {
		return 0
	}
	// Append final terminating null character for the double-null-terminated list
	utf16Units = append(utf16Units, 0)

	headerSize := unsafe.Sizeof(dropFilesHeader{})
	payloadBytes := uintptr(len(utf16Units) * 2)
	totalBytes := headerSize + payloadBytes

	hGlobal, _, _ := procGlobalAlloc.Call(gmemMoveable|gmemZeroInit, totalBytes)
	if hGlobal == 0 {
		return 0
	}

	ptr, _, _ := procGlobalLock.Call(hGlobal)
	if ptr == 0 {
		_, _, _ = procGlobalFree.Call(hGlobal)
		return 0
	}

	//nolint:govet // Win32 HGLOBAL memory layout
	header := (*dropFilesHeader)(unsafe.Pointer(ptr))
	header.pFiles = uint32(headerSize)
	header.fWide = 1

	//nolint:govet // Win32 HGLOBAL memory layout
	dstSlice := unsafe.Slice((*uint16)(unsafe.Pointer(ptr+headerSize)), len(utf16Units))
	copy(dstSlice, utf16Units)

	_, _, _ = procGlobalUnlock.Call(hGlobal)
	return hGlobal
}

// buildUnicodeTextMedium allocates a global memory block containing null-terminated UTF-16 text.
func buildUnicodeTextMedium(text string) uintptr {
	encoded, err := syscall.UTF16FromString(text)
	if err != nil || len(encoded) == 0 {
		return 0
	}
	totalBytes := uintptr(len(encoded) * 2)
	hGlobal, _, _ := procGlobalAlloc.Call(gmemMoveable|gmemZeroInit, totalBytes)
	if hGlobal == 0 {
		return 0
	}
	ptr, _, _ := procGlobalLock.Call(hGlobal)
	if ptr == 0 {
		_, _, _ = procGlobalFree.Call(hGlobal)
		return 0
	}
	//nolint:govet // Win32 HGLOBAL memory layout
	dstSlice := unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(encoded))
	copy(dstSlice, encoded)
	_, _, _ = procGlobalUnlock.Call(hGlobal)
	return hGlobal
}

// ShelfFileItem represents a local file or folder staged in the Vinta File Holding Shelf.
type ShelfFileItem struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	Extension    string `json:"extension"`
	IsDir        bool   `json:"isDir"`
	Category     string `json:"category"`
	PreviewData  string `json:"previewData,omitempty"`
	ModifiedTime int64  `json:"modifiedTime"`
	AddedAt      int64  `json:"addedAt"`
}

// classifyFileCategory maps a file extension or directory flag to a semantic shelf category.
func classifyFileCategory(ext string, isDir bool) string {
	if isDir {
		return "folder"
	}
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".ico", ".svg":
		return "image"
	case ".zip", ".rar", ".7z", ".tar", ".gz", ".tgz", ".bz2", ".xz":
		return "archive"
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv", ".rtf":
		return "document"
	case ".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac":
		return "audio"
	case ".mp4", ".mkv", ".webm", ".mov", ".avi":
		return "video"
	case ".go", ".ts", ".js", ".svelte", ".py", ".rs", ".c", ".cpp", ".h", ".json", ".yaml", ".yml", ".toml", ".html", ".css", ".sql", ".sh", ".bat", ".ps1":
		return "code"
	default:
		return "other"
	}
}

// buildImagePreviewData generates a compact base64 data URI thumbnail for supported image files.
func buildImagePreviewData(path string, ext string, size int64) string {
	const maxImageBytes = 15 * 1024 * 1024
	if size <= 0 || size > maxImageBytes {
		return ""
	}

	extLower := strings.ToLower(ext)
	// For small vector/icon/webp formats under 512KB, encode directly
	if (extLower == ".svg" || extLower == ".ico" || extLower == ".webp" || extLower == ".bmp") && size <= 512*1024 {
		raw, err := os.ReadFile(path) // #nosec G304 -- user-dropped local file path
		if err != nil || len(raw) == 0 {
			return ""
		}
		mime := "image/png"
		switch extLower {
		case ".svg":
			mime = "image/svg+xml"
		case ".ico":
			mime = "image/x-icon"
		case ".webp":
			mime = "image/webp"
		case ".bmp":
			mime = "image/bmp"
		}
		return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(raw))
	}

	if extLower != ".png" && extLower != ".jpg" && extLower != ".jpeg" && extLower != ".gif" {
		return ""
	}

	f, err := os.Open(path) // #nosec G304 -- user-dropped local file path
	if err != nil {
		return ""
	}
	defer func() {
		_ = f.Close()
	}()

	var src image.Image
	switch extLower {
	case ".png":
		src, err = png.Decode(f)
	case ".jpg", ".jpeg":
		src, err = jpeg.Decode(f)
	case ".gif":
		src, err = gif.Decode(f)
	}
	if err != nil || src == nil {
		return ""
	}

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return ""
	}

	const maxThumbDim = 96
	dstW := srcW
	dstH := srcH
	if srcW > maxThumbDim || srcH > maxThumbDim {
		if srcW >= srcH {
			dstW = maxThumbDim
			dstH = (srcH * maxThumbDim) / srcW
			if dstH < 1 {
				dstH = 1
			}
		} else {
			dstH = maxThumbDim
			dstW = (srcW * maxThumbDim) / srcH
			if dstW < 1 {
				dstW = 1
			}
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		sy := bounds.Min.Y + (y*srcH)/dstH
		for x := 0; x < dstW; x++ {
			sx := bounds.Min.X + (x*srcW)/dstW
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// extractFileNativeIcon retrieves the authentic Windows Shell 32x32 icon for the given path,
// rendering it onto a 32-bit DIB section and encoding as a transparent PNG base64 data URI.
func extractFileNativeIcon(path string, isDir bool) string {
	u16Path, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return ""
	}

	var sfi shFileInfoW
	flags := uint32(0x00000100) // SHGFI_ICON
	var fileAttr uint32
	if isDir {
		fileAttr = 0x00000010 // FILE_ATTRIBUTE_DIRECTORY
	}

	ret, _, _ := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(u16Path)),
		uintptr(fileAttr),
		uintptr(unsafe.Pointer(&sfi)),
		uintptr(unsafe.Sizeof(sfi)),
		uintptr(flags),
	)
	if ret == 0 || sfi.hIcon == 0 {
		ret, _, _ = procSHGetFileInfoW.Call(
			uintptr(unsafe.Pointer(u16Path)),
			uintptr(fileAttr),
			uintptr(unsafe.Pointer(&sfi)),
			uintptr(unsafe.Sizeof(sfi)),
			uintptr(flags|0x00000010), // SHGFI_USEFILEATTRIBUTES
		)
		if ret == 0 || sfi.hIcon == 0 {
			return ""
		}
	}
	defer func() { _, _, _ = procDestroyIcon.Call(sfi.hIcon) }()

	const iconSize = 32
	bmi := bitmapInfoHeader{
		biSize:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		biWidth:    iconSize,
		biHeight:   -iconSize, // top-down DIB
		biPlanes:   1,
		biBitCount: 32,
	}

	var bitsPtr uintptr
	hbmp, _, _ := procCreateDIBSection.Call(
		0,
		uintptr(unsafe.Pointer(&bmi)),
		0,
		uintptr(unsafe.Pointer(&bitsPtr)),
		0,
		0,
	)
	if hbmp == 0 || bitsPtr == 0 {
		return ""
	}
	defer func() { _, _, _ = procDeleteObject.Call(hbmp) }()

	hdc, _, _ := procCreateCompatibleDC.Call(0)
	if hdc == 0 {
		return ""
	}
	defer func() { _, _, _ = procDeleteDC.Call(hdc) }()

	oldBmp, _, _ := procSelectObject.Call(hdc, hbmp)
	defer func() { _, _, _ = procSelectObject.Call(hdc, oldBmp) }()

	res, _, _ := procDrawIconEx.Call(
		hdc,
		0,
		0,
		sfi.hIcon,
		iconSize,
		iconSize,
		0,
		0,
		3, // DI_NORMAL
	)
	if res == 0 {
		return ""
	}

	//nolint:govet // Win32 DIB memory buffer
	pixels := unsafe.Slice((*uint32)(unsafe.Pointer(bitsPtr)), iconSize*iconSize)
	img := image.NewRGBA(image.Rect(0, 0, iconSize, iconSize))

	hasAlpha := false
	for _, p := range pixels {
		if (p >> 24) != 0 {
			hasAlpha = true
			break
		}
	}

	for y := 0; y < iconSize; y++ {
		for x := 0; x < iconSize; x++ {
			p := pixels[y*iconSize+x]
			a := uint8(p >> 24)
			r := uint8(p >> 16)
			g := uint8(p >> 8)
			b := uint8(p)

			if hasAlpha {
				if a > 0 {
					unR := uint8((uint32(r) * 255) / uint32(a))
					unG := uint8((uint32(g) * 255) / uint32(a))
					unB := uint8((uint32(b) * 255) / uint32(a))
					img.SetRGBA(x, y, color.RGBA{R: unR, G: unG, B: unB, A: a})
				} else {
					img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
				}
			} else {
				if r != 0 || g != 0 || b != 0 {
					img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				} else {
					img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// InspectShelfFiles validates local file paths and returns their metadata and thumbnail previews.
func (a *App) InspectShelfFiles(paths []string) []ShelfFileItem {
	now := time.Now().UnixMilli()
	seen := make(map[string]bool, len(paths))
	items := make([]ShelfFileItem, 0, len(paths))

	for _, rawPath := range paths {
		trimmed := strings.TrimSpace(rawPath)
		if trimmed == "" {
			continue
		}
		cleanPath := filepath.Clean(trimmed)
		lowerKey := strings.ToLower(cleanPath)
		if seen[lowerKey] {
			continue
		}

		info, err := os.Stat(cleanPath)
		if err != nil {
			continue
		}
		seen[lowerKey] = true

		ext := ""
		if !info.IsDir() {
			ext = strings.ToLower(filepath.Ext(cleanPath))
		}
		category := classifyFileCategory(ext, info.IsDir())
		preview := ""
		if category == "image" && !info.IsDir() {
			preview = buildImagePreviewData(cleanPath, ext, info.Size())
		}
		if preview == "" {
			preview = extractFileNativeIcon(cleanPath, info.IsDir())
		}

		items = append(items, ShelfFileItem{
			Path:         cleanPath,
			Name:         filepath.Base(cleanPath),
			Size:         info.Size(),
			Extension:    ext,
			IsDir:        info.IsDir(),
			Category:     category,
			PreviewData:  preview,
			ModifiedTime: info.ModTime().UnixMilli(),
			AddedAt:      now,
		})
	}

	return items
}

// createShellDataObject builds a native Windows Shell IDataObject for the specified file paths
// using SHCreateShellItemArrayFromIDLists and BindToHandler(BHID_DataObject),
// falling back to the internal CF_HDROP COM data object if Shell initialization fails.
func createShellDataObject(paths []string) (uintptr, func(), error) {
	initCOMDragObjects()

	activeDragMu.Lock()
	activeDragPaths = paths
	activeDragMu.Unlock()

	var pidls []uintptr
	for _, p := range paths {
		u16, err := syscall.UTF16PtrFromString(p)
		if err != nil {
			continue
		}
		var pidl uintptr
		var sfgao uint32
		hr, _, _ := procSHParseDisplayName.Call(
			uintptr(unsafe.Pointer(u16)),
			0,
			uintptr(unsafe.Pointer(&pidl)),
			0,
			uintptr(unsafe.Pointer(&sfgao)),
		)
		if int32(hr) >= 0 && pidl != 0 {
			pidls = append(pidls, pidl)
		}
	}

	freePidls := func() {
		for _, pidl := range pidls {
			_, _, _ = procCoTaskMemFree.Call(pidl)
		}
	}

	if len(pidls) > 0 {
		// The root must be the desktop (an empty ITEMIDLIST), not NULL. With a NULL root,
		// SHCreateDataObject returns an object with no formats and every drop target rejects it.
		var desktopPidl uint16
		var shellDataObj uintptr
		hr, _, _ := procSHCreateDataObject.Call(
			uintptr(unsafe.Pointer(&desktopPidl)),
			uintptr(len(pidls)),
			uintptr(unsafe.Pointer(&pidls[0])),
			0,
			uintptr(unsafe.Pointer(&iidIDataObject)),
			uintptr(unsafe.Pointer(&shellDataObj)),
		)
		if int32(hr) >= 0 && shellDataObj != 0 {
			cleanup := func() {
				//nolint:govet // Win32 COM vtable invocation
				objVtbl := *(*uintptr)(unsafe.Pointer(shellDataObj))
				//nolint:govet // Win32 COM vtable invocation
				objRelease := *(*uintptr)(unsafe.Pointer(objVtbl + 2*unsafe.Sizeof(uintptr(0))))
				_, _, _ = syscall.SyscallN(objRelease, shellDataObj)
				freePidls()
			}
			return shellDataObj, cleanup, nil
		}
	}
	freePidls()

	dataObjInst := &comDataObject{
		lpVtbl: &dataObjectVtbl,
	}
	cleanup := func() {
		goruntime.KeepAlive(dataObjInst)
		enumPoolMu.Lock()
		activeEnums = nil
		enumPoolMu.Unlock()
	}
	return uintptr(unsafe.Pointer(dataObjInst)), cleanup, nil
}

// attachDragPreviewBitmap creates a 32-bit premultiplied ARGB drag preview card bitmap in memory
// with the file's authentic Windows icon and name, and attaches it via IDragSourceHelper::InitializeFromBitmap.
func attachDragPreviewBitmap(dataObj uintptr, paths []string) {
	if len(paths) == 0 || dataObj == 0 {
		return
	}

	const (
		cardW = 200
		cardH = 50
		rad   = 10
		pad   = 2
	)

	bmi := bitmapInfoHeader{
		biSize:     uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		biWidth:    cardW,
		biHeight:   -cardH, // top-down DIB
		biPlanes:   1,
		biBitCount: 32,
	}

	var bitsPtr uintptr
	hbmp, _, _ := procCreateDIBSection.Call(
		0,
		uintptr(unsafe.Pointer(&bmi)),
		0,
		uintptr(unsafe.Pointer(&bitsPtr)),
		0,
		0,
	)
	if hbmp == 0 || bitsPtr == 0 {
		return
	}

	//nolint:govet // Win32 DIB memory buffer
	pixels := unsafe.Slice((*uint32)(unsafe.Pointer(bitsPtr)), cardW*cardH)

	// Draw rounded rectangle card into premultiplied ARGB
	setPixelPremul := func(x, y int, r, g, b, a uint8) {
		if x < 0 || x >= cardW || y < 0 || y >= cardH {
			return
		}
		pr := (uint32(r) * uint32(a)) / 255
		pg := (uint32(g) * uint32(a)) / 255
		pb := (uint32(b) * uint32(a)) / 255
		pixels[y*cardW+x] = (uint32(a) << 24) | (pr << 16) | (pg << 8) | pb
	}

	for y := pad; y < cardH-pad; y++ {
		for x := pad; x < cardW-pad; x++ {
			dx, dy := 0, 0
			if x < pad+rad {
				dx = (pad + rad) - x
			} else if x >= cardW-pad-rad {
				dx = x - (cardW - pad - rad - 1)
			}
			if y < pad+rad {
				dy = (pad + rad) - y
			} else if y >= cardH-pad-rad {
				dy = y - (cardH - pad - rad - 1)
			}
			distSq := dx*dx + dy*dy
			if distSq > rad*rad {
				continue
			}
			isBorder := x <= pad+1 || x >= cardW-pad-2 || y <= pad+1 || y >= cardH-pad-2 || distSq >= (rad-2)*(rad-2)
			if isBorder {
				setPixelPremul(x, y, 16, 185, 129, 235) // emerald border
			} else {
				setPixelPremul(x, y, 24, 25, 30, 225) // dark frosted island slate
			}
		}
	}

	// Render the native Windows file icon and filename text using GDI
	hdc, _, _ := procCreateCompatibleDC.Call(0)
	if hdc != 0 {
		oldBmp, _, _ := procSelectObject.Call(hdc, hbmp)

		firstPath := paths[0]
		u16Path, err := syscall.UTF16PtrFromString(firstPath)
		if err == nil {
			var sfi shFileInfoW
			ret, _, _ := procSHGetFileInfoW.Call(
				uintptr(unsafe.Pointer(u16Path)),
				0,
				uintptr(unsafe.Pointer(&sfi)),
				uintptr(unsafe.Sizeof(sfi)),
				0x00000100, // SHGFI_ICON
			)
			if ret != 0 && sfi.hIcon != 0 {
				_, _, _ = procDrawIconEx.Call(hdc, 10, 9, sfi.hIcon, 32, 32, 0, 0, 3)
				_, _, _ = procDestroyIcon.Call(sfi.hIcon)
			}
		}

		fileName := filepath.Base(firstPath)
		u16Text, errText := syscall.UTF16PtrFromString(fileName)
		if errText == nil {
			_, _, _ = procSetBkMode.Call(hdc, 1) // TRANSPARENT
			_, _, _ = procSetTextColor.Call(hdc, 0x00FFFFFF)
			textRight := int32(cardW - 12)
			if len(paths) > 1 {
				textRight = int32(cardW - 38)
			}
			rc := rect32{Left: 48, Top: 8, Right: textRight, Bottom: 42}
			_, _, _ = procDrawTextW.Call(
				hdc,
				uintptr(unsafe.Pointer(u16Text)),
				^uintptr(0),
				uintptr(unsafe.Pointer(&rc)),
				0x00000000|0x00000004|0x00000020|0x00008000, // DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS
			)
		}

		_, _, _ = procSelectObject.Call(hdc, oldBmp)
		_, _, _ = procDeleteDC.Call(hdc)
	}

	// Post-process pixels inside the card bounds: ensure icon and text pixels have full premultiplied alpha
	for y := pad; y < cardH-pad; y++ {
		for x := pad; x < cardW-pad; x++ {
			p := pixels[y*cardW+x]
			a := uint8(p >> 24)
			r := uint8(p >> 16)
			g := uint8(p >> 8)
			b := uint8(p)
			if a == 0 && (r != 0 || g != 0 || b != 0) {
				// GDI blitted icon or text pixels with alpha 0; restore alpha to 245 and premultiply
				setPixelPremul(x, y, r, g, b, 245)
			}
		}
	}

	// If dragging multiple files, draw an emerald stack badge in the top-right corner
	if len(paths) > 1 {
		const bx, by, br = cardW - 20, 16, 11
		for y := by - br; y <= by+br; y++ {
			for x := bx - br; x <= bx+br; x++ {
				dx := x - bx
				dy := y - by
				if dx*dx+dy*dy <= br*br {
					setPixelPremul(x, y, 16, 185, 129, 255)
				}
			}
		}

		// Draw file count text into badge
		countStr := fmt.Sprintf("%d", len(paths))
		if u16Count, err := syscall.UTF16PtrFromString(countStr); err == nil {
			hdcBadge, _, _ := procCreateCompatibleDC.Call(0)
			if hdcBadge != 0 {
				oldBmp, _, _ := procSelectObject.Call(hdcBadge, hbmp)
				_, _, _ = procSetBkMode.Call(hdcBadge, 1)
				_, _, _ = procSetTextColor.Call(hdcBadge, 0x00000000) // black text inside emerald badge
				rcBadge := rect32{Left: bx - br, Top: by - br + 1, Right: bx + br, Bottom: by + br + 1}
				_, _, _ = procDrawTextW.Call(
					hdcBadge,
					uintptr(unsafe.Pointer(u16Count)),
					^uintptr(0),
					uintptr(unsafe.Pointer(&rcBadge)),
					0x00000001|0x00000004|0x00000020, // DT_CENTER | DT_VCENTER | DT_SINGLELINE
				)
				_, _, _ = procSelectObject.Call(hdcBadge, oldBmp)
				_, _, _ = procDeleteDC.Call(hdcBadge)
			}
		}
	}

	var dsh uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidDragDropHelper)),
		0,
		1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&iidIDragSourceHelper)),
		uintptr(unsafe.Pointer(&dsh)),
	)
	if int32(hr) < 0 || dsh == 0 {
		_, _, _ = procDeleteObject.Call(hbmp)
		return
	}
	defer func() {
		//nolint:govet // Win32 COM vtable invocation
		dshVtbl := *(*uintptr)(unsafe.Pointer(dsh))
		//nolint:govet // Win32 COM vtable invocation
		dshRelease := *(*uintptr)(unsafe.Pointer(dshVtbl + 2*unsafe.Sizeof(uintptr(0))))
		_, _, _ = syscall.SyscallN(dshRelease, dsh)
	}()

	shdi := shDragImage{
		cx:            int32(cardW),
		cy:            int32(cardH),
		ptOffsetX:     24,
		ptOffsetY:     24,
		hbmpDragImage: hbmp,
		crColorKey:    0xFFFFFFFF,
	}
	//nolint:govet // Win32 COM vtable invocation
	dshVtbl := *(*uintptr)(unsafe.Pointer(dsh))
	//nolint:govet // Win32 COM vtable invocation
	dshInitFromBitmap := *(*uintptr)(unsafe.Pointer(dshVtbl + 3*unsafe.Sizeof(uintptr(0))))
	hrInit, _, _ := syscall.SyscallN(
		dshInitFromBitmap,
		dsh,
		uintptr(unsafe.Pointer(&shdi)),
		dataObj,
	)
	if int32(hrInit) < 0 {
		_, _, _ = procDeleteObject.Call(hbmp)
	}
}

// PromptChooseShelfFiles opens a native multi-file selection dialog and returns the inspected shelf items.
func (a *App) PromptChooseShelfFiles() ([]ShelfFileItem, error) {
	app := application.Get()
	if app == nil || app.Dialog == nil {
		return nil, fmt.Errorf("dialog service unavailable")
	}
	paths, err := app.Dialog.OpenFile().
		SetTitle("Add Files to Vinta Shelf").
		CanChooseFiles(true).
		CanChooseDirectories(true).
		PromptForMultipleSelection()
	if err != nil || len(paths) == 0 {
		return nil, nil
	}
	return a.InspectShelfFiles(paths), nil
}

// CancelNativeFileDrag interrupts any in-progress native shell drag-and-drop session.
func (a *App) CancelNativeFileDrag() {
	dragCancelRequested.Store(true)
}

// StartNativeFileDrag initiates a native Windows Shell OLE drag-and-drop operation for the given file paths.
//
// The drag loop runs on the Wails main UI thread. That thread shares its input queue with the
// WebView2 child windows, so DoDragDrop can take over the mouse capture the webview holds from
// the initial click. Running it on any other thread makes Windows cancel the drag immediately.
func (a *App) StartNativeFileDrag(paths []string) error {
	validPaths := make([]string, 0, len(paths))
	for _, raw := range paths {
		clean := filepath.Clean(strings.TrimSpace(raw))
		if clean == "" {
			continue
		}
		if _, err := os.Stat(clean); err == nil {
			validPaths = append(validPaths, clean)
		}
	}
	if len(validPaths) == 0 {
		return fmt.Errorf("no valid files to drag")
	}

	dragSessionMu.Lock()
	if dragSessionActive {
		dragSessionMu.Unlock()
		return nil
	}
	dragSessionActive = true
	dragSessionMu.Unlock()

	initCOMDragObjects()
	dragCancelRequested.Store(false)

	application.InvokeAsync(func() {
		a.runNativeFileDrag(validPaths)
	})
	return nil
}

// runNativeFileDrag performs the modal SHDoDragDrop loop. It must be called on the main UI thread.
func (a *App) runNativeFileDrag(validPaths []string) {
	var hrDrop uintptr
	var effect uint32
	started := false

	hasLeftIsland.Store(false)
	isDragOverIsland.Store(false)

	a.pillHitMu.RLock()
	rects := make([]pillHitBounds, len(a.pillHitRects))
	copy(rects, a.pillHitRects)
	a.pillHitMu.RUnlock()

	activeIslandMu.Lock()
	activeIslandHwnd = a.getHwnd()
	activeIslandRects = rects
	activeIslandMu.Unlock()

	defer func() {
		hasLeftIsland.Store(false)
		if isDragOverIsland.Swap(false) {
			if app := application.Get(); app != nil && app.Event != nil {
				app.Event.Emit("fileshelf:drag-over-shelf", false)
			}
		}

		activeIslandMu.Lock()
		activeIslandHwnd = 0
		activeIslandRects = nil
		activeIslandMu.Unlock()

		removed := resolveDroppedPaths(validPaths, started, hrDrop, effect)
		if app := application.Get(); app != nil && app.Event != nil {
			app.Event.Emit("fileshelf:drag-ended", map[string]interface{}{
				"paths":   removed,
				"dropped": len(removed) > 0,
				"effect":  effect,
			})
		}
		dragSessionMu.Lock()
		dragSessionActive = false
		dragSessionMu.Unlock()
	}()

	// The UI thread is already an STA for WebView2; OleInitialize adds the OLE drag-drop layer.
	if hrInit, _, _ := procOleInitialize.Call(0); int32(hrInit) >= 0 {
		defer func() { _, _, _ = procOleUninitialize.Call() }()
	}

	dataObj, cleanupDataObj, err := createShellDataObject(validPaths)
	if err != nil || dataObj == 0 {
		return
	}
	defer cleanupDataObj()

	attachDragPreviewBitmap(dataObj, validPaths)

	a.setOutboundDragging(true)
	defer a.setOutboundDragging(false)

	if app := application.Get(); app != nil && app.Event != nil {
		app.Event.Emit("fileshelf:drag-started", validPaths)
	}

	started = true
	allowedEffects := uintptr(dropEffectCopy | dropEffectMove | dropEffectLink)

	hrDrop, _, _ = procSHDoDragDrop.Call(
		0,
		dataObj,
		uintptr(unsafe.Pointer(&dropSourceInst)),
		allowedEffects,
		uintptr(unsafe.Pointer(&effect)),
	)
}

// resolveDroppedPaths returns the shelf paths that actually changed location (moved or removed from original path).
func resolveDroppedPaths(paths []string, started bool, hrDrop uintptr, effect uint32) []string {
	if !started || hrDrop == dragdropSCancel {
		return []string{}
	}
	isDrop := hrDrop == dragdropSDrop || hrDrop == sOK
	if !isDrop {
		return []string{}
	}

	// If effect was an explicit move, or if the original file no longer exists at path, clear it
	if effect == dropEffectMove {
		return paths
	}
	moved := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			moved = append(moved, p)
		}
	}
	return moved
}

// CopyShelfPathsToClipboard copies the files to the system clipboard both as text and as native CF_HDROP file objects.
func (a *App) CopyShelfPathsToClipboard(paths []string) error {
	clean := make([]string, 0, len(paths))
	for _, p := range paths {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			clean = append(clean, filepath.Clean(trimmed))
		}
	}
	if len(clean) == 0 {
		return fmt.Errorf("no file paths to copy")
	}

	app := application.Get()
	if app != nil && app.Clipboard != nil {
		app.Clipboard.SetText(strings.Join(clean, "\r\n"))
	}

	// Also write CF_HDROP and CF_UNICODETEXT to the Win32 clipboard so Windows Explorer and text targets recognize them
	hGlobal := buildCFHDropMedium(clean)
	hGlobalText := buildUnicodeTextMedium(strings.Join(clean, "\r\n"))

	if res, _, _ := procOpenClipboard.Call(0); res != 0 {
		_, _, _ = procEmptyClipboard.Call()
		if hGlobal != 0 {
			if setRes, _, _ := procSetClipboardData.Call(cfHDrop, hGlobal); setRes == 0 {
				_, _, _ = procGlobalFree.Call(hGlobal)
			}
		}
		if hGlobalText != 0 {
			if setRes, _, _ := procSetClipboardData.Call(cfUnicodeText, hGlobalText); setRes == 0 {
				_, _, _ = procGlobalFree.Call(hGlobalText)
			}
		}
		_, _, _ = procCloseClipboard.Call()
	} else {
		if hGlobal != 0 {
			_, _, _ = procGlobalFree.Call(hGlobal)
		}
		if hGlobalText != 0 {
			_, _, _ = procGlobalFree.Call(hGlobalText)
		}
	}

	return nil
}

// RevealShelfFileInExplorer opens Windows Explorer with the specified file or folder selected.
func (a *App) RevealShelfFileInExplorer(path string) error {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" {
		return fmt.Errorf("empty path")
	}
	return exec.Command("explorer.exe", fmt.Sprintf("/select,%s", clean)).Start()
}

// BundleShelfZip compresses the specified files and directories into a timestamped .zip archive.
func (a *App) BundleShelfZip(paths []string) (ShelfFileItem, error) {
	validPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		clean := filepath.Clean(strings.TrimSpace(p))
		if clean == "" {
			continue
		}
		if _, err := os.Stat(clean); err == nil {
			validPaths = append(validPaths, clean)
		}
	}
	if len(validPaths) == 0 {
		return ShelfFileItem{}, fmt.Errorf("no valid files to bundle")
	}

	baseDir := filepath.Dir(validPaths[0])
	zipName := fmt.Sprintf("Vinta_Shelf_%s.zip", time.Now().Format("20060102_150405"))
	zipPath := filepath.Join(baseDir, zipName)

	outFile, err := os.Create(zipPath) // #nosec G304 -- user-initiated zip bundle in local folder
	if err != nil {
		return ShelfFileItem{}, err
	}

	zw := zip.NewWriter(outFile)
	writeErr := func() error {
		for _, itemPath := range validPaths {
			info, statErr := os.Stat(itemPath)
			if statErr != nil {
				continue
			}
			parentDir := filepath.Dir(itemPath)
			if info.IsDir() {
				walkErr := filepath.Walk(itemPath, func(p string, fi os.FileInfo, err error) error {
					if err != nil || fi.IsDir() {
						return err
					}
					rel, relErr := filepath.Rel(parentDir, p)
					if relErr != nil {
						rel = fi.Name()
					}
					return addFileToZip(zw, p, filepath.ToSlash(rel), fi)
				})
				if walkErr != nil {
					return walkErr
				}
			} else {
				if err := addFileToZip(zw, itemPath, filepath.Base(itemPath), info); err != nil {
					return err
				}
			}
		}
		return nil
	}()

	closeZipErr := zw.Close()
	closeFileErr := outFile.Close()
	if writeErr != nil {
		_ = os.Remove(zipPath)
		return ShelfFileItem{}, writeErr
	}
	if closeZipErr != nil {
		_ = os.Remove(zipPath)
		return ShelfFileItem{}, closeZipErr
	}
	if closeFileErr != nil {
		return ShelfFileItem{}, closeFileErr
	}

	inspected := a.InspectShelfFiles([]string{zipPath})
	if len(inspected) == 0 {
		return ShelfFileItem{}, fmt.Errorf("failed to inspect created archive")
	}
	return inspected[0], nil
}

func addFileToZip(zw *zip.Writer, srcPath string, zipRelPath string, info os.FileInfo) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = zipRelPath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(srcPath) // #nosec G304 -- user-staged local file
	if err != nil {
		return err
	}
	defer func() {
		_ = srcFile.Close()
	}()
	_, err = io.Copy(writer, srcFile)
	return err
}

// StripShelfImageMetadata re-encodes supported images (.jpg, .jpeg, .png) without EXIF or ancillary metadata.
func (a *App) StripShelfImageMetadata(paths []string) ([]ShelfFileItem, error) {
	cleanedPaths := make([]string, 0, len(paths))

	for _, raw := range paths {
		cleanPath := filepath.Clean(strings.TrimSpace(raw))
		if cleanPath == "" {
			continue
		}
		ext := strings.ToLower(filepath.Ext(cleanPath))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		srcFile, err := os.Open(cleanPath) // #nosec G304 -- user-staged local image
		if err != nil {
			continue
		}
		var img image.Image
		if ext == ".png" {
			img, err = png.Decode(srcFile)
		} else {
			img, err = jpeg.Decode(srcFile)
		}
		_ = srcFile.Close()
		if err != nil || img == nil {
			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(cleanPath), filepath.Ext(cleanPath))
		outPath := filepath.Join(filepath.Dir(cleanPath), fmt.Sprintf("%s_clean%s", baseName, ext))

		outFile, err := os.Create(outPath) // #nosec G304 -- user-requested clean copy
		if err != nil {
			continue
		}
		if ext == ".png" {
			err = png.Encode(outFile, img)
		} else {
			err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 95})
		}
		_ = outFile.Close()
		if err != nil {
			_ = os.Remove(outPath)
			continue
		}
		cleanedPaths = append(cleanedPaths, outPath)
	}

	if len(cleanedPaths) == 0 {
		return nil, fmt.Errorf("no supported JPEG or PNG images found to strip")
	}
	return a.InspectShelfFiles(cleanedPaths), nil
}
